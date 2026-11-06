package catalog

// schemaColumn is one column row from a schema-wide information_schema read.
type schemaColumn struct {
	table  string
	column Column
}

// schemaRelation is one foreign-key row from a schema-wide information_schema read.
type schemaRelation struct {
	table    string
	relation Relation
}

// newColumn maps an information_schema column onto the catalog model.
// primaryKey is true for MySQL column_key PRI and for Postgres primary-key columns.
func newColumn(name, dataType, nullable string, primaryKey bool) Column {
	return Column{
		Name:       name,
		Type:       dataType,
		Nullable:   nullable == "YES",
		PrimaryKey: primaryKey,
	}
}

// assembleSchema attaches bulk-loaded columns and relations to tables in one schema.
// tableNames is the catalog order. Columns and relations are appended in the order
// given, which the loaders establish with ORDER BY.
func assembleSchema(schema string, tableNames []string, columns []schemaColumn, relations []schemaRelation) []Table {
	if len(tableNames) == 0 {
		return nil
	}

	tables := make([]Table, len(tableNames))
	index := make(map[string]int, len(tableNames))
	for i, name := range tableNames {
		tables[i] = Table{Schema: schema, Name: name}
		index[name] = i
	}
	for _, column := range columns {
		i, ok := index[column.table]
		if !ok {
			continue
		}
		tables[i].Columns = append(tables[i].Columns, column.column)
	}
	for _, relation := range relations {
		i, ok := index[relation.table]
		if !ok {
			continue
		}
		tables[i].Relations = append(tables[i].Relations, relation.relation)
	}
	return tables
}
