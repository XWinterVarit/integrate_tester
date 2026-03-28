package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/XWinterVarit/integrate_tester/db_viewer/backend/internal/model"
)

type OracleRepository struct {
	db     *sql.DB
	schema string
}

func NewOracle(db *sql.DB, schema string) *OracleRepository {
	return &OracleRepository{db: db, schema: schema}
}

func (r *OracleRepository) QueryRows(ctx context.Context, params model.RowQueryParams, table string) ([]map[string]any, error) {
	cols := "*"
	if params.Select != "" {
		cols = params.Select
	}

	query := fmt.Sprintf("SELECT %s FROM %s.%s", cols, r.schema, table)

	if params.Sort != "" {
		dir := "ASC"
		if params.SortDir == "desc" {
			dir = "DESC"
		}
		query += fmt.Sprintf(" ORDER BY %s %s", params.Sort, dir)
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 100
	}
	query += fmt.Sprintf(" FETCH FIRST %d ROWS ONLY", limit)

	return r.executeQuery(ctx, query)
}

func (r *OracleRepository) ExecuteRawQuery(ctx context.Context, query string, args []any) ([]map[string]any, error) {
	return r.executeQueryWithArgs(ctx, query, args)
}

func (r *OracleRepository) GetColumns(ctx context.Context, table string) ([]map[string]any, error) {
	query := `SELECT COLUMN_NAME, DATA_TYPE, DATA_LENGTH, NULLABLE, DATA_DEFAULT
		FROM ALL_TAB_COLUMNS WHERE OWNER = :1 AND TABLE_NAME = :2 ORDER BY COLUMN_ID`
	return r.executeQueryWithArgs(ctx, query, []any{r.schema, table})
}

func (r *OracleRepository) GetConstraints(ctx context.Context, table string) ([]map[string]any, error) {
	query := `SELECT c.CONSTRAINT_NAME, c.CONSTRAINT_TYPE, c.STATUS,
		LISTAGG(cc.COLUMN_NAME, ', ') WITHIN GROUP (ORDER BY cc.POSITION) AS COLUMNS
		FROM ALL_CONSTRAINTS c
		JOIN ALL_CONS_COLUMNS cc ON c.CONSTRAINT_NAME = cc.CONSTRAINT_NAME AND c.OWNER = cc.OWNER
		WHERE c.OWNER = :1 AND c.TABLE_NAME = :2
		GROUP BY c.CONSTRAINT_NAME, c.CONSTRAINT_TYPE, c.STATUS`
	return r.executeQueryWithArgs(ctx, query, []any{r.schema, table})
}

func (r *OracleRepository) GetIndexes(ctx context.Context, table string) ([]map[string]any, error) {
	query := `SELECT i.INDEX_NAME, i.INDEX_TYPE, i.UNIQUENESS,
		LISTAGG(ic.COLUMN_NAME, ', ') WITHIN GROUP (ORDER BY ic.COLUMN_POSITION) AS COLUMNS
		FROM ALL_INDEXES i
		JOIN ALL_IND_COLUMNS ic ON i.INDEX_NAME = ic.INDEX_NAME AND i.OWNER = ic.INDEX_OWNER
		WHERE i.OWNER = :1 AND i.TABLE_NAME = :2
		GROUP BY i.INDEX_NAME, i.INDEX_TYPE, i.UNIQUENESS`
	return r.executeQueryWithArgs(ctx, query, []any{r.schema, table})
}

func (r *OracleRepository) GetTableSize(ctx context.Context, table string) ([]map[string]any, error) {
	query := fmt.Sprintf(`SELECT s.BYTES, s.BLOCKS, (SELECT COUNT(*) FROM %s.%s) AS ROW_COUNT
		FROM ALL_SEGMENTS s WHERE s.OWNER = :1 AND s.SEGMENT_NAME = :2`, r.schema, table)
	return r.executeQueryWithArgs(ctx, query, []any{r.schema, table})
}

func (r *OracleRepository) UpdateCell(ctx context.Context, table, column, value, whereCol, whereVal string) error {
	query := fmt.Sprintf("UPDATE %s.%s SET %s = :1 WHERE %s = :2",
		r.schema, table, column, whereCol)
	_, err := r.db.ExecContext(ctx, query, value, whereVal)
	return err
}

func (r *OracleRepository) ExportRows(ctx context.Context, query string) ([]map[string]any, error) {
	return r.executeQuery(ctx, query)
}

func (r *OracleRepository) Schema() string {
	return r.schema
}

func (r *OracleRepository) executeQuery(ctx context.Context, query string) ([]map[string]any, error) {
	return r.executeQueryWithArgs(ctx, query, nil)
}

func (r *OracleRepository) executeQueryWithArgs(ctx context.Context, query string, args []any) ([]map[string]any, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

func scanRows(rows *sql.Rows) ([]map[string]any, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []map[string]any
	for rows.Next() {
		values := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
		row := make(map[string]any)
		for i, col := range cols {
			switch v := values[i].(type) {
			case []byte:
				row[col] = string(v)
			default:
				row[col] = v
			}
		}
		result = append(result, row)
	}
	return result, rows.Err()
}
