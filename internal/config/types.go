package config

type Config struct {
	Env      string   `json:"env"`
	AppName  string   `json:"app_name"`
	Timezone string   `json:"timezone"`
	Server   Server   `json:"server"`
	Database Database `json:"database"`
	JWT      JWT      `json:"jwt"`
	Log      Log      `json:"log"`
}

type Server struct {
	Port               string   `json:"port"`
	CORSAllowedOrigins []string `json:"cors_allowed_origins"`
}

type Database struct {
	URL string `json:"url"`
}

type JWT struct {
	PrivateKey      string `json:"private_key"`
	PublicKey       string `json:"public_key"`
	AccessTokenTTL  int    `json:"access_token_ttl"`
	RefreshTokenTTL int    `json:"refresh_token_ttl"`
}

type Log struct {
	Level  string `json:"level"`
	Format string `json:"format"`
	File   string `json:"file"`
}

// GetCORSOrigins returns the list of allowed CORS origins
func (c *Config) GetCORSOrigins() []string {
	return c.Server.CORSAllowedOrigins
}

// GetEnv returns the environment name
func (c *Config) GetEnv() string {
	return c.Env
}
