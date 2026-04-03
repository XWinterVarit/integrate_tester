package service

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/sijms/go-ora/v2"

	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/repository"
)

// ConnectionPool manages dynamic DB connections keyed by client name.
type ConnectionPool struct {
	mu       sync.RWMutex
	dbs      map[string]*sql.DB
	repos    map[string]*repository.OracleRepository
	appRepos map[string]*repository.AppDataRepository
}

func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		dbs:      make(map[string]*sql.DB),
		repos:    make(map[string]*repository.OracleRepository),
		appRepos: make(map[string]*repository.AppDataRepository),
	}
}

// Add opens a connection and registers it in the pool.
func (p *ConnectionPool) Add(name, host string, port int, serviceName, username, password, schema string) error {
	connStr := fmt.Sprintf("oracle://%s:%s@%s:%d/%s", username, password, host, port, serviceName)
	db, err := sql.Open("oracle", connStr)
	if err != nil {
		return fmt.Errorf("open connection for %s: %w", name, err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	p.mu.Lock()
	defer p.mu.Unlock()

	// Close existing connection if any
	if old, ok := p.dbs[name]; ok {
		old.Close()
	}

	p.dbs[name] = db
	p.repos[name] = repository.NewOracle(db, schema)
	p.appRepos[name] = repository.NewAppDataRepository(db)
	return nil
}

// Remove closes and removes a connection from the pool.
func (p *ConnectionPool) Remove(name string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if db, ok := p.dbs[name]; ok {
		db.Close()
		delete(p.dbs, name)
		delete(p.repos, name)
		delete(p.appRepos, name)
	}
}

// GetRepo returns the OracleRepository for a client.
func (p *ConnectionPool) GetRepo(name string) (*repository.OracleRepository, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	r, ok := p.repos[name]
	return r, ok
}

// GetAppDataRepo returns the AppDataRepository for a client.
func (p *ConnectionPool) GetAppDataRepo(name string) (*repository.AppDataRepository, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	r, ok := p.appRepos[name]
	return r, ok
}

// GetDB returns the raw *sql.DB for a client.
func (p *ConnectionPool) GetDB(name string) (*sql.DB, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	db, ok := p.dbs[name]
	return db, ok
}

// AllRepos returns a snapshot of all OracleRepositories.
func (p *ConnectionPool) AllRepos() map[string]*repository.OracleRepository {
	p.mu.RLock()
	defer p.mu.RUnlock()
	m := make(map[string]*repository.OracleRepository, len(p.repos))
	for k, v := range p.repos {
		m[k] = v
	}
	return m
}

// AllAppDataRepos returns a snapshot of all AppDataRepositories.
func (p *ConnectionPool) AllAppDataRepos() map[string]*repository.AppDataRepository {
	p.mu.RLock()
	defer p.mu.RUnlock()
	m := make(map[string]*repository.AppDataRepository, len(p.appRepos))
	for k, v := range p.appRepos {
		m[k] = v
	}
	return m
}

// ClientNames returns all registered client names.
func (p *ConnectionPool) ClientNames() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	names := make([]string, 0, len(p.dbs))
	for k := range p.dbs {
		names = append(names, k)
	}
	return names
}
