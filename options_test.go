package golog

import (
	"testing"
)

func TestEnvStrings(t *testing.T) {
	var (
		s string
	)

	o := Options{}
	o.Environment = EnvAuto
	s = o.EnvAsString()
	if s != "EvnAuto" {
		t.Errorf("\nWant: %sHave: %s", "EvnAuto", s)
	}

	o.Environment = EnvDevelopment
	s = o.EnvAsString()
	if s != "EnvDevelopment" {
		t.Errorf("\nWant: %sHave: %s", "EnvDevelopment", s)
	}

	o.Environment = EnvQuality
	s = o.EnvAsString()
	if s != "EnvQuality" {
		t.Errorf("\nWant: %sHave: %s", "EnvQuality", s)
	}

	o.Environment = EnvProduction
	s = o.EnvAsString()
	if s != "EnvProduction" {
		t.Errorf("\nWant: %sHave: %s", "EnvProduction", s)
	}

}
