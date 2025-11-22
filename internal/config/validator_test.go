package config

import (
	"testing"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Database: Database{
					URL: "postgres://user:pass@localhost:5432/db",
				},
				JWT: JWT{
					PrivateKey:      "-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----",
					PublicKey:       "-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----",
					AccessTokenTTL:  900,
					RefreshTokenTTL: 604800,
				},
				Server: Server{
					Port:               "8080",
					CORSAllowedOrigins: []string{"http://localhost:3000"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing database URL",
			config: &Config{
				JWT: JWT{
					PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----",
				},
				Server: Server{
					Port: "8080",
				},
			},
			wantErr: true,
		},
		{
			name: "missing JWT private key",
			config: &Config{
				Database: Database{
					URL: "postgres://user:pass@localhost:5432/db",
				},
				Server: Server{
					Port: "8080",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: &Config{
				Database: Database{
					URL: "postgres://user:pass@localhost:5432/db",
				},
				JWT: JWT{
					PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----",
				},
				Server: Server{
					Port: "",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDatabaseConfig(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid postgres URL",
			url:     "postgres://user:pass@localhost:5432/db",
			wantErr: false,
		},
		{
			name:    "valid postgresql URL",
			url:     "postgresql://user:pass@localhost:5432/db",
			wantErr: false,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "invalid scheme",
			url:     "mysql://user:pass@localhost:3306/db",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Database{URL: tt.url}
			err := config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DatabaseConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
