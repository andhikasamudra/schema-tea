package catalog

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/andhika-wibawa/db-library/internal/config"
	"github.com/go-sql-driver/mysql"
)

func LoadMySQL(ctx context.Context, definition config.ResolvedDatabase) (Catalog, error) {
	username := definition.UsernameEnv
	password := definition.PasswordEnv
	if username == "" || password == "" {
		return Catalog{}, fmt.Errorf("database %q credentials are not configured", definition.ID)
	}

	port := definition.Port
	if port == 0 {
		port = 3306
	}
	mysqlConfig := mysql.NewConfig()
	mysqlConfig.User = username
	mysqlConfig.Passwd = password
	mysqlConfig.Net = "tcp"
	mysqlConfig.Addr = definition.Host + ":" + strconv.Itoa(port)
	mysqlConfig.DBName = definition.Database.Database
	mysqlConfig.ParseTime = true
	if definition.SSLMode != "disable" {
		mysqlConfig.TLSConfig = "true"
	}

	connectCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	db, err := sql.Open("mysql", mysqlConfig.FormatDSN())
	if err != nil {
		return Catalog{}, fmt.Errorf("open database %q: %w", definition.ID, err)
	}
	defer db.Close()
	if err := db.PingContext(connectCtx); err != nil {
		return Catalog{}, fmt.Errorf("connect to database %q: %w", definition.ID, err)
	}

	fmt.Println("db connection successful with database ", definition.Database.Database)
	fmt.Println("loading tables for schemas ", definition.Schemas)

	result := Catalog{Name: definition.Name}
	for _, schema := range definition.Schemas {
		tables, err := loadMySQLTables(ctx, db, schema)
		if err != nil {
			return Catalog{}, fmt.Errorf("load schema %q from database %q: %w", schema, definition.ID, err)
		}
		result.Tables = append(result.Tables, tables...)
	}
	return result, nil
}

func loadMySQLTables(ctx context.Context, db *sql.DB, schema string) ([]Table, error) {
	names, err := loadMySQLTableNames(ctx, db, schema)
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, nil
	}

	columns, err := loadMySQLColumns(ctx, db, schema)
	if err != nil {
		return nil, err
	}
	relations, err := loadMySQLRelations(ctx, db, schema)
	if err != nil {
		return nil, err
	}
	return assembleSchema(schema, names, columns, relations), nil
}

func loadMySQLTableNames(ctx context.Context, db *sql.DB, schema string) ([]string, error) {
	const tablesQuery = `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = ? AND table_type IN ('BASE TABLE', 'VIEW')
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

func loadMySQLColumns(ctx context.Context, db *sql.DB, schema string) ([]schemaColumn, error) {
	const columnsQuery = `
		SELECT table_name, column_name, data_type, is_nullable, column_key
		FROM information_schema.columns
		WHERE table_schema = ?
		ORDER BY table_name, ordinal_position`

	rows, err := db.QueryContext(ctx, columnsQuery, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []schemaColumn
	for rows.Next() {
		var tableName, columnName, dataType, nullable, key string
		if err := rows.Scan(&tableName, &columnName, &dataType, &nullable, &key); err != nil {
			return nil, err
		}
		columns = append(columns, schemaColumn{
			table:  tableName,
			column: newColumn(columnName, dataType, nullable, key == "PRI"),
		})
	}
	return columns, rows.Err()
}

func loadMySQLRelations(ctx context.Context, db *sql.DB, schema string) ([]schemaRelation, error) {
	const relationsQuery = `
		SELECT table_name, column_name, referenced_table_name, referenced_column_name
		FROM information_schema.key_column_usage
		WHERE table_schema = ?
		  AND referenced_table_name IS NOT NULL
		ORDER BY table_name, column_name`

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

func isMySQL(engine string) bool {
	return strings.EqualFold(engine, "mysql") || strings.EqualFold(engine, "mariadb")
}
