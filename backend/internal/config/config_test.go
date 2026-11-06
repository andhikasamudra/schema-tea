package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "databases.json")
	contents := `{"connections":[{"id":"risk-cluster","name":"Risk Cluster","engine":"postgres","host":"localhost","username_env":"RISK_USER","password_env":"RISK_PASS","databases":[{"id":"risk","name":"Risk","database":"risk","schemas":["risk"]}]}]}`
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.ResolvedDatabases()) != 1 || cfg.ResolvedDatabases()[0].ID != "risk" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestValidateRejectsDuplicateIDs(t *testing.T) {
	cfg := Config{Connections: []Connection{
		{ID: "risk", Name: "Risk", Engine: "postgres", Host: "localhost", UsernameEnv: "U", PasswordEnv: "P", Databases: []Database{{ID: "risk-db", Name: "Risk", Database: "risk", Schemas: []string{"risk"}}}},
		{ID: "other", Name: "Other", Engine: "postgres", Host: "localhost", UsernameEnv: "U2", PasswordEnv: "P2", Databases: []Database{{ID: "risk-db", Name: "Other", Database: "other", Schemas: []string{"public"}}}},
	}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected duplicate ID validation error")
	}
}
