package catalog

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/andhika-wibawa/db-library/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func LoadPostgres(ctx context.Context, definition config.ResolvedDatabase) (Catalog, error) {
	username := os.Getenv(definition.UsernameEnv)
	password := os.Getenv(definition.PasswordEnv)
	if username == "" || password == "" {
		return Catalog{}, fmt.Errorf("database %q credentials are not configured", definition.ID)
	}

	sslMode := definition.SSLMode
	if sslMode == "" {
		sslMode = "require"
	}
	port := definition.Port
	if port == 0 {
		port = 5432
	}
	dsn := (&url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(username, password),
		Host:   definition.Host + ":" + strconv.Itoa(port),
		Path:   "/" + definition.Database.Database,
		RawQuery: url.Values{
			"sslmode": []string{sslMode},
		}.Encode(),
	}).String()

	connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return Catalog{}, fmt.Errorf("open database %q: %w", definition.ID, err)
	}
	defer db.Close()
	if err := db.PingContext(connectCtx); err != nil {
		return Catalog{}, fmt.Errorf("connect to database %q: %w", definition.ID, err)
	}

	catalog := Catalog{Name: definition.Name}
	for _, schema := range definition.Schemas {
		tables, err := loadTables(ctx, db, schema)
		if err != nil {
			return Catalog{}, fmt.Errorf("load schema %q from database %q: %w", schema, definition.ID, err)
		}
		catalog.Tables = append(catalog.Tables, tables...)
	}
	return catalog, nil
}

func loadTables(ctx context.Context, db *sql.DB, schema string) ([]Table, error) {
	names, err := loadPostgresTableNames(ctx, db, schema)
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, nil
	}

	columns, err := loadColumns(ctx, db, schema)
	if err != nil {
		return nil, err
	}
	relations, err := loadRelations(ctx, db, schema)
	if err != nil {
		return nil, err
	}
	return assembleSchema(schema, names, columns, relations), nil
}

func loadPostgresTableNames(ctx context.Context, db *sql.DB, schema string) ([]string, error) {
	const tablesQuery = `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1 AND table_type IN ('BASE TABLE', 'VIEW')
		ORDER BY table_name`

	rows, err := db.QueryContext(ctx, tablesQuery, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

func loadColumns(ctx context.Context, db *sql.DB, schema string) ([]schemaColumn, error) {
	const columnsQuery = `
		SELECT c.table_name, c.column_name, c.data_type, c.is_nullable,
		       (pk.column_name IS NOT NULL) AS is_primary_key
		FROM information_schema.columns c
		LEFT JOIN (
			SELECT DISTINCT kcu.table_name, kcu.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu
			  ON tc.constraint_name = kcu.constraint_name
			 AND tc.table_schema = kcu.table_schema
			 AND tc.table_name = kcu.table_name
			WHERE tc.constraint_type = 'PRIMARY KEY'
			  AND tc.table_schema = $1
		) pk
		  ON pk.table_name = c.table_name
		 AND pk.column_name = c.column_name
		WHERE c.table_schema = $1
		ORDER BY c.table_name, c.ordinal_position`

	rows, err := db.QueryContext(ctx, columnsQuery, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []schemaColumn
	for rows.Next() {
		var tableName, columnName, dataType, nullable string
		var primaryKey bool
		if err := rows.Scan(&tableName, &columnName, &dataType, &nullable, &primaryKey); err != nil {
			return nil, err
		}
		columns = append(columns, schemaColumn{
			table:  tableName,
			column: newColumn(columnName, dataType, nullable, primaryKey),
		})
	}
	return columns, rows.Err()
}

func loadRelations(ctx context.Context, db *sql.DB, schema string) ([]schemaRelation, error) {
	const relationsQuery = `
		SELECT tc.table_name, kcu.column_name, ccu.table_name, ccu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
		  ON tc.constraint_name = kcu.constraint_name
		 AND tc.table_schema = kcu.table_schema
		 AND tc.table_name = kcu.table_name
		JOIN information_schema.constraint_column_usage ccu
		  ON ccu.constraint_name = tc.constraint_name
		 AND ccu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY'
		  AND tc.table_schema = $1
		ORDER BY tc.table_name, kcu.column_name`

	rows, err := db.QueryContext(ctx, relationsQuery, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relations []schemaRelation
	for rows.Next() {
		var tableName string
		var relation Relation
		if err := rows.Scan(&tableName, &relation.FromColumn, &relation.ToTable, &relation.ToColumn); err != nil {
			return nil, err
		}
		relations = append(relations, schemaRelation{table: tableName, relation: relation})
	}
	return relations, rows.Err()
}

func SafeDatabaseSummary(definition config.ResolvedDatabase) map[string]any {
	return map[string]any{
		"id":              definition.ID,
		"name":            definition.Name,
		"engine":          strings.ToLower(definition.Engine),
		"environment":     definition.Environment,
		"schemas":         definition.Schemas,
		"connection_id":   definition.ConnectionID,
		"connection_name": definition.ConnectionName,
	}
}
