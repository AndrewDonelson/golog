package golog

import (
	"net/http"
	"testing"
	"time"
)

func TestInit(t *testing.T) {

	SetGoLogBuildEnv("testing", "1.23.456", "")
	Assert(t, (Log.Options.Module == "testing"), "Module name is not correct")
}

func TestEnvironment(t *testing.T) {

	SetGoLogBuildEnv("testing", "1.23.456", "")

	Log.SetEnvironment(EnvDevelopment)
	Assert(t, (Log.Options.Environment == EnvDevelopment), "Environment should be Development")

	Log.SetEnvironment(EnvQuality)
	Assert(t, (Log.Options.Environment == EnvQuality), "Environment should be Quality")

	Log.SetEnvironment(EnvProduction)
	Assert(t, (Log.Options.Environment == EnvProduction), "Environment should be Production")

}

func TestProductionLock(t *testing.T) {
	var r *http.Request
	r.Host = "https://localhost:8080"
}

func TestMonitor(t *testing.T) {
	NewMonitor(60, "debug")
	SetGoLogBuildEnv("testing", "1.23.456", "")
	Log.SetEnvironment(EnvDevelopment)

	r, _ := http.NewRequest("GET", "test.localhost:8080", nil)
	start := GologMonitor.IncEndpoint("clickURL", r)

	delay, _ := time.ParseDuration("550ms")
	time.Sleep(delay)

	GologMonitor.TimeEndpoint("clickURL", int64(time.Since(start).Milliseconds()))

	delay, _ = time.ParseDuration("1550ms")
	time.Sleep(delay)

	GologMonitor.TimeEndpoint("clickURL", int64(time.Since(start).Milliseconds()))
}
