package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andhika-wibawa/db-library/internal/catalog"
	"github.com/andhika-wibawa/db-library/internal/config"
)

func main() {
	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	configPath := os.Getenv("DBLIB_CONFIG")
	if configPath == "" {
		configPath = filepath.Join(root, "config", "databases.json")
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/databases", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		summaries := make([]map[string]any, 0, len(cfg.Connections))
		resolvedDatabases := cfg.ResolvedDatabases()
		for _, connection := range cfg.Connections {
			databases := make([]map[string]any, 0, len(connection.Databases))
			for _, definition := range resolvedDatabases {
				if definition.ConnectionID == connection.ID {
					databases = append(databases, catalog.SafeDatabaseSummary(definition))
				}
			}
			summaries = append(summaries, map[string]any{
				"id":        connection.ID,
				"name":      connection.Name,
				"engine":    connection.Engine,
				"databases": databases,
			})
		}
		writeJSON(w, http.StatusOK, summaries)
	})

	mux.HandleFunc("/api/databases/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) != 4 || parts[0] != "api" || parts[1] != "databases" || parts[3] != "catalog" {
			writeError(w, http.StatusNotFound, "not found")
			return
		}

		definition, ok := findDatabase(cfg, parts[2])
		if !ok {
			writeError(w, http.StatusNotFound, "database not found")
			return
		}

		fmt.Println("definition", definition)

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		result, err := catalog.Load(ctx, definition)
		if err != nil {
			log.Printf("load catalog %q: %v", definition.ID, err)
			writeError(w, http.StatusBadGateway, "unable to load database catalog")
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	webDir := filepath.Join(root, "web", "dist")
	mux.Handle("/", http.FileServer(http.Dir(webDir)))

	addr := ":8080"
	fmt.Printf("DB Library listening on http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func findDatabase(cfg config.Config, id string) (config.ResolvedDatabase, bool) {
	for _, definition := range cfg.ResolvedDatabases() {
		if definition.ID == id {
			return definition, true
		}
	}
	return config.ResolvedDatabase{}, false
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
