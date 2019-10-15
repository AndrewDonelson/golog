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

type Environment int

const (
	EnvNotSet Environment = -2 + iota
	EnvTesting
	EnvDevelopment
	EnvQuality
	EnvProduction
)

type ColorMode int

const (
	ClrNotSet ColorMode = -1 + iota
	ClrDisabled
	ClrEnabled
	ClrAuto
)

// Options allow customization of the logger by the end user
type Options struct {
	Module      string      // Name of running module
	Environment Environment // Override default handling
	UseColor    ColorMode   // Enable color (override) default handling
	Out         io.Writer   // Where to write output
	FmtProd     string      // for use with production environment
	FmtDev      string      // for use with development environment
}

// NewDefaultOptions returns a new Options object with all defaults
func NewDefaultOptions() *Options {
	return &Options{
		Module:      "unknown",
		Environment: EnvNotSet,
		UseColor:    ClrAuto,
		Out:         os.Stderr,
		FmtProd:     defProductionFmt,
		FmtDev:      defDevelopmentFmt,
	}
}

// NewCustomOptions returns a new Options object with all user options
func NewCustomOptions(module string, env Environment, clr ColorMode, out io.Writer, fmtProd, fmtDev string) *Options {
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
