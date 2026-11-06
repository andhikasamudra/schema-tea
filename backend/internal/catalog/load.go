package catalog

import (
	"context"
	"fmt"
	"strings"

	"github.com/andhika-wibawa/db-library/internal/config"
)

func Load(ctx context.Context, definition config.ResolvedDatabase) (Catalog, error) {
	switch {
	case strings.EqualFold(definition.Engine, "postgres"), strings.EqualFold(definition.Engine, "postgresql"):
		return LoadPostgres(ctx, definition)
	case strings.EqualFold(definition.Engine, "mysql"), strings.EqualFold(definition.Engine, "mariadb"):
		return LoadMySQL(ctx, definition)
	default:
		return Catalog{}, fmt.Errorf("unsupported engine %q", definition.Engine)
	}
}
