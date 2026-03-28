package repository

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
	"time"

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

	var query string
	if cols == "*" {
		query = fmt.Sprintf("SELECT t.ROWID, t.* FROM %s.%s t", r.schema, table)
	} else {
		query = fmt.Sprintf("SELECT ROWID, %s FROM %s.%s", cols, r.schema, table)
	}

	if params.Where != "" {
		query += fmt.Sprintf(" WHERE %s", params.Where)
	}

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
	if params.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d ROWS", params.Offset)
	}
	query += fmt.Sprintf(" FETCH NEXT %d ROWS ONLY", limit)

	return r.executeQuery(ctx, query)
}

func (r *OracleRepository) ExecuteRawQuery(ctx context.Context, query string, args []any) ([]map[string]any, error) {
	return r.executeQueryWithArgs(ctx, query, args)
}

func (r *OracleRepository) GetColumns(ctx context.Context, table string) ([]map[string]any, error) {
	query := `SELECT COLUMN_NAME, DATA_TYPE, DATA_LENGTH, NULLABLE, DATA_DEFAULT
		FROM USER_TAB_COLUMNS WHERE TABLE_NAME = :1 ORDER BY COLUMN_ID`
	return r.executeQueryWithArgs(ctx, query, []any{table})
}

func (r *OracleRepository) GetConstraints(ctx context.Context, table string) ([]map[string]any, error) {
	query := `SELECT c.CONSTRAINT_NAME, c.CONSTRAINT_TYPE, c.STATUS,
		LISTAGG(cc.COLUMN_NAME, ', ') WITHIN GROUP (ORDER BY cc.POSITION) AS COLUMNS
		FROM USER_CONSTRAINTS c
		JOIN USER_CONS_COLUMNS cc ON c.CONSTRAINT_NAME = cc.CONSTRAINT_NAME
		WHERE c.TABLE_NAME = :1
		GROUP BY c.CONSTRAINT_NAME, c.CONSTRAINT_TYPE, c.STATUS`
	return r.executeQueryWithArgs(ctx, query, []any{table})
}

func (r *OracleRepository) GetIndexes(ctx context.Context, table string) ([]map[string]any, error) {
	query := `SELECT i.INDEX_NAME, i.INDEX_TYPE, i.UNIQUENESS,
		LISTAGG(ic.COLUMN_NAME, ', ') WITHIN GROUP (ORDER BY ic.COLUMN_POSITION) AS COLUMNS
		FROM USER_INDEXES i
		JOIN USER_IND_COLUMNS ic ON i.INDEX_NAME = ic.INDEX_NAME
		WHERE i.TABLE_NAME = :1
		GROUP BY i.INDEX_NAME, i.INDEX_TYPE, i.UNIQUENESS`
	return r.executeQueryWithArgs(ctx, query, []any{table})
}

func (r *OracleRepository) GetTableSize(ctx context.Context, table string) ([]map[string]any, error) {
	query := `SELECT
		NVL((SELECT s.BYTES FROM USER_SEGMENTS s WHERE s.SEGMENT_NAME = :1), 0) AS BYTES,
		NVL((SELECT s.BLOCKS FROM USER_SEGMENTS s WHERE s.SEGMENT_NAME = :1), 0) AS BLOCKS,
		(SELECT COUNT(*) FROM ` + table + `) AS ROW_COUNT
		FROM DUAL`
	return r.executeQueryWithArgs(ctx, query, []any{table})
}

func (r *OracleRepository) UpdateCell(ctx context.Context, table, column, value, rowid string) error {
	query := fmt.Sprintf("UPDATE %s.%s SET %s = :1 WHERE ROWID = :2",
		r.schema, table, column)
	_, err := r.db.ExecContext(ctx, query, value, rowid)
	return err
}

func (r *OracleRepository) DeleteRow(ctx context.Context, table, rowid string) error {
	query := fmt.Sprintf("DELETE FROM %s.%s WHERE ROWID = :1", r.schema, table)
	_, err := r.db.ExecContext(ctx, query, rowid)
	return err
}

// isoDateFormats lists common ISO date/timestamp formats returned by Oracle via go-ora.
var isoDateFormats = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

// parseISODate tries to parse s as an ISO date/timestamp and returns a time.Time if successful.
func parseISODate(s string) (time.Time, bool) {
	for _, layout := range isoDateFormats {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

func (r *OracleRepository) InsertRow(ctx context.Context, table string, columns, values []string) error {
	// Fetch virtual columns for this table so we can skip them in the INSERT.
	virtualCols := map[string]bool{}
	rows, err := r.db.QueryContext(ctx,
		`SELECT COLUMN_NAME FROM USER_TAB_COLUMNS WHERE TABLE_NAME = :1 AND VIRTUAL_COLUMN = 'YES'`,
		table)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var col string
			if rows.Scan(&col) == nil {
				virtualCols[col] = true
			}
		}
	}

	filteredCols := make([]string, 0, len(columns))
	filteredVals := make([]any, 0, len(values))
	for i, col := range columns {
		if virtualCols[col] {
			continue
		}
		filteredCols = append(filteredCols, col)
		// Try to parse ISO date/timestamp strings so go-ora binds them correctly.
		if t, ok := parseISODate(values[i]); ok {
			filteredVals = append(filteredVals, t)
		} else {
			filteredVals = append(filteredVals, values[i])
		}
	}

	placeholders := make([]string, len(filteredCols))
	for i := range filteredCols {
		placeholders[i] = fmt.Sprintf(":%d", i+1)
	}
	query := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s)",
		r.schema, table, strings.Join(filteredCols, ", "), strings.Join(placeholders, ", "))
	_, err = r.db.ExecContext(ctx, query, filteredVals...)
	return err
}

func (r *OracleRepository) BuildDeleteQuery(table, rowid string) string {
	return fmt.Sprintf("DELETE FROM %s.%s WHERE ROWID = '%s'", r.schema, table, rowid)
}

func (r *OracleRepository) BuildInsertQuery(table string, columns, values []string) string {
	quotedVals := make([]string, len(values))
	for i, v := range values {
		if v == "" {
			quotedVals[i] = "NULL"
		} else {
			quotedVals[i] = "'" + strings.ReplaceAll(v, "'", "''") + "'"
		}
	}
	return fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s)",
		r.schema, table, strings.Join(columns, ", "), strings.Join(quotedVals, ", "))
}

func (r *OracleRepository) ExportRows(ctx context.Context, query string) ([]map[string]any, error) {
	return r.executeQuery(ctx, query)
}

func (r *OracleRepository) GetBlobData(ctx context.Context, table, column, rowid string) ([]byte, error) {
	query := fmt.Sprintf("SELECT %s FROM %s.%s WHERE ROWID = :1", column, r.schema, table)
	row := r.db.QueryRowContext(ctx, query, rowid)
	var data []byte
	if err := row.Scan(&data); err != nil {
		return nil, err
	}
	return data, nil
}

func (r *OracleRepository) GetRowCount(ctx context.Context, table string) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) AS CNT FROM %s.%s", r.schema, table)
	rows, err := r.executeQuery(ctx, query)
	if err != nil {
		return 0, err
	}
	if len(rows) > 0 {
		if v, ok := rows[0]["CNT"]; ok {
			switch n := v.(type) {
			case float64:
				return int(n), nil
			case int64:
				return int(n), nil
			case string:
				var c int
				fmt.Sscanf(n, "%d", &c)
				return c, nil
			}
		}
	}
	return 0, nil
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

const blobTruncateSize = 200

func scanRows(rows *sql.Rows) ([]map[string]any, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Detect BLOB columns via column type info
	blobCols := make(map[int]bool)
	colTypes, err := rows.ColumnTypes()
	if err == nil {
		for i, ct := range colTypes {
			dbType := strings.ToUpper(ct.DatabaseTypeName())
			if dbType == "BLOB" || dbType == "RAW" || dbType == "LONG RAW" {
				blobCols[i] = true
			}
		}
	}

	var blobColNames []string
	for i, col := range cols {
		if blobCols[i] {
			blobColNames = append(blobColNames, col)
		}
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
			if blobCols[i] {
				row[col] = formatBlobValue(values[i])
			} else {
				switch v := values[i].(type) {
				case []byte:
					row[col] = string(v)
				default:
					row[col] = v
				}
			}
		}
		if len(blobColNames) > 0 {
			row["__blob_columns"] = blobColNames
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func formatBlobValue(v any) string {
	var data []byte
	switch val := v.(type) {
	case []byte:
		data = val
	case io.Reader:
		data, _ = io.ReadAll(val)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
	if len(data) == 0 {
		return ""
	}
	truncated := len(data) > blobTruncateSize
	if truncated {
		data = data[:blobTruncateSize]
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	if truncated {
		encoded += "..."
	}
	return encoded
}
