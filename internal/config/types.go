package config

type Config struct {
	Database Database `json:"database"`
	Server   Server   `json:"server"`
	JWT      JWT      `json:"jwt"`
	Log      Log      `json:"log"`
	Env      string   `json:"env"`
}

type Database struct {
	URL string `json:"url"`
}

type Server struct {
	Port string `json:"port"`
}

type JWT struct {
	PrivateKey string `json:"private_key"`
}

type Log struct {
	Level    string `json:"level"`
	FilePath string `json:"file_path"`
}