package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sort"

	_ "github.com/sijms/go-ora/v2"

	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/model"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/repository"
)

const featureClientConfig = "CLIENT_CONFIG"

// clientConfigJSON is the JSON stored in DATA column for CLIENT_CONFIG rows.
type clientConfigJSON struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Host        string   `json:"host"`
	Port        int      `json:"port"`
	ServiceName string   `json:"service_name"`
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	Tables      []string `json:"tables"`
}

// ClientService manages client configurations stored in DB_VIEWER_APP_DATA.
type ClientService struct {
	pool *ConnectionPool
	// adminRepo is one AppDataRepository used for reading/writing CLIENT_CONFIG rows.
	// We pick the first available connection's repo (or a dedicated admin connection).
	adminRepo *repository.AppDataRepository
}

func NewClientService(pool *ConnectionPool, adminRepo *repository.AppDataRepository) *ClientService {
	return &ClientService{pool: pool, adminRepo: adminRepo}
}

func (s *ClientService) ListClients(ctx context.Context) ([]model.ClientConfigResponse, error) {
	rows, err := s.adminRepo.List(ctx, featureClientConfig, "", "")
	if err != nil {
		return nil, err
	}
	var result []model.ClientConfigResponse
	for _, row := range rows {
		var cfg clientConfigJSON
		if err := json.Unmarshal([]byte(row.Data), &cfg); err != nil {
			log.Printf("Warning: invalid CLIENT_CONFIG JSON for key=%s: %v", row.ItemKey, err)
			continue
		}
		result = append(result, model.ClientConfigResponse{
			Name:        cfg.Name,
			DisplayName: cfg.DisplayName,
			Host:        cfg.Host,
			Port:        cfg.Port,
			ServiceName: cfg.ServiceName,
			Username:    cfg.Username,
			Tables:      cfg.Tables,
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

func (s *ClientService) GetClient(ctx context.Context, name string) (model.ClientConfigResponse, error) {
	row, err := s.adminRepo.Get(ctx, featureClientConfig, "", "", name)
	if err != nil {
		return model.ClientConfigResponse{}, fmt.Errorf("client not found: %s", name)
	}
	var cfg clientConfigJSON
	if err := json.Unmarshal([]byte(row.Data), &cfg); err != nil {
		return model.ClientConfigResponse{}, fmt.Errorf("invalid config for %s: %w", name, err)
	}
	return model.ClientConfigResponse{
		Name:        cfg.Name,
		DisplayName: cfg.DisplayName,
		Host:        cfg.Host,
		Port:        cfg.Port,
		ServiceName: cfg.ServiceName,
		Username:    cfg.Username,
		Tables:      cfg.Tables,
	}, nil
}

func (s *ClientService) SaveClient(ctx context.Context, req model.SaveClientRequest, isCreate bool) error {
	if req.Name == "" {
		return fmt.Errorf("client name is required")
	}

	// Check duplicate on create
	if isCreate {
		_, err := s.adminRepo.Get(ctx, featureClientConfig, "", "", req.Name)
		if err == nil {
			return fmt.Errorf("CONFLICT:client '%s' already exists", req.Name)
		}
	}

	// Build JSON — for update, if password is empty, keep existing password
	password := req.Password
	if !isCreate && password == "" {
		existing, err := s.adminRepo.Get(ctx, featureClientConfig, "", "", req.Name)
		if err == nil {
			var old clientConfigJSON
			if json.Unmarshal([]byte(existing.Data), &old) == nil {
				password = old.Password
			}
		}
	}

	cfg := clientConfigJSON{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Host:        req.Host,
		Port:        req.Port,
		ServiceName: req.ServiceName,
		Username:    req.Username,
		Password:    password,
		Tables:      req.Tables,
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := s.adminRepo.Upsert(ctx, featureClientConfig, "", "", req.Name, string(data)); err != nil {
		return err
	}

	// Update connection pool
	schema := req.Username
	if err := s.pool.Add(req.Name, req.Host, req.Port, req.ServiceName, req.Username, password, schema); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	return nil
}

func (s *ClientService) DeleteClient(ctx context.Context, name string) error {
	// Delete all related data (presets, field descs, locks) for this client
	if err := s.adminRepo.DeleteByScope(ctx, name); err != nil {
		log.Printf("Warning: failed to delete scoped data for client %s: %v", name, err)
	}

	// Delete the CLIENT_CONFIG row itself
	if err := s.adminRepo.Delete(ctx, featureClientConfig, "", "", name); err != nil {
		return fmt.Errorf("delete client config: %w", err)
	}

	// Remove from connection pool
	s.pool.Remove(name)
	return nil
}

func (s *ClientService) TestConnection(_ context.Context, req model.TestConnectionRequest) error {
	connStr := fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
		req.Username, req.Password, req.Host, req.Port, req.ServiceName)
	db, err := sql.Open("oracle", connStr)
	if err != nil {
		return fmt.Errorf("failed to open: %w", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	return nil
}

// ListAllTables queries USER_TABLES from the given client's connection.
func (s *ClientService) ListAllTables(ctx context.Context, clientName string) ([]string, error) {
	db, ok := s.pool.GetDB(clientName)
	if !ok {
		return nil, fmt.Errorf("client not found: %s", clientName)
	}
	rows, err := db.QueryContext(ctx, "SELECT TABLE_NAME FROM USER_TABLES ORDER BY TABLE_NAME")
	if err != nil {
		return nil, fmt.Errorf("query tables: %w", err)
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, rows.Err()
}

// ListAllTablesFromConnection queries USER_TABLES using provided connection info (without saving).
func (s *ClientService) ListAllTablesFromConnection(ctx context.Context, req model.TestConnectionRequest) ([]string, error) {
	connStr := fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
		req.Username, req.Password, req.Host, req.Port, req.ServiceName)
	db, err := sql.Open("oracle", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open: %w", err)
	}
	defer db.Close()
	rows, err := db.QueryContext(ctx, "SELECT TABLE_NAME FROM USER_TABLES ORDER BY TABLE_NAME")
	if err != nil {
		return nil, fmt.Errorf("query tables: %w", err)
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tables = append(tables, name)
	}
	return tables, rows.Err()
}

// LoadClientsFromDB reads all CLIENT_CONFIG rows and opens connections in the pool.
func (s *ClientService) LoadClientsFromDB(ctx context.Context) error {
	rows, err := s.adminRepo.List(ctx, featureClientConfig, "", "")
	if err != nil {
		return err
	}
	for _, row := range rows {
		var cfg clientConfigJSON
		if err := json.Unmarshal([]byte(row.Data), &cfg); err != nil {
			log.Printf("Warning: skipping invalid CLIENT_CONFIG key=%s: %v", row.ItemKey, err)
			continue
		}
		schema := cfg.Username
		if err := s.pool.Add(cfg.Name, cfg.Host, cfg.Port, cfg.ServiceName, cfg.Username, cfg.Password, schema); err != nil {
			log.Printf("Warning: failed to connect client %s: %v", cfg.Name, err)
		}
	}
	return nil
}

// GetClientConfigs returns a map of client name → model.ClientConfig for use by TableService/PresetService.
func (s *ClientService) GetClientConfigs(ctx context.Context) map[string]model.ClientConfig {
	rows, err := s.adminRepo.List(ctx, featureClientConfig, "", "")
	if err != nil {
		return nil
	}
	configs := make(map[string]model.ClientConfig)
	for _, row := range rows {
		var cfg clientConfigJSON
		if err := json.Unmarshal([]byte(row.Data), &cfg); err != nil {
			continue
		}
		var tables []model.TableConfig
		for _, t := range cfg.Tables {
			tables = append(tables, model.TableConfig{Name: t})
		}
		configs[cfg.Name] = model.ClientConfig{
			Name:    cfg.Name,
			User:    cfg.Username,
			Host:    cfg.Host,
			Port:    cfg.Port,
			Service: cfg.ServiceName,
			Schema:  cfg.Username,
			Tables:  tables,
		}
	}
	return configs
}
