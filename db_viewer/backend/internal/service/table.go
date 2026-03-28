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

	"github.com/XWinterVarit/integrate_tester/db_viewer/backend/internal/model"
	"github.com/XWinterVarit/integrate_tester/db_viewer/backend/internal/repository"
	"github.com/XWinterVarit/integrate_tester/db_viewer/backend/internal/tracker"
)

type TableService struct {
	repos         map[string]*repository.OracleRepository
	clientConfigs map[string]model.ClientConfig
	recentFilters *tracker.RecentTracker
	recentQueries *tracker.RecentTracker
}

func NewTableService(
	repos map[string]*repository.OracleRepository,
	clientConfigs map[string]model.ClientConfig,
	recentFilters *tracker.RecentTracker,
	recentQueries *tracker.RecentTracker,
) *TableService {
	return &TableService{
		repos:         repos,
		clientConfigs: clientConfigs,
		recentFilters: recentFilters,
		recentQueries: recentQueries,
	}
}

func (s *TableService) ListClients() []model.ClientInfo {
	var result []model.ClientInfo
	for _, c := range s.clientConfigs {
		result = append(result, model.ClientInfo{Name: c.Name, Schema: c.Schema})
	}
	return result
}

func (s *TableService) ListTables(client string) ([]string, error) {
	cfg, ok := s.clientConfigs[client]
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

func (s *TableService) InsertRow(ctx context.Context, client, table string, req model.InsertRowRequest) error {
	repo, err := s.getRepo(client)
	if err != nil {
		return err
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

func (s *TableService) GetFilters(client, table string) []model.PresetFilterResponse {
	tc, ok := s.getTableConfig(client, table)
	if !ok {
		return nil
	}

	var keys []string
	filterMap := make(map[string]model.PresetFilterResponse)
	for _, f := range tc.PresetFilters {
		key := client + ":" + table + ":" + f.Name
		keys = append(keys, key)
		filterMap[key] = model.PresetFilterResponse{
			Name:    f.Name,
			Details: f.Details,
			Columns: f.Columns,
		}
	}

	sorted := s.recentFilters.SortByRecent(keys)
	var result []model.PresetFilterResponse
	for _, key := range sorted {
		result = append(result, filterMap[key])
	}
	return result
}

func (s *TableService) GetPresetQueries(client, table string) []model.PresetQueryResponse {
	tc, ok := s.getTableConfig(client, table)
	if !ok {
		return nil
	}

	cfg := s.clientConfigs[client]
	var result []model.PresetQueryResponse
	for i, q := range tc.PresetQueries {
		resolved := strings.ReplaceAll(q.Query, "{THIS_TABLE}", cfg.Schema+"."+table)
		var args []model.PresetQueryArgResponse
		for _, a := range q.Arguments {
			args = append(args, model.PresetQueryArgResponse{
				Name:        a.Name,
				Type:        a.Type,
				Description: a.Description,
			})
		}
		result = append(result, model.PresetQueryResponse{
			Index:     i,
			Name:      q.Name,
			Query:     resolved,
			Arguments: args,
		})
	}
	return result
}

func (s *TableService) ResolvePresetQuery(client, table string, index int) (string, error) {
	tc, ok := s.getTableConfig(client, table)
	if !ok {
		return "", fmt.Errorf("table not found")
	}
	if index < 0 || index >= len(tc.PresetQueries) {
		return "", fmt.Errorf("query index out of range")
	}

	cfg := s.clientConfigs[client]
	preset := tc.PresetQueries[index]
	resolved := strings.ReplaceAll(preset.Query, "{THIS_TABLE}", cfg.Schema+"."+table)

	s.recentQueries.Touch(client + ":" + table + ":" + preset.Name)
	return resolved, nil
}

func (s *TableService) ExportTable(ctx context.Context, w io.Writer, client, table, exportType, format string, params model.RowQueryParams) error {
	repo, err := s.getRepo(client)
	if err != nil {
		return err
	}

	cfg := s.clientConfigs[client]
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
	repo, ok := s.repos[client]
	if !ok {
		return nil, fmt.Errorf("client not found: %s", client)
	}
	return repo, nil
}

func (s *TableService) getTableConfig(client, table string) (model.TableConfig, bool) {
	cfg, ok := s.clientConfigs[client]
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
