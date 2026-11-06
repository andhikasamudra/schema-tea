package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Connections []Connection `json:"connections"`
}

type Connection struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Engine      string     `json:"engine"`
	Host        string     `json:"host"`
	Port        int        `json:"port"`
	UsernameEnv string     `json:"username_env"`
	PasswordEnv string     `json:"password_env"`
	SSLMode     string     `json:"sslmode"`
	Databases   []Database `json:"databases"`
}

type Database struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Environment string   `json:"environment"`
	Database    string   `json:"database"`
	Schemas     []string `json:"schemas"`
}

// ResolvedDatabase is the server-side connection target selected by a user.
// It combines one database entry with its parent connection profile.
type ResolvedDatabase struct {
	Database
	ConnectionID   string
	ConnectionName string
	Engine         string
	Host           string
	Port           int
	UsernameEnv    string
	PasswordEnv    string
	SSLMode        string
}

func Load(path string) (Config, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := json.Unmarshal(contents, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	seenConnections := make(map[string]struct{}, len(c.Connections))
	seenDatabases := make(map[string]struct{})
	for _, connection := range c.Connections {
		if connection.ID == "" || connection.Name == "" {
			return fmt.Errorf("connection id and name are required")
		}
		if _, exists := seenConnections[connection.ID]; exists {
			return fmt.Errorf("duplicate connection id %q", connection.ID)
		}
		seenConnections[connection.ID] = struct{}{}
		if connection.Engine != "postgres" && connection.Engine != "postgresql" && connection.Engine != "mysql" && connection.Engine != "mariadb" {
			return fmt.Errorf("connection %q: unsupported engine %q", connection.ID, connection.Engine)
		}
		if connection.Host == "" || connection.UsernameEnv == "" || connection.PasswordEnv == "" {
			return fmt.Errorf("connection %q: host, username_env, and password_env are required", connection.ID)
		}
		if len(connection.Databases) == 0 {
			return fmt.Errorf("connection %q: at least one database is required", connection.ID)
		}
		for _, database := range connection.Databases {
			if database.ID == "" || database.Name == "" || database.Database == "" {
				return fmt.Errorf("connection %q: database id, name, and database are required", connection.ID)
			}
			if _, exists := seenDatabases[database.ID]; exists {
				return fmt.Errorf("duplicate database id %q", database.ID)
			}
			seenDatabases[database.ID] = struct{}{}
			if len(database.Schemas) == 0 {
				return fmt.Errorf("database %q: at least one schema is required", database.ID)
			}
		}
	}
	return nil
}

func (c Config) ResolvedDatabases() []ResolvedDatabase {
	var result []ResolvedDatabase
	for _, connection := range c.Connections {
		for _, database := range connection.Databases {
			result = append(result, ResolvedDatabase{
				Database:       database,
				ConnectionID:   connection.ID,
				ConnectionName: connection.Name,
				Engine:         connection.Engine,
				Host:           connection.Host,
				Port:           connection.Port,
				UsernameEnv:    connection.UsernameEnv,
				PasswordEnv:    connection.PasswordEnv,
				SSLMode:        connection.SSLMode,
			})
		}
	}
	return result
}
