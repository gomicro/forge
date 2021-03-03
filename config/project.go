package config

// Project represents metadata about the project being built that can be
// included in the config file.
type Project struct {
	Name string `yaml:"name"`
}
