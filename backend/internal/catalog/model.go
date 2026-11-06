package catalog

// Catalog is the documentation model rendered by the prototype UI.
type Catalog struct {
	Name   string  `json:"name"`
	Tables []Table `json:"tables"`
}

type Table struct {
	Schema      string     `json:"schema"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Columns     []Column   `json:"columns"`
	Relations   []Relation `json:"relations"`
}

type Column struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Nullable    bool   `json:"nullable"`
	PrimaryKey  bool   `json:"primary_key"`
	Description string `json:"description"`
}

type Relation struct {
	FromColumn string `json:"from_column"`
	ToTable    string `json:"to_table"`
	ToColumn   string `json:"to_column"`
}
