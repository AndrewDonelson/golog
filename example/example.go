package main

import (
	"github.com/AndrewDonelson/golog"
)

func doLogs(log *golog.Logger) {
	// Critically log critical
	log.Critical("This is Critical!")
	// Show the error
	log.Error("This is Error!")
	// Show the success
	log.Success("This is Success!")
	// Give the Warning
	log.Warning("This is Warning!")
	// Notice
	log.Notice("This is Notice!")
	// Show the info
	log.Info("This is Info!")
	// Debug
	log.Debug("This is Debug!")
}

func main() {
	// Get the instance for logger class
	// Third option is optional and is instance of type io.Writer, defaults to os.Stderr
	println("\nProduction Output:")
	log, err := golog.NewLogger(nil)
	if err != nil {
		panic(err) // Check for error
	}
	log.SetEnvironment(0)
	doLogs(log)

	println("\nTest/QA Output:")
	log, err = golog.NewLogger(nil)
	if err != nil {
		panic(err) // Check for error
	}
	log.SetEnvironment(1)
	doLogs(log)

	println("\nDevelopment Output:")
	log, err = golog.NewLogger(nil)
	if err != nil {
		panic(err) // Check for error
	}
	log.SetEnvironment(0)
	doLogs(log)
}
