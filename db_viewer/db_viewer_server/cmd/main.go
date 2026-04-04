package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	_ "github.com/sijms/go-ora/v2"

	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/config"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/handler"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/logger"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/middleware"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/repository"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/service"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/tracker"
)

func main() {
	log := logger.New("STARTUP")

	cfgPath := "db_viewer/db_viewer_server/config.yml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	log.Info("Loading config from: %s", cfgPath)
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatal("Failed to load config: %v", err)
	}
	log.Info("Config loaded — server port: %d, CORS origin: %s", cfg.Server.Port, cfg.Server.CORSOrigin)

	// Use data_store from config.yml as the server-side connection for storing app data (DB_VIEWER_APP_DATA).
	// This connection is not exposed to the web UI.
	ds := cfg.DataStore
	adminConnStr := fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
		ds.User, ds.Password, ds.Host, ds.Port, ds.Service)

	log.Info("Opening admin DB connection — host: %s, port: %d, service: %s, user: %s",
		ds.Host, ds.Port, ds.Service, ds.User)
	adminDB, err := sql.Open("oracle", adminConnStr)
	if err != nil {
		log.Fatal("Failed to open admin DB: %v", err)
	}
	adminDB.SetMaxOpenConns(10)
	adminDB.SetMaxIdleConns(5)

	ctx := context.Background()

	// Verify connectivity to the main DB — this is required to run the server.
	log.Info("Pinging admin DB to verify connectivity...")
	if err := adminDB.PingContext(ctx); err != nil {
		log.Fatal("Cannot connect to admin DB (%s:%d/%s): %v — server cannot start without main DB",
			ds.Host, ds.Port, ds.Service, err)
	}
	log.Info("Admin DB connection established successfully")

	adminRepo := repository.NewAppDataRepository(adminDB)

	// Ensure the app data table exists on the admin DB — required for server operation.
	log.Info("Ensuring app data table (DB_VIEWER_APP_DATA) exists on admin DB...")
	if err := adminRepo.EnsureTable(ctx); err != nil {
		log.Fatal("Failed to ensure app data table on admin DB: %v — server cannot start without app data table", err)
	}
	log.Info("App data table is ready")

	// Create connection pool and services
	pool := service.NewConnectionPool()
	clientSvc := service.NewClientService(pool, adminRepo)
	lockSvc := service.NewLockService(adminRepo)

	// Load all client configs from DB and open connections
	log.Info("Loading client configurations from DB...")
	if err := clientSvc.LoadClientsFromDB(ctx); err != nil {
		log.Warn("Failed to load clients from DB: %v", err)
	} else {
		log.Info("Client configurations loaded — active connections: %d", len(pool.ClientNames()))
	}

	recentFilters := tracker.New()
	recentQueries := tracker.New()

	tableSvc := service.NewTableService(pool, clientSvc, recentFilters, recentQueries)
	presetSvc := service.NewPresetService(pool, clientSvc)

	// Ensure app data tables on all client connections (non-fatal — clients may be added later)
	log.Info("Ensuring app data tables on all client connections...")
	if err := presetSvc.EnsureTables(ctx); err != nil {
		log.Warn("Failed to ensure app data tables on one or more client connections: %v", err)
	} else {
		log.Info("App data tables verified on all client connections")
	}

	router := handler.NewRouter(tableSvc, presetSvc, clientSvc, lockSvc)
	srv := middleware.CORS(cfg.Server.CORSOrigin, router)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Info("DB Viewer server is ready — listening on %s", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatal("Server stopped unexpectedly: %v", err)
	}
}
