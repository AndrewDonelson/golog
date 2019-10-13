package main

import (
	"os"

	"github.com/AndrewDonelson/golog"
)

func main() {
	// Get the instance for logger class
	// Third option is optional and is instance of type io.Writer, defaults to os.Stderr
	log, err := golog.New("test", 1, os.Stdout)
	if err != nil {
		panic(err) // Check for error
	}

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

	// Show warning with format message
	log.SetFormat("[%{module}] [%{level}] %{message}")
	log.Warning("This is Warning!")
	// Also you can set your format as default format for all new loggers
	golog.SetDefaultFormat("%{message}")
	log2, _ := golog.New("pkg", 1, os.Stdout)
	log2.Error("This is Error!")
}
