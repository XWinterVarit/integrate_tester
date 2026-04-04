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

type appDataRowScan struct {
	ID          int64
	Feature     string
	ScopeClient sql.NullString
	ScopeTable  sql.NullString
	ItemKey     sql.NullString
	Data        sql.NullString
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

// ListByFeature returns all rows matching only the feature (ignores scope columns).
func (r *AppDataRepository) ListByFeature(ctx context.Context, feature string) ([]AppDataRow, error) {
	query := `SELECT ID, FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA
		FROM DB_VIEWER_APP_DATA
		WHERE FEATURE = :1
		ORDER BY ID`
	rows, err := r.db.QueryContext(ctx, query, feature)
	if err != nil {
		return nil, fmt.Errorf("appdata list by feature: %w", err)
	}
	defer rows.Close()
	return scanAppDataRows(rows)
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

// GetByFeatureAndKey returns a single row matching only feature and item key (scope columns are NULL).
func (r *AppDataRepository) GetByFeatureAndKey(ctx context.Context, feature, key string) (AppDataRow, error) {
	query := `SELECT ID, FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA
		FROM DB_VIEWER_APP_DATA
		WHERE FEATURE = :1 AND ITEM_KEY = :2 AND SCOPE_CLIENT IS NULL AND SCOPE_TABLE IS NULL`
	row := r.db.QueryRowContext(ctx, query, feature, key)
	var s appDataRowScan
	err := row.Scan(&s.ID, &s.Feature, &s.ScopeClient, &s.ScopeTable, &s.ItemKey, &s.Data)
	if err != nil {
		return AppDataRow{}, fmt.Errorf("appdata get by feature+key: %w", err)
	}
	return AppDataRow{ID: s.ID, Feature: s.Feature, ScopeClient: s.ScopeClient.String, ScopeTable: s.ScopeTable.String, ItemKey: s.ItemKey.String, Data: s.Data.String}, nil
}

// Get returns a single row by feature, client, table, and key.
func (r *AppDataRepository) Get(ctx context.Context, feature, client, table, key string) (AppDataRow, error) {
	query := `SELECT ID, FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA
		FROM DB_VIEWER_APP_DATA
		WHERE FEATURE = :1 AND SCOPE_CLIENT = :2 AND SCOPE_TABLE = :3 AND ITEM_KEY = :4`
	row := r.db.QueryRowContext(ctx, query, feature, client, table, key)
	var s appDataRowScan
	err := row.Scan(&s.ID, &s.Feature, &s.ScopeClient, &s.ScopeTable, &s.ItemKey, &s.Data)
	if err != nil {
		return AppDataRow{}, fmt.Errorf("appdata get: %w", err)
	}
	return AppDataRow{ID: s.ID, Feature: s.Feature, ScopeClient: s.ScopeClient.String, ScopeTable: s.ScopeTable.String, ItemKey: s.ItemKey.String, Data: s.Data.String}, nil
}

// UpsertByFeatureAndKey inserts or updates a row where scope columns are NULL (used for CLIENT_CONFIG).
func (r *AppDataRepository) UpsertByFeatureAndKey(ctx context.Context, feature, key, jsonData string) error {
	query := `MERGE INTO DB_VIEWER_APP_DATA d
		USING (SELECT :1 AS FEATURE, :2 AS ITEM_KEY FROM DUAL) s
		ON (d.FEATURE = s.FEATURE AND d.ITEM_KEY = s.ITEM_KEY AND d.SCOPE_CLIENT IS NULL AND d.SCOPE_TABLE IS NULL)
		WHEN MATCHED THEN UPDATE SET d.DATA = :3, d.UPDATED_AT = SYSTIMESTAMP
		WHEN NOT MATCHED THEN INSERT (FEATURE, SCOPE_CLIENT, SCOPE_TABLE, ITEM_KEY, DATA)
			VALUES (:4, NULL, NULL, :5, :6)`
	_, err := r.db.ExecContext(ctx, query, feature, key, jsonData, feature, key, jsonData)
	if err != nil {
		return fmt.Errorf("appdata upsert by feature+key: %w", err)
	}
	return nil
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

// DeleteByFeatureAndKey deletes a row matching only feature and item key (scope columns are NULL).
func (r *AppDataRepository) DeleteByFeatureAndKey(ctx context.Context, feature, key string) error {
	query := `DELETE FROM DB_VIEWER_APP_DATA WHERE FEATURE = :1 AND ITEM_KEY = :2 AND SCOPE_CLIENT IS NULL AND SCOPE_TABLE IS NULL`
	res, err := r.db.ExecContext(ctx, query, feature, key)
	if err != nil {
		return fmt.Errorf("appdata delete by feature+key: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("appdata delete by feature+key: not found")
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
		var s appDataRowScan
		if err := rows.Scan(&s.ID, &s.Feature, &s.ScopeClient, &s.ScopeTable, &s.ItemKey, &s.Data); err != nil {
			return nil, fmt.Errorf("appdata scan: %w", err)
		}
		result = append(result, AppDataRow{
			ID:          s.ID,
			Feature:     s.Feature,
			ScopeClient: s.ScopeClient.String,
			ScopeTable:  s.ScopeTable.String,
			ItemKey:     s.ItemKey.String,
			Data:        s.Data.String,
		})
	}
	return result, rows.Err()
}
