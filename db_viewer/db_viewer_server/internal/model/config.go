package model

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	DataStore DataStoreConfig `yaml:"data_store"`
}

type ServerConfig struct {
	Port       int    `yaml:"port"`
	CORSOrigin string `yaml:"cors_origin"`
}

// DataStoreConfig is the single server-side Oracle connection used to store app data (DB_VIEWER_APP_DATA).
// It is not exposed to the web UI.
type DataStoreConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Service  string `yaml:"service"`
	Schema   string `yaml:"schema"`
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
	Name string `yaml:"name"`
}
