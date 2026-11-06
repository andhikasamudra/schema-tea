package catalog

func Sample() Catalog {
	return Catalog{
		Name: "Risk Platform",
		Tables: []Table{
			{
				Schema:      "risk",
				Name:        "applications",
				Description: "A credit application submitted for assessment.",
				Columns: []Column{
					{Name: "id", Type: "uuid", Nullable: false, PrimaryKey: true, Description: "Stable identifier for the application."},
					{Name: "customer_id", Type: "uuid", Nullable: false, Description: "Customer being assessed."},
					{Name: "status", Type: "text", Nullable: false, Description: "Current lifecycle status of the application."},
				},
				Relations: []Relation{{FromColumn: "customer_id", ToTable: "customers", ToColumn: "id"}},
			},
			{
				Schema:      "risk",
				Name:        "customers",
				Description: "A customer subject to one or more risk assessments.",
				Columns: []Column{
					{Name: "id", Type: "uuid", Nullable: false, PrimaryKey: true, Description: "Stable identifier for the customer."},
					{Name: "segment", Type: "text", Nullable: false, Description: "Business segment used by risk policies."},
				},
			},
		},
	}
}
