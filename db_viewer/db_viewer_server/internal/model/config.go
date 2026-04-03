package model

type Config struct {
	Server  ServerConfig   `yaml:"server"`
	Clients []ClientConfig `yaml:"clients"`
}

type ServerConfig struct {
	Port       int    `yaml:"port"`
	CORSOrigin string `yaml:"cors_origin"`
}

type ClientConfig struct {
	Name     string        `yaml:"name"`
	User     string        `yaml:"user"`
	Password string        `yaml:"password"`
	Host     string        `yaml:"host"`
	Port     int           `yaml:"port"`
	Service  string        `yaml:"service"`
	Schema   string        `yaml:"schema"`
	Tables   []TableConfig `yaml:"tables"`
}

type TableConfig struct {
	Name          string               `yaml:"name"`
	PresetFilters []PresetFilterConfig `yaml:"preset_filters"`
	PresetQueries []PresetQueryConfig  `yaml:"preset_queries"`
}

type PresetFilterConfig struct {
	Name    string   `yaml:"name"`
	Details string   `yaml:"details"`
	Columns []string `yaml:"columns"`
}

type PresetQueryConfig struct {
	Name      string                 `yaml:"name"`
	Query     string                 `yaml:"query"`
	Arguments []PresetQueryArgConfig `yaml:"arguments"`
}

type PresetQueryArgConfig struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
}
