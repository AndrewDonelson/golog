package main

import (
	"github.com/AndrewDonelson/golog"
)

func doLogs(log *golog.Logger) {
	method := "doLogs"
	log.Trace(method, "example.go", 7)
	log.SetFunction(method)

	// Critically log critical
	log.Critical("This is Critical message!")
	// Show the error
	log.Error("This is Error message!")
	// Show the success
	log.Success("This is Success message!")
	// Give the Warning
	log.Warning("This is Warning message!")
	// Notice
	log.Notice("This is Notice message!")
	// Show the info
	log.Info("This is Info message, Fatal & Panic skipped!")
	// Debug
	log.Debug("This is Debug message!")
}

func main() {
	// Get the instance for logger class
	// Third option is optional and is instance of type io.Writer, defaults to os.Stderr
	println("\nProduction Output: as Log")
	log, err := golog.NewLogger(&golog.Options{Module: "prod-example"})
	if err != nil {
		panic(err) // Check for error
	}
	log.SetEnvironment(golog.EnvProduction)
	doLogs(log)

	println("\nProduction Output: as JSON")
	log.UseJSONForProduction()
	doLogs(log)

	println("\nTest/QA Output:")
	log, err = golog.NewLogger(&golog.Options{Module: "qa-example"})
	if err != nil {
		panic(err) // Check for error
	}
	log.SetEnvironment(golog.EnvQuality)
	doLogs(log)

	println("\nDevelopment Output:")
	log, err = golog.NewLogger(&golog.Options{Module: "dev-example"})
	if err != nil {
		panic(err) // Check for error
	}
	log.SetEnvironment(golog.EnvDevelopment)
	doLogs(log)
}
