package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/model"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/repository"
)

func (s *PresetService) getAppRepo(client string) (*repository.AppDataRepository, error) {
	repo, ok := s.pool.GetAppDataRepo(client)
	if !ok {
		return nil, fmt.Errorf("client not found: %s", client)
	}
	return repo, nil
}

func (s *PresetService) getSchema(client string) string {
	configs := s.clientSvc.GetClientConfigs(context.Background())
	if cfg, ok := configs[client]; ok {
		return cfg.Schema
	}
	return client
}

const (
	featurePresetFilter = "PRESET_FILTER"
	featurePresetQuery  = "PRESET_QUERY"
	featureFieldDesc    = "FIELD_DESC"
)

// PresetService handles preset filter and query CRUD via the app data table.
type PresetService struct {
	pool      *ConnectionPool
	clientSvc *ClientService
}

func NewPresetService(
	pool *ConnectionPool,
	clientSvc *ClientService,
) *PresetService {
	return &PresetService{
		pool:      pool,
		clientSvc: clientSvc,
	}
}

// EnsureTables creates the app data table for each client connection.
func (s *PresetService) EnsureTables(ctx context.Context) error {
	for name, repo := range s.pool.AllAppDataRepos() {
		if err := repo.EnsureTable(ctx); err != nil {
			return fmt.Errorf("ensure app data table for %s: %w", name, err)
		}
	}
	return nil
}

// --- Preset Filters ---

func (s *PresetService) ListPresetFilters(ctx context.Context, client, table string) ([]model.PresetFilterResponse, error) {
	repo, err := s.getAppRepo(client)
	if err != nil {
		return nil, err
	}
	rows, err := repo.List(ctx, featurePresetFilter, client, table)
	if err != nil {
		return nil, err
	}
	var result []model.PresetFilterResponse
	for _, row := range rows {
		var f model.PresetFilterResponse
		if err := json.Unmarshal([]byte(row.Data), &f); err != nil {
			continue
		}
		result = append(result, f)
	}
	return result, nil
}

func (s *PresetService) SavePresetFilter(ctx context.Context, client, table string, req model.SavePresetFilterRequest) error {
	repo, err := s.getAppRepo(client)
	if err != nil {
		return err
	}
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(req.Columns) == 0 {
		return fmt.Errorf("at least one column is required")
	}
	data, err := json.Marshal(model.PresetFilterResponse{
		Name:    req.Name,
		Details: req.Details,
		Columns: req.Columns,
	})
	if err != nil {
		return err
	}
	return repo.Upsert(ctx, featurePresetFilter, client, table, req.Name, string(data))
}

func (s *PresetService) DeletePresetFilter(ctx context.Context, client, table, name string) error {
	repo, err := s.getAppRepo(client)
	if err != nil {
		return err
	}
	return repo.Delete(ctx, featurePresetFilter, client, table, name)
}

// --- Preset Queries ---

func (s *PresetService) ListPresetQueries(ctx context.Context, client, table string) ([]model.PresetQueryResponse, error) {
	repo, err := s.getAppRepo(client)
	if err != nil {
		return nil, err
	}
	rows, err := repo.List(ctx, featurePresetQuery, client, table)
	if err != nil {
		return nil, err
	}
	schema := s.getSchema(client)
	var result []model.PresetQueryResponse
	for _, row := range rows {
		var q model.PresetQueryResponse
		if err := json.Unmarshal([]byte(row.Data), &q); err != nil {
			continue
		}
		q.Query = strings.ReplaceAll(q.Query, "{THIS_TABLE}", schema+"."+table)
		result = append(result, q)
	}
	return result, nil
}

func (s *PresetService) SavePresetQuery(ctx context.Context, client, table string, req model.SavePresetQueryRequest) error {
	repo, err := s.getAppRepo(client)
	if err != nil {
		return err
	}
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Query == "" {
		return fmt.Errorf("query is required")
	}
	// Validate query syntax and argument references before persisting
	if validation := s.ValidateQuery(ctx, client, table, model.ValidateQueryRequest{
		Query:     req.Query,
		Arguments: req.Arguments,
	}); !validation.Valid {
		return fmt.Errorf("invalid query: %s", validation.Error)
	}
	resp := model.PresetQueryResponse{
		Name:  req.Name,
		Query: req.Query,
	}
	for _, a := range req.Arguments {
		resp.Arguments = append(resp.Arguments, model.PresetQueryArgResponse{
			Name:        a.Name,
			Type:        a.Type,
			Description: a.Description,
		})
	}
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return repo.Upsert(ctx, featurePresetQuery, client, table, req.Name, string(data))
}

func (s *PresetService) DeletePresetQuery(ctx context.Context, client, table, name string) error {
	repo, err := s.getAppRepo(client)
	if err != nil {
		return err
	}
	return repo.Delete(ctx, featurePresetQuery, client, table, name)
}

// --- Query Validation ---

var argPattern = regexp.MustCompile(`:([A-Za-z_][A-Za-z0-9_]*)`)

// ValidateQuery checks argument references AND validates SQL syntax against the live DB.
func (s *PresetService) ValidateQuery(ctx context.Context, client, table string, req model.ValidateQueryRequest) model.ValidateQueryResponse {
	if strings.TrimSpace(req.Query) == "" {
		return model.ValidateQueryResponse{Valid: false, Error: "query is empty"}
	}

	// Step 1: check for undefined :ARGUMENT references
	matches := argPattern.FindAllStringSubmatch(req.Query, -1)
	definedArgs := make(map[string]bool)
	argTypeMap := make(map[string]string)
	for _, a := range req.Arguments {
		upper := strings.ToUpper(a.Name)
		definedArgs[upper] = true
		argTypeMap[upper] = a.Type
	}

	var undefined []string
	seen := make(map[string]bool)
	for _, m := range matches {
		name := strings.ToUpper(m[1])
		if !definedArgs[name] && !seen[name] {
			undefined = append(undefined, m[1])
			seen[name] = true
		}
	}
	if len(undefined) > 0 {
		return model.ValidateQueryResponse{
			Valid:         false,
			Error:         fmt.Sprintf("undefined arguments: %s", strings.Join(undefined, ", ")),
			UndefinedArgs: undefined,
		}
	}

	// Step 2: validate SQL syntax by doing a zero-row dry-run against the real DB.
	// Wrapping in SELECT * FROM (...) WHERE 1=0 forces Oracle to fully parse the
	// inner query without returning any rows or mutating data.
	db, ok := s.pool.GetDB(client)
	if !ok {
		return model.ValidateQueryResponse{Valid: false, Error: "client not found"}
	}

	schema := s.getSchema(client)
	sqlToCheck := strings.ReplaceAll(req.Query, "{THIS_TABLE}", schema+"."+table)
	dryRun := "SELECT * FROM (" + sqlToCheck + ") WHERE 1=0"

	// Build dummy bind args for every :ARGNAME so Oracle accepts the statement.
	var sqlArgs []any
	seenBind := make(map[string]bool)
	for _, m := range argPattern.FindAllStringSubmatch(sqlToCheck, -1) {
		name := m[1]
		upper := strings.ToUpper(name)
		if seenBind[upper] {
			continue
		}
		seenBind[upper] = true
		switch argTypeMap[upper] {
		case "number":
			sqlArgs = append(sqlArgs, sql.Named(name, 0))
		default:
			sqlArgs = append(sqlArgs, sql.Named(name, ""))
		}
	}

	rows, err := db.QueryContext(ctx, dryRun, sqlArgs...)
	if err != nil {
		return model.ValidateQueryResponse{Valid: false, Error: err.Error()}
	}
	rows.Close()

	return model.ValidateQueryResponse{Valid: true}
}

// ResolvePresetQuery resolves a preset query by name, replacing {THIS_TABLE} and returning the resolved SQL.
func (s *PresetService) ResolvePresetQuery(ctx context.Context, client, table, name string) (model.PresetQueryResponse, error) {
	repo, err := s.getAppRepo(client)
	if err != nil {
		return model.PresetQueryResponse{}, err
	}
	row, err := repo.Get(ctx, featurePresetQuery, client, table, name)
	if err != nil {
		return model.PresetQueryResponse{}, fmt.Errorf("preset query not found: %s", name)
	}
	var q model.PresetQueryResponse
	if err := json.Unmarshal([]byte(row.Data), &q); err != nil {
		return model.PresetQueryResponse{}, err
	}
	schema := s.getSchema(client)
	q.Query = strings.ReplaceAll(q.Query, "{THIS_TABLE}", schema+"."+table)
	return q, nil
}

// --- Field Descriptions ---

func (s *PresetService) GetFieldDescriptions(ctx context.Context, client, table string) (map[string]string, error) {
	repo, err := s.getAppRepo(client)
	if err != nil {
		return nil, err
	}
	row, err := repo.Get(ctx, featureFieldDesc, client, table, "descriptions")
	if err != nil {
		// No descriptions yet — return empty map
		return map[string]string{}, nil
	}
	var descs map[string]string
	if err := json.Unmarshal([]byte(row.Data), &descs); err != nil {
		return map[string]string{}, nil
	}
	return descs, nil
}

func (s *PresetService) SaveFieldDescriptions(ctx context.Context, client, table string, descs map[string]string) error {
	repo, err := s.getAppRepo(client)
	if err != nil {
		return err
	}
	data, err := json.Marshal(descs)
	if err != nil {
		return err
	}
	return repo.Upsert(ctx, featureFieldDesc, client, table, "descriptions", string(data))
}
