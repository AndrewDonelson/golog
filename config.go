// file: config.go
// This file will load a yaml configuration file named 'golog.yaml' from provided path, parse it and store to Config struct.
package golog

// Config struct
type Config struct{
	Name string `yaml:"name"`
	
}
