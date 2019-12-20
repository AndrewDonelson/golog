package main

import (
	"github.com/AndrewDonelson/golog"
)

func doLogs() {
	method := "doLogs"
	golog.Log.Trace(method, "example.go", 7)
	golog.Log.SetFunction(method)

	// Critically log critical
	golog.Log.Critical("This is Critical message!")
	// Show the error
	golog.Log.Error("This is Error message!")
	// Show the success
	golog.Log.Success("This is Success message!")
	// Give the Warning
	golog.Log.Warning("This is Warning message!")
	// Notice
	golog.Log.Notice("This is Notice message!")
	// Show the info
	golog.Log.Info("This is Info message, Fatal & Panic skipped!")
	// Debug
	golog.Log.Debug("This is Debug message!")
}

func main() {
	// Get the instance for logger class
	// Third option is optional and is instance of type io.Writer, defaults to os.Stderr
	println("\nProduction Output: as Log")
	golog.Log.SetModuleName("prod-example")
	golog.Log.SetEnvironment(golog.EnvProduction)
	doLogs()

	println("\nProduction Output: as JSON")
	golog.Log.UseJSONForProduction()
	doLogs()

	println("\nTest/QA Output:")
	golog.Log.SetModuleName("qa-example")
	golog.Log.SetEnvironment(golog.EnvQuality)
	doLogs()

	println("\nDevelopment Output:")
	golog.Log.SetModuleName("dev-example")
	golog.Log.SetEnvironment(golog.EnvDevelopment)
	doLogs()
}
