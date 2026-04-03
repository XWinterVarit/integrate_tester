package repository

import (
	"context"
	"database/sql"
	"fmt"
)

// AppDataRow represents a row in the DB_VIEWER_APP_DATA table.
type AppDataRow struct {
	ID          int64
	Feature     string
	ScopeClient string
	ScopeTable  string
	ItemKey     string
	Data        string
}

// AppDataRepository provides CRUD operations for the DB_VIEWER_APP_DATA table.
type AppDataRepository struct {
	db *sql.DB
}

func NewAppDataRepository(db *sql.DB) *AppDataRepository {
	return &AppDataRepository{db: db}
}

// EnsureTable creates the DB_VIEWER_APP_DATA table and indexes if they don't exist.
func (r *AppDataRepository) EnsureTable(ctx context.Context) error {
	ddl := `
DECLARE
  v_cnt NUMBER;
BEGIN
  SELECT COUNT(*) INTO v_cnt FROM USER_TABLES WHERE TABLE_NAME = 'DB_VIEWER_APP_DATA';
  IF v_cnt = 0 THEN
    EXECUTE IMMEDIATE '
      CREATE TABLE DB_VIEWER_APP_DATA (
        ID            NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
        FEATURE       VARCHAR2(50)   NOT NULL,
        SCOPE_CLIENT  VARCHAR2(100),
        SCOPE_TABLE   VARCHAR2(100),
        ITEM_KEY      VARCHAR2(200),
        DATA          CLOB,
        CREATED_AT    TIMESTAMP DEFAULT SYSTIMESTAMP,
        UPDATED_AT    TIMESTAMP DEFAULT SYSTIMESTAMP
      )';
    EXECUTE IMMEDIATE 'CREATE INDEX IDX_APPDATA_FEATURE ON DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE)';
    EXECUTE IMMEDIATE 'CREATE UNIQUE INDEX IDX_APPDATA_UNIQUE ON DB_VIEWER_APP_DATA (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY)';
  END IF;
END;`
	_, err := r.db.ExecContext(ctx, ddl)
	return err
}

// List returns all rows matching the given feature, client, and table scope.
func (r *AppDataRepository) List(ctx context.Context, feature, client, table string) ([]AppDataRow, error) {
	query := `SELECT ID, FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA
		FROM DB_VIEWER_APP_DATA
		WHERE FEATURE = :1 AND SCOPE_CLIENT = :2 AND SCOPE_TABLE = :3
		ORDER BY ID`
	rows, err := r.db.QueryContext(ctx, query, feature, client, table)
	if err != nil {
		return nil, fmt.Errorf("appdata list: %w", err)
	}
	defer rows.Close()
	return scanAppDataRows(rows)
}

// Get returns a single row by feature, client, table, and key.
func (r *AppDataRepository) Get(ctx context.Context, feature, client, table, key string) (AppDataRow, error) {
	query := `SELECT ID, FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA
		FROM DB_VIEWER_APP_DATA
		WHERE FEATURE = :1 AND SCOPE_CLIENT = :2 AND SCOPE_TABLE = :3 AND ITEM_KEY = :4`
	row := r.db.QueryRowContext(ctx, query, feature, client, table, key)
	var d AppDataRow
	err := row.Scan(&d.ID, &d.Feature, &d.ScopeClient, &d.ScopeTable, &d.ItemKey, &d.Data)
	if err != nil {
		return AppDataRow{}, fmt.Errorf("appdata get: %w", err)
	}
	return d, nil
}

// Upsert inserts or updates a row by unique key (feature + client + table + key).
func (r *AppDataRepository) Upsert(ctx context.Context, feature, client, table, key, jsonData string) error {
	query := `MERGE INTO DB_VIEWER_APP_DATA d
		USING (SELECT :1 AS FEATURE, :2 AS SCOPE_CLIENT, :3 AS SCOPE_TABLE, :4 AS ITEM_KEY FROM DUAL) s
		ON (d.FEATURE = s.FEATURE AND d.SCOPE_CLIENT = s.SCOPE_CLIENT AND d.SCOPE_TABLE = s.SCOPE_TABLE AND d.ITEM_KEY = s.ITEM_KEY)
		WHEN MATCHED THEN UPDATE SET d.DATA = :5, d.UPDATED_AT = SYSTIMESTAMP
		WHEN NOT MATCHED THEN INSERT (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA)
			VALUES (:6, :7, :8, :9, :10)`
	_, err := r.db.ExecContext(ctx, query, feature, client, table, key, jsonData, feature, client, table, key, jsonData)
	if err != nil {
		return fmt.Errorf("appdata upsert: %w", err)
	}
	return nil
}

// Delete removes a row by feature, client, table, and key.
func (r *AppDataRepository) Delete(ctx context.Context, feature, client, table, key string) error {
	query := `DELETE FROM DB_VIEWER_APP_DATA WHERE FEATURE = :1 AND SCOPE_CLIENT = :2 AND SCOPE_TABLE = :3 AND ITEM_KEY = :4`
	res, err := r.db.ExecContext(ctx, query, feature, client, table, key)
	if err != nil {
		return fmt.Errorf("appdata delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("appdata delete: not found")
	}
	return nil
}

// DeleteByScope deletes all rows where SCOPE_CLIENT matches (for cascade delete).
func (r *AppDataRepository) DeleteByScope(ctx context.Context, client string) error {
	query := `DELETE FROM DB_VIEWER_APP_DATA WHERE SCOPE_CLIENT = :1`
	_, err := r.db.ExecContext(ctx, query, client)
	if err != nil {
		return fmt.Errorf("appdata delete by scope: %w", err)
	}
	return nil
}

func scanAppDataRows(rows *sql.Rows) ([]AppDataRow, error) {
	var result []AppDataRow
	for rows.Next() {
		var d AppDataRow
		if err := rows.Scan(&d.ID, &d.Feature, &d.ScopeClient, &d.ScopeTable, &d.ItemKey, &d.Data); err != nil {
			return nil, fmt.Errorf("appdata scan: %w", err)
		}
		result = append(result, d)
	}
	return result, rows.Err()
}
