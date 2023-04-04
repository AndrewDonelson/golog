package golog

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"
)

/*
- This is used for API Mock Unit Testing

  - See: api/oauth_test for example
    *

- USAGE:

	func TestDoStuffWithRoundTripper(t *testing.T) {

	if util.InitializePreprodStorageSources(false) {

	client := util.NewTestClient(func(req *http.Request) *http.Response {
	// Test request parameters
	return &http.Response{
	StatusCode: 200,
	// Send response to be tested
	Body: ioutil.NopCloser(bytes.NewBufferString(`OK`)),
	// Must be set to non-nil value or it panics
	Header: make(http.Header),
	}
	})

	// Test Users
	// 85111:ceb1e2bc-fb26-462e-91b2-875dc22299c5
	// 85110:0123456789
	// 85109:0123456789
	testEndpoints := []string{
	"/v0/user",
	"/v0/merchants?limit=6&schema=v4",
	}

	rounds := 10

	api := util.API{Client: client, BaseURL: "https://andrew-dev-v0.incentivenetworks.com"}
	for i := 0; i < rounds; i++ {
	fmt.Printf("\nRound %d:\n", i)
	for _, ep := range testEndpoints {
	//body, err := api.ConsumeV0Endpoint(ep, GetUserOAuthToken(85111, "ceb1e2bc-fb26-462e-91b2-875dc22299c5"))
	body, err := api.ConsumeV0OAuthEndpoint(ep, GetUserOAuthToken(85109, "0123456789", ""))
	util.Ok(t, err)
	if strings.Contains(string(body), "error_message") {
	t.Fail()
	}
	}
	}
	}
	}
*/
type API struct {
	Client  *http.Client
	BaseURL string
}

// ConsumeV0AuthWL tests an endpoint using AuthWL Users
func (api *API) ConsumeV0AuthWLEndpoint(endpoint, sessionID string) ([]byte, error) {
	url := api.BaseURL + endpoint
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Basic YXNoaXNoOmEyNkIpbl5BPyU2V2YhQXNKdw==")
	req.Header.Add("Cookie", sessionID)

	start := time.Now()
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	elapsed := time.Since(start)

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code [%d] returned instead of 200", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%s: Status [%d] Time: [%s]\n", endpoint, res.StatusCode, elapsed)
	return body, nil
}

// ConsumeV0OAuthEndpoint tests an endpoint using OAuth Users
func (api *API) ConsumeV0OAuthEndpoint(endpoint, oAuthUser string) ([]byte, error) {
	url := api.BaseURL + endpoint
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Basic YXNoaXNoOmEyNkIpbl5BPyU2V2YhQXNKdw==")
	req.Header.Add("Oauth", oAuthUser)

	start := time.Now()
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	elapsed := time.Since(start)

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code [%d] returned instead of 200", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%s: Status [%d] Time: [%s]\n", endpoint, res.StatusCode, elapsed)
	return body, nil
}

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func IsUnitTesting() bool {
	return flag.Lookup("test.v") != nil
}

// assert fails the test if the condition is false.
func Assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: "+msg+"\033[39m\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// ok fails the test if an err is not nil.
func Ok(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// NotOk fails the test if an err is nil.
func NotOk(tb testing.TB, err error) {
	if err == nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d: unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// equals fails the test if exp is not equal to act.
func Equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

func InitializePreprodStorageSources(production bool) bool {

	if DbRO != nil && DbRW != nil {
		return true
	}
	// dbSettings = nil

	Log.SetModuleName("UnitTest")

	// if dbSettings == nil {
	// 	if production {
	// 		Log.Print("***** CAUTION PRODUCTION DATABASE IN TESTING ENVIRONMENT *****")
	// 		//dbSettings = &DbSettings{}
	// 	} else {
	// 		//dbSettings = &DbSettings{}
	// 	}
	// }

	// InitDb(dbSettings.CredentialsFile)

	// Set DBRO connection resource string
	// err = ConnectDbRO()
	// if err != nil {
	// 	panic(err)
	// }

	// Set DBRW connection resource string
	// err = ConnectDbRW()
	// if err != nil {
	// 	panic(err)
	// }

	// Connect Redis (DEV)
	// RedisHost:        "v0-stage-redis.fwyhds.ng.0001.use1.cache.amazonaws.com:6379"
	// RedisPort:        "6379"
	// err = RedisConnect("localhost:6379")
	// if err != nil {
	// 	Log.Error(err)
	// 	os.Exit(1)
	// }
	// defer RedisPool.Empty()

	// conn, err := getConnection()
	// if conn == nil {
	// 	Log.Errorf("unable to get Redis connection: %v", err)
	// }

	//Log.Printf("RedisPool Avail: [%d]", RedisPool.Avail())
	// return (DbRO != nil && DbRW != nil && RedisPool != nil)

	return (DbRO != nil && DbRW != nil)
}
