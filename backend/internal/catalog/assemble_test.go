package catalog

import "testing"

func TestNewColumnNullabilityAndPrimaryKey(t *testing.T) {
	pk := newColumn("id", "bigint", "NO", true)
	if pk.Name != "id" || pk.Type != "bigint" || pk.Nullable || !pk.PrimaryKey {
		t.Fatalf("primary key column: %+v", pk)
	}

	optional := newColumn("note", "varchar", "YES", false)
	if !optional.Nullable || optional.PrimaryKey {
		t.Fatalf("nullable column: %+v", optional)
	}

	// MySQL reports UNI and MUL in COLUMN_KEY. Only PRI is a primary key.
	if newColumn("email", "varchar", "NO", "UNI" == "PRI").PrimaryKey {
		t.Fatal("unique column was marked primary key")
	}
	if newColumn("org_id", "bigint", "YES", "MUL" == "PRI").PrimaryKey {
		t.Fatal("non-unique index column was marked primary key")
	}
}

func TestAssembleSchemaPreservesOrderAndEndpoints(t *testing.T) {
	tables := assembleSchema("sales", []string{"orders", "customers"}, []schemaColumn{
		{table: "customers", column: newColumn("id", "uuid", "NO", true)},
		{table: "orders", column: newColumn("id", "bigint", "NO", true)},
		{table: "orders", column: newColumn("customer_id", "uuid", "YES", false)},
		{table: "ignored", column: newColumn("extra", "text", "YES", false)},
	}, []schemaRelation{
		{table: "orders", relation: Relation{FromColumn: "customer_id", ToTable: "customers", ToColumn: "id"}},
		{table: "orders", relation: Relation{FromColumn: "status_id", ToTable: "statuses", ToColumn: "id"}},
		{table: "missing", relation: Relation{FromColumn: "a", ToTable: "b", ToColumn: "c"}},
	})

	if len(tables) != 2 {
		t.Fatalf("table count: got %d", len(tables))
	}
	if tables[0].Schema != "sales" || tables[0].Name != "orders" || tables[1].Name != "customers" {
		t.Fatalf("table order: %+v", tables)
	}

	if len(tables[0].Columns) != 2 || tables[0].Columns[0].Name != "id" || !tables[0].Columns[0].PrimaryKey {
		t.Fatalf("orders columns: %+v", tables[0].Columns)
	}
	if tables[0].Columns[1].Name != "customer_id" || !tables[0].Columns[1].Nullable || tables[0].Columns[1].PrimaryKey {
		t.Fatalf("orders foreign-key column: %+v", tables[0].Columns[1])
	}
	if len(tables[0].Relations) != 2 ||
		tables[0].Relations[0].FromColumn != "customer_id" ||
		tables[0].Relations[0].ToTable != "customers" ||
		tables[0].Relations[0].ToColumn != "id" ||
		tables[0].Relations[1].FromColumn != "status_id" {
		t.Fatalf("orders relations: %+v", tables[0].Relations)
	}

	if len(tables[1].Columns) != 1 || tables[1].Columns[0].Name != "id" || tables[1].Relations != nil {
		t.Fatalf("customers: %+v", tables[1])
	}
}

func TestAssembleSchemaEmptyAndUnmatched(t *testing.T) {
	if tables := assembleSchema("public", nil, nil, nil); tables != nil {
		t.Fatalf("expected nil tables, got %#v", tables)
	}

	tables := assembleSchema("public", []string{"alone"}, nil, nil)
	if len(tables) != 1 || tables[0].Name != "alone" || tables[0].Columns != nil || tables[0].Relations != nil {
		t.Fatalf("empty table: %+v", tables)
	}
}
