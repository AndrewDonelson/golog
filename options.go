// Package golog Simple flexible go logging
// This file contains logger configuration implementation
package golog

import (
	"io"
	"os"
)

// TODO:
// - Add support for loading from file JSON || YAML
// - Add support for loading from Environment Variables

// Environment enumeration
type Environment int

const (
	// EnvAuto - No Environment set (initial) Will detect by looking for BUILD_ENV os variable
	EnvAuto Environment = 0 + iota
	// EnvTesting - Internal, Used with `go test`, `goveralls`, ect
	EnvTesting
	// EnvDevelopment - All Log levels, color enabled and extra info on errors
	EnvDevelopment
	// EnvQuality - No debug level logging, color enabled, no extra info on errors
	EnvQuality
	// EnvProduction - Error level & higher, no color, minimum information
	EnvProduction
)

// ColorMode enumeration
type ColorMode int

const (
	// ClrNotSet - No color mode is set (initial)
	ClrNotSet ColorMode = -1 + iota
	// ClrDisabled - Do not use color. Overrides defaults
	ClrDisabled
	// ClrEnabled - Force use of color. Overrides defaults
	ClrEnabled
	// ClrAuto - Use color based on detected (or set) Environment
	ClrAuto
)

// Options allow customization of the logger by the end user
type Options struct {
	Module      string      // Name of running module
	Environment Environment // Override default handling
	UseColor    ColorMode   // Enable color (override) default handling
	SmartError  bool        // Extended error that adapts by environment
	Out         io.Writer   // Where to write output
	FmtProd     string      // for use with production environment
	FmtDev      string      // for use with development environment
}

// NewDefaultOptions returns a new Options object with all defaults
func NewDefaultOptions() *Options {
	return &Options{
		Module:      "unknown",
		Environment: detectEnvironment(true),
		UseColor:    ClrAuto,
		SmartError:  true,
		Out:         os.Stderr,
		FmtProd:     FmtProductionLog,
		FmtDev:      FmtDevelopmentLog,
	}
}

// NewCustomOptions returns a new Options object with all user options
func NewCustomOptions(module string, env Environment, clr ColorMode, SmartError bool, out io.Writer, fmtProd, fmtDev string) *Options {
	o := NewDefaultOptions()

	// If given module is valid use it, otherwise keep default
	if len(module) > 3 {
		o.Module = module
	}

	if env == EnvProduction || env == EnvQuality || env == EnvDevelopment {
		o.Environment = env
	}

	if clr != ClrNotSet {
		o.UseColor = clr
	}

	o.SmartError = SmartError

	if out != nil {
		o.Out = out
	}

	if len(fmtProd) > 10 {
		o.FmtProd = fmtProd
	}

	if len(fmtDev) > 10 {
		o.FmtDev = fmtDev
	}

	return o
}

// EnvAsString returns the current envirnment for options as a string
func (o *Options) EnvAsString() string {
	environments := [...]string{
		"EvnAuto",
		"EnvTesting",
		"EnvDevelopment",
		"EnvQuality",
		"EnvProduction",
	}
	return environments[o.Environment]
}
