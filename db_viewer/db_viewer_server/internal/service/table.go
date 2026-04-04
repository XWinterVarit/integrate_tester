package service

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/model"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/repository"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/tracker"
)

type TableService struct {
	pool          *ConnectionPool
	clientSvc     *ClientService
	recentFilters *tracker.RecentTracker
	recentQueries *tracker.RecentTracker
}

func NewTableService(
	pool *ConnectionPool,
	clientSvc *ClientService,
	recentFilters *tracker.RecentTracker,
	recentQueries *tracker.RecentTracker,
) *TableService {
	return &TableService{
		pool:          pool,
		clientSvc:     clientSvc,
		recentFilters: recentFilters,
		recentQueries: recentQueries,
	}
}

func (s *TableService) ListClients() []model.ClientInfo {
	configs := s.clientSvc.GetClientConfigs(context.Background())
	result := make([]model.ClientInfo, 0)
	for _, c := range configs {
		result = append(result, model.ClientInfo{Name: c.Name, DisplayName: c.Name, Schema: c.Schema})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result
}

func (s *TableService) ListTables(client string) ([]string, error) {
	configs := s.clientSvc.GetClientConfigs(context.Background())
	cfg, ok := configs[client]
	if !ok {
		return nil, fmt.Errorf("client not found: %s", client)
	}
	var tables []string
	for _, t := range cfg.Tables {
		tables = append(tables, t.Name)
	}
	return tables, nil
}

func (s *TableService) GetRows(ctx context.Context, client, table string, params model.RowQueryParams) ([]map[string]any, error) {
	repo, err := s.getRepo(client)
	if err != nil {
		return nil, err
	}
	return repo.QueryRows(ctx, params, table)
}

func (s *TableService) ExecuteQuery(ctx context.Context, client string, req model.ExecuteQueryRequest) ([]map[string]any, error) {
	repo, err := s.getRepo(client)
	if err != nil {
		return nil, err
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}

	query := req.Query

	// Inject ROWID so results are editable — wrap as subquery
	upper := strings.TrimSpace(strings.ToUpper(query))
	if strings.HasPrefix(upper, "SELECT") && !strings.Contains(upper, "ROWID") {
		query = "SELECT q.ROWID, q.* FROM (" + query + ") q"
	}

	if req.Sort != "" {
		dir := "ASC"
		if strings.ToLower(req.SortDir) == "desc" {
			dir = "DESC"
		}
		query = "SELECT * FROM (" + query + ") ORDER BY " + req.Sort + " " + dir
	}

	if !strings.Contains(strings.ToUpper(query), "FETCH") {
		if req.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d ROWS", req.Offset)
		}
		query += fmt.Sprintf(" FETCH NEXT %d ROWS ONLY", limit)
	}

	var args []any
	for k, v := range req.Args {
		args = append(args, sql.Named(k, v))
	}

	return repo.ExecuteRawQuery(ctx, query, args)
}

func (s *TableService) GetColumns(ctx context.Context, client, table string) ([]map[string]any, error) {
	repo, err := s.getRepo(client)
	if err != nil {
		return nil, err
	}
	return repo.GetColumns(ctx, table)
}

func (s *TableService) GetConstraints(ctx context.Context, client, table string) ([]map[string]any, error) {
	repo, err := s.getRepo(client)
	if err != nil {
		return nil, err
	}
	return repo.GetConstraints(ctx, table)
}

func (s *TableService) GetIndexes(ctx context.Context, client, table string) ([]map[string]any, error) {
	repo, err := s.getRepo(client)
	if err != nil {
		return nil, err
	}
	return repo.GetIndexes(ctx, table)
}

func (s *TableService) GetRowCount(ctx context.Context, client, table string) (int, error) {
	repo, err := s.getRepo(client)
	if err != nil {
		return 0, err
	}
	return repo.GetRowCount(ctx, table)
}

func (s *TableService) GetTableSize(ctx context.Context, client, table string) ([]map[string]any, error) {
	repo, err := s.getRepo(client)
	if err != nil {
		return nil, err
	}
	return repo.GetTableSize(ctx, table)
}

func (s *TableService) UpdateCell(ctx context.Context, client, table string, req model.UpdateCellRequest) error {
	repo, err := s.getRepo(client)
	if err != nil {
		return err
	}
	return repo.UpdateCell(ctx, table, req.Column, req.Value, req.Rowid)
}

func (s *TableService) DeleteRow(ctx context.Context, client, table string, req model.DeleteRowRequest) error {
	repo, err := s.getRepo(client)
	if err != nil {
		return err
	}
	return repo.DeleteRow(ctx, table, req.Rowid)
}

func (s *TableService) InsertRow(ctx context.Context, client, table string, req model.InsertRowRequest) (string, error) {
	repo, err := s.getRepo(client)
	if err != nil {
		return "", err
	}
	return repo.InsertRow(ctx, table, req.Columns, req.Values)
}

func (s *TableService) BuildDeleteQuery(client, table, rowid string) (string, error) {
	repo, err := s.getRepo(client)
	if err != nil {
		return "", err
	}
	return repo.BuildDeleteQuery(table, rowid), nil
}

func (s *TableService) BuildUpdateQuery(client, table, column, value, rowid string) (string, error) {
	repo, err := s.getRepo(client)
	if err != nil {
		return "", err
	}
	return repo.BuildUpdateQuery(table, column, value, rowid), nil
}

func (s *TableService) BuildInsertQuery(client, table string, columns, values []string) (string, error) {
	repo, err := s.getRepo(client)
	if err != nil {
		return "", err
	}
	return repo.BuildInsertQuery(table, columns, values), nil
}

func (s *TableService) GetBlobData(ctx context.Context, client, table, column, rowid string) ([]byte, error) {
	repo, err := s.getRepo(client)
	if err != nil {
		return nil, err
	}
	return repo.GetBlobData(ctx, table, column, rowid)
}

func (s *TableService) UploadBlobData(ctx context.Context, client, table, column, rowid string, data []byte) error {
	repo, err := s.getRepo(client)
	if err != nil {
		return err
	}
	return repo.UploadBlobData(ctx, table, column, rowid, data)
}

func (s *TableService) ExportTable(ctx context.Context, w io.Writer, client, table, exportType, format string, params model.RowQueryParams) error {
	repo, err := s.getRepo(client)
	if err != nil {
		return err
	}

	configs := s.clientSvc.GetClientConfigs(context.Background())
	cfg := configs[client]
	var query string
	if exportType == "full" {
		query = fmt.Sprintf("SELECT * FROM %s.%s", cfg.Schema, table)
	} else {
		cols := "*"
		if params.Select != "" {
			cols = params.Select
		}
		query = fmt.Sprintf("SELECT %s FROM %s.%s", cols, cfg.Schema, table)
		if params.Sort != "" {
			dir := "ASC"
			if params.SortDir == "desc" {
				dir = "DESC"
			}
			query += fmt.Sprintf(" ORDER BY %s %s", params.Sort, dir)
		}
		if params.Limit > 0 {
			query += fmt.Sprintf(" FETCH FIRST %d ROWS ONLY", params.Limit)
		}
	}

	rows, err := repo.ExportRows(ctx, query)
	if err != nil {
		return err
	}

	if format == "json" {
		return json.NewEncoder(w).Encode(rows)
	}

	csvW := csv.NewWriter(w)
	if len(rows) > 0 {
		var headers []string
		for k := range rows[0] {
			headers = append(headers, k)
		}
		sort.Strings(headers)
		csvW.Write(headers)
		for _, row := range rows {
			var vals []string
			for _, h := range headers {
				vals = append(vals, fmt.Sprintf("%v", row[h]))
			}
			csvW.Write(vals)
		}
	}
	csvW.Flush()
	return csvW.Error()
}

func (s *TableService) TouchRecentFilter(key string) {
	s.recentFilters.Touch(key)
}

func (s *TableService) TouchRecentQuery(key string) {
	s.recentQueries.Touch(key)
}

func (s *TableService) getRepo(client string) (*repository.OracleRepository, error) {
	repo, ok := s.pool.GetRepo(client)
	if !ok {
		return nil, fmt.Errorf("client not found: %s", client)
	}
	return repo, nil
}

func (s *TableService) getTableConfig(client, table string) (model.TableConfig, bool) {
	configs := s.clientSvc.GetClientConfigs(context.Background())
	cfg, ok := configs[client]
	if !ok {
		return model.TableConfig{}, false
	}
	for _, t := range cfg.Tables {
		if strings.EqualFold(t.Name, table) {
			return t, true
		}
	}
	return model.TableConfig{}, false
}
