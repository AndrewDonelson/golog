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

// Options allow customization of the logger by the end user
type Options struct {
	Module      string    // Name of running module
	Environment int       // Override default handling
	UseColor    bool      // Enable color (override) default handling
	Out         io.Writer // Where to write output
	FmtProd     string    // for use with production environment
	FmtDev      string    // for use with development environment
}

// NewDefaultOptions returns a new Options object with all defaults
func NewDefaultOptions() *Options {
	return &Options{
		Module:      "unknown",
		Environment: 0,
		UseColor:    false,
		Out:         os.Stderr,
		FmtProd:     defProductionFmt,
		FmtDev:      defDevelopmentFmt,
	}
}

// NewCustomOptions returns a new Options object with all user options
func NewCustomOptions(module string, environment int, useColor bool, out io.Writer, fmtProd, fmtDev string) *Options {
	return &Options{
		Module:      module,
		Environment: environment,
		UseColor:    useColor,
		Out:         out,
		FmtProd:     fmtProd,
		FmtDev:      fmtDev,
	}
}
