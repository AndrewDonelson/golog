package main

import (
	"github.com/AndrewDonelson/golog"
)

func doLogs(log *golog.Logger) {
	// Critically log critical
	log.Critical("This is Critical!")
	// Debug
	log.Debug("This is Debug!")
	// Give the Warning
	log.Warning("This is Warning!")
	// Show the error
	log.Error("This is Error!")
	// Notice
	log.Notice("This is Notice!")
	// Show the info
	log.Info("This is Info!")
	// Show the info
	log.Success("This is Success!")
}

func main() {
	// Get the instance for logger class
	// Third option is optional and is instance of type io.Writer, defaults to os.Stderr
	println("\nProduction Output:")
	log, err := golog.NewLogger("production", 0)
	if err != nil {
		panic(err) // Check for error
	}
	doLogs(log)

	println("\nTest/QA Output:")
	log, err = golog.NewLogger("test-qa", 1)
	if err != nil {
		panic(err) // Check for error
	}
	doLogs(log)

	println("\nDevelopment Output:")
	log, err = golog.NewLogger("development", 2)
	if err != nil {
		panic(err) // Check for error
	}
	doLogs(log)
}
