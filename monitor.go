package golog

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/hako/durafmt"
	"github.com/jmoiron/sqlx"
	"github.com/julienschmidt/httprouter"
)

// GoLogInitialEnv used to hold the Default (or first time set) Environment
var (
	GoLogInitialEnv = EnvProduction
	DbRO            *sqlx.DB // Temporary Place holder - will be removed
	DbRW            *sqlx.DB // Temporary Place holder - will be removed
)

// GologError Error Golog Response Object
type GologError struct {
	Message string `json:"message"` // [Generated] message returned in response
}

// GologResponse Successful Golog Response Object
type GologResponse struct {
	Command     string     `json:"command"`     // [Request] The requested command
	Value       string     `json:"value"`       // [Request] A value value for the given command
	Message     string     `json:"message"`     // [Generated] message returned in response
	Module      string     `json:"module"`      // [Options] Name of running module
	Environment string     `json:"environment"` // [Options] Override default handling (Environment)
	Monitor     MonitorAPI `json:"monitor"`     // [Monitor] Displays API Info & Statistics
}

// TimeTrack is used to measure how long a function to execute
// Usage:
//
//	func factorial(n *big.Int) (result *big.Int) {
//	    defer timeTrack(time.Now())
//	    // ... do some things, maybe even return under some condition
//	    return n
//	}
func TimeTrack(start time.Time) {
	elapsed := time.Since(start)
	Log.Printf("took %s", elapsed)
}

// HostIsproduction is used to check if the API is running on a production instance
// and is called to make sure if the production Server IS NOT in EnvProduction mode
// after an alloted time, it will force it back into Production Logging Mode
func HostIsProduction(r *http.Request) bool {
	// This will return true on
	// - https://apiv0-prod.incentivenetworks.com
	// - https://apiv0-prod-blue.incentivenetworks.com
	// - https://apiv0-prod-green.incentivenetworks.com
	// - https://apiv0-prodtest.incentivenetworks.com
	// - https://apiv0-prod*.incentivenetworks.com
	Log.Debugf("Request Host: %s", r.Host)
	Log.Debugf("Request URL Host: %s", r.URL.Host)
	//return strings.Contains(r.Host, "apiv0-prod") // Add r.URL.Host

	// Suggest changing https://apiv0-testprod.incentivenetworks.com to https://apiv0-prodtest.incentivenetworks.com
	// so that the logic will detect [apiv0-prod] which will it be tested on Production Test Server as well before going
	// to production

	// check for apiv0-prod which will be true for apiv0-prod, apiv0-prod-green, apiv0-prod-blue and apiv0-prod-prodtest
	isProdLevel := strings.Contains(r.Host, "apiv0-prod") || strings.Contains(r.URL.Host, "apiv0-prod")

	// check for apiv0-prodtest which will be true for apiv0-prodtest - we want prodtest to return false for production
	isProdTest := strings.Contains(r.Host, "apiv0-prodtest") || strings.Contains(r.URL.Host, "apiv0-prodtest")

	//return strings.Contains(r.Host, "apiv0-prod") || strings.Contains(r.URL.Host, "apiv0-prod")
	return (isProdLevel && !isProdTest)
}

func MapOptions(r GologResponse, opts Options) (resp GologResponse) {
	resp = r
	resp.Environment = opts.EnvAsString()
	resp.Module = opts.Module
	resp.Monitor = GologMonitor

	return
}

// getEnvironmentFromString is used to get the log environment to either development, testing or production
func getEnvironmentFromString(env string) Environment {
	switch env {
	case "dev":
		return EnvDevelopment
	case "qa":
		return EnvQuality
	case "prod":
		return EnvProduction
	default:
		return EnvAuto
	}
}

// EnvAsString returns the current envirnment for options as a string
func envAsString(env Environment) string {
	environments := [...]string{
		"EvnAuto",
		"EnvDevelopment",
		"EnvQuality",
		"EnvProduction",
	}
	return environments[env]
}

// GologRouter given a valid router will add golog routes to the router
func GologRouter(r *httprouter.Router) {
	r.Handle("GET", "/golog", GologHandler())
}

// GologHandler request handler used for realtime Golog information and control
// To use just add to your router, example: router.Handle("GET", "/golog", util.GologHandler())
// and change the 32 character password in the handlers constants.
func GologHandler() func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		const (
			AuthPassword = "LAQkSaqqLC^M8^_k$M?y33hD8=6TT*pb"
			strSuccess   = "Successful Authorization"
			strInvalid   = "Invalid Authorization"
		)

		var (
			err            error
			qpCmd, qpValue string
			js             []byte
			resp           GologResponse
			admin          bool
		)

		Log.HandlerLog(w, r)

		resp = GologResponse{}

		// Validate AuthPassword
		auth := r.Header.Get("auth")
		if auth != AuthPassword {
			Log.Warning(strInvalid)
			resp.Message = strInvalid

			// Marshal and Return the response object
			msg := GologError{Message: strInvalid}
			js, err = json.Marshal(msg)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			Log.Success(strSuccess)
			resp.Message = strSuccess
			admin = true
		}

		if admin {
			// Get current LIVE Datbase Statistics on ever request
			GologMonitor.Info.GetDatabaseStats()

			// Handle the cmd Query Parameter (optional)
			qpCmd, _ = GetQueryParameter(r, "cmd", false, true, "get")
			resp.Command = strings.ToLower(qpCmd)

			// Handle the value Query Parameter (optional)
			qpValue, _ = GetQueryParameter(r, "value", false, true, "status")
			resp.Value = strings.ToLower(qpValue)

			if resp.Command == "get" {
				if resp.Value == "status" {
					resp.Message = "success - current status"
				}
			} else if resp.Command == "env" {
				// Set / Change Environment - DO NOT allow in Production (default, blue & green)
				if HostIsProduction(r) {
					Log.Print(PrettyPrint(r))
					resp.Message = fmt.Sprintf("Host [%s] Denied - Not allowed to change production logging level to [%s]", r.Host, Log.Options.EnvAsString())
					// Production is defaulting to Development. This is a quick fix to force back into prod mode
					Log.SetEnvironmentFromString("prod")
				} else {
					if resp.Value == "dev" || resp.Value == "qa" || resp.Value == "prod" {

						oldEnv := Log.Options.Environment

						// Only change environment is different and current
						if oldEnv != getEnvironmentFromString(resp.Value) {
							Log.SetEnvironmentFromString(resp.Value)

							resp.Message = "success - environment changed from [" + envAsString(oldEnv) + "] to " + Log.Options.EnvAsString() + "]"

						} else {
							resp.Message = "Ignoring, Already in [" + envAsString(oldEnv) + "] environment"
						}

					} else {
						resp.Message = "Valid options for [env] are (dev, qa & prod)"
					}
				}
			} else {
				resp.Message = "Error - Invalid Command [" + resp.Command + "]"
			}

			resp = MapOptions(resp, Log.Options)

			// Marshal and Return the response object
			js, err = json.Marshal(resp)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	}
}

// SetGoLogBuildEnv provides consistent log level and startup tag across all services and servers
func SetGoLogBuildEnv(serverName, version, dbFile string) {
	Log.SetModuleName(serverName)

	// Lets put all Servers & Services into Development mode for DEV BUILD_ENV
	dbFile = strings.ToLower(dbFile)
	if strings.Contains(dbFile, "dev") {
		Log.SetEnvironment(EnvDevelopment)
	} else {
		Log.SetEnvironment(EnvProduction)
	}

	Log.Printf("%s Server [Version %s] (%s)\n", serverName, version, Log.Options.EnvAsString())
	go NewMonitor(60, version)
}

type EPMetrics struct {
	Min int32 `json:"min"`
	Max int32 `json:"max"`
	Avg int32 `json:"avg"`
}

func (m *EPMetrics) Hit(current int32) {
	// Get the Max per interval
	if current > m.Max {
		m.Max = current
	}

	// Get the Min per interval
	if current < m.Min || m.Min == 0 {
		m.Min = current
	}

	// Get the current average
	// m.Avg =  m.Avg + int32((m.Min + m.Max) / 2)
	m.Avg = int32((m.Avg + current + m.Min + m.Max) / 4)
}

type EPStats struct {
	// YTD is the total number of times endpoint has been called since start
	YTD int32 `json:"ytd"`
	// Count is the number of times endpoints has been called since last check
	ReqCurrent   int32     `json:"current"`
	RequestHits  EPMetrics `json:"hits"`
	RequestTimes EPMetrics `json:"times"`
}

func (e *EPStats) Hit() {
	e.YTD++
	e.ReqCurrent++
	e.RequestHits.Hit(e.ReqCurrent)
}

func (e *EPStats) Time(elapsed int64) {
	e.RequestTimes.Hit(int32(elapsed))
}

func (e *EPStats) Reset() {
	e.ReqCurrent = 0
}

type V0EndpointMonitor map[string]EPStats

type Endpoints struct {
	Totals   EPStats // API Combined Statistics
	GetOAuth EPStats // Called from OAuth-Manager (Special Case)
	GetUser  EPStats // makeHandleV0User(cs)
	PostUser EPStats // makeHandleV0UserUpdate(cs)
}

func (e *Endpoints) Reset() {
	e.Totals.Reset()
	e.GetOAuth.Reset()
	e.GetUser.Reset()
	e.PostUser.Reset()
}

type APIRefreshInfo struct {
	LastRefresh     string `json:"last_refresh"`
	NextRefresh     string `json:"next_refresh"`
	YTDRefreshes    int64  `json:"total_refreshes_completed"`
	RefreshDuration string `json:"refresh_duration"`
}

func (r *APIRefreshInfo) updateNextRefresh() error {
	t, err := time.Parse(DefaultDateFormat, r.LastRefresh)
	if err != nil {
		return err
	}

	refreshInterval := 20 * time.Minute
	r.NextRefresh = fmt.Sprint(time.Until(t.Add(refreshInterval)))
	return nil
}

func (r *APIRefreshInfo) Update(elapsed time.Duration) {

	nexttime := time.Now().Add(time.Minute * time.Duration(20))
	r.YTDRefreshes++
	r.RefreshDuration = durafmt.Parse(elapsed).LimitFirstN(2).String() // // limit first two parts.
	r.LastRefresh = time.Now().Format(DefaultDateFormat)
	r.NextRefresh = durafmt.Parse(time.Until(nexttime)).LimitFirstN(2).String()
}

// DBInfo Stores information about each of the database connections (DBRO & DBRW) of the deployed API.
type DBInfo struct {
	// Which database? this is the host connected to
	Database string `json:"database"`
	Stats    sql.DBStats
}

func (i *DBInfo) DBInfo(db *sqlx.DB) {
	if db != nil {
		i.Stats = db.Stats()
	}
}

type APIDatabase struct {
	DBRO DBInfo `json:"dbro"`
	DBRW DBInfo `json:"dbrw"`
}

func (i *APIDatabase) Update() {
	i.DBRO.DBInfo(DbRO)
	i.DBRW.DBInfo(DbRW)
}

// APIInfo Stores information about the deployed API.
type APIInfo struct {
	Version  string         `json:"version"`
	Built    string         `json:"build_time"`
	Deployed string         `json:"start_time"`
	UpTime   string         `json:"up_time"`
	Host     string         `json:"host"`
	Database APIDatabase    `json:"database"`
	Refresh  APIRefreshInfo `json:"refresh_info"`
}

func (i *APIInfo) Get() *APIInfo {
	return i
}

func (i *APIInfo) uptime() string {
	t, err := time.Parse(DefaultDateFormat, i.Deployed)

	if err != nil {
		Log.Warning(err)
		return ""
	}
	return durafmt.Parse(time.Since(t)).LimitFirstN(2).String() // // limit first two parts.
}

func (i *APIInfo) SetHost(host string) {
	i.Host = host
}

func (i *APIInfo) GetDatabaseStats() {
	i.Database.Update()
}

func (i *APIInfo) SetDatabaseDBRO(host string) {
	i.Database.DBRO.Database = host
}

func (i *APIInfo) SetDatabaseDBRW(host string) {
	i.Database.DBRW.Database = host
}

func (i *APIInfo) Update() {
	//	i.GetDatabaseStats()
	i.UpTime = i.uptime()
}

// APIResources stores information about the resources used by the API.
type APIResources struct {
	Sys          string `json:"system_memory"`      // Sys is the total bytes of memory obtained from the OS. Sys is the sum of the XSys fields below. Sys measures the virtual address space reserved by the Go runtime for the heap, stacks, and other internal data structures. It's likely that not all of the virtual address space is backed by physical memory at any given moment, though in general it all was at some point.
	Mallocs      string `json:"memory_allocations"` // Mallocs is the cumulative count of heap objects allocated. The number of live objects is Mallocs - Frees.
	Frees        string `json:"memory_frees"`
	Alloc        string `json:"current_memory"`
	Idle         string `json:"memory_idle"`
	NumGC        uint32 `json:"total_garbage_collections"`
	NumGoroutine int    `json:"current_goroutines"`
}

func (r *APIResources) bigNumberToString(n uint64) string {
	var result string

	if n >= 1000000000000 {
		result = fmt.Sprintf("%d TB", r.ToTera(n))
	} else if n >= 1000000000 {
		result = fmt.Sprintf("%d GB", r.ToGiga(n))
	} else if n >= 1000000 {
		result = fmt.Sprintf("%d MB", r.ToMega(n))
	} else if n >= 1000 {
		result = fmt.Sprintf("%d KB", r.ToKilo(n))
	} else {
		result = fmt.Sprintf("%d bytes", n)
	}

	return result

}

func (r *APIResources) ToKilo(b uint64) uint64 {
	return b / 1024
}

func (r *APIResources) ToMega(b uint64) uint64 {
	return b / 1024 / 1024
}

func (r *APIResources) ToGiga(b uint64) uint64 {
	return b / 1024 / 1024 / 1024
}

func (r *APIResources) ToTera(b uint64) uint64 {
	return b / 1024 / 1024 / 1024 / 1024
}

func (r *APIResources) Set(rtm *runtime.MemStats) {
	// Misc memory stats
	r.Sys = r.bigNumberToString(rtm.Sys)
	r.Alloc = r.bigNumberToString(rtm.Alloc)
	r.Mallocs = r.bigNumberToString(rtm.Mallocs)
	r.Frees = r.bigNumberToString(rtm.Frees)
	r.Idle = r.bigNumberToString(rtm.HeapIdle - rtm.HeapReleased)
	// GC Stats
	r.NumGC = rtm.NumGC
	// Number of goroutines
	r.NumGoroutine = runtime.NumGoroutine()

}

func (r *APIResources) Get() *APIResources {

	return r
}

// MonitorAPI is the response for golog status of API.
type MonitorAPI struct {
	Info      APIInfo      `json:"information"`
	Resources APIResources `json:"resources"`
	Endpoints Endpoints    `json:"endpoints"`
}

func (m *MonitorAPI) TimeEndpoint(method string, elapsed int64) {
	m.Endpoints.Totals.Time(elapsed)
	switch method {
	case "OAuth":
		m.Endpoints.GetOAuth.Time(elapsed)
	case "getUser":
		m.Endpoints.GetUser.Time(elapsed)
	case "updateUser":
		m.Endpoints.PostUser.Time(elapsed)
	}

	if elapsed < 2000 {
		Log.Printf("handler response elapsed %d ms", elapsed)
	} else {
		Log.Printf("handler response elapsed %d ms", elapsed)
		//Log.Warningf("handler response elapsed [%v] exceeded 2sec", elapsed)
	}
}

// IncEndpoint called by handlers to track endpoint comsumption. Will ignore
// any endpoints with header variable [internal] set to [alert||debug]
func (m *MonitorAPI) IncEndpoint(method string, r *http.Request) time.Time {
	start := time.Now()

	// Make sure this is not a ALERT or DEBUG request
	if len(r.Header.Get("internal")) > 0 {
		return start
	}

	m.Endpoints.Totals.Hit()

	switch method {
	case "OAuth":
		m.Endpoints.GetOAuth.Hit()
	case "getUser":
		m.Endpoints.GetUser.Hit()
	case "updateUser":
		m.Endpoints.PostUser.Hit()
	}

	return start
}

func (m *MonitorAPI) UpdateRefresh(elapsed time.Duration) {
	m.Info.Refresh.Update(elapsed)
}

func (m *MonitorAPI) UpdateResources(rtm *runtime.MemStats) {

	m.Resources.Set(rtm)
	m.Info.Update()
	m.Info.Refresh.updateNextRefresh()
}

func (m *MonitorAPI) Reset() {
	m.Endpoints.Reset()
}

var GologMonitor MonitorAPI

func NewMonitor(duration int, version string) {
	var rtm runtime.MemStats
	var interval = time.Duration(duration) * time.Second
	currentTime := time.Now()

	// Initialize
	GologMonitor = MonitorAPI{}

	fields := strings.Fields(version)
	GologMonitor.Info.Version = fields[0]
	for idx, val := range fields {
		val = strings.ToLower(val)
		if val == "build" || val == "built" {
			GologMonitor.Info.Built = fields[idx+1]
			fmt.Printf("Set Build = %s\n", fields[idx+1])
		}
	}

	GologMonitor.Info.Deployed = currentTime.Format(DefaultDateFormat)

	runtime.ReadMemStats(&rtm)
	GologMonitor.UpdateResources(&rtm)

	for {
		<-time.After(interval)
		GologMonitor.Reset()

		// Read full mem stats
		runtime.ReadMemStats(&rtm)
		GologMonitor.UpdateResources(&rtm)
	}
}
