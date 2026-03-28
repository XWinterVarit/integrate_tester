package repository

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

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

	query := fmt.Sprintf("SELECT ROWID, %s FROM %s.%s", cols, r.schema, table)

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
