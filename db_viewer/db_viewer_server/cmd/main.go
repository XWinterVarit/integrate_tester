package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/sijms/go-ora/v2"

	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/config"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/handler"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/middleware"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/repository"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/service"
	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/tracker"
)

func main() {
	cfgPath := "db_viewer/sql_test/config.yml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Use the first client from config.yml as the "admin" connection for reading CLIENT_CONFIG rows.
	// All client configs are stored in DB_VIEWER_APP_DATA on this admin DB.
	if len(cfg.Clients) == 0 {
		log.Fatal("At least one client must be defined in config.yml for the admin connection")
	}
	adminClient := cfg.Clients[0]
	adminConnStr := fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
		adminClient.User, adminClient.Password, adminClient.Host, adminClient.Port, adminClient.Service)
	adminDB, err := sql.Open("oracle", adminConnStr)
	if err != nil {
		log.Fatalf("Failed to open admin connection: %v", err)
	}
	adminDB.SetMaxOpenConns(10)
	adminDB.SetMaxIdleConns(5)

	adminRepo := repository.NewAppDataRepository(adminDB)

	// Ensure the app data table exists on the admin DB
	if err := adminRepo.EnsureTable(context.Background()); err != nil {
		log.Printf("Warning: failed to ensure app data table: %v", err)
	}

	// Create connection pool and services
	pool := service.NewConnectionPool()
	clientSvc := service.NewClientService(pool, adminRepo)
	lockSvc := service.NewLockService(adminRepo)

	// Load all client configs from DB and open connections
	if err := clientSvc.LoadClientsFromDB(context.Background()); err != nil {
		log.Printf("Warning: failed to load clients from DB: %v", err)
	}

	recentFilters := tracker.New()
	recentQueries := tracker.New()

	tableSvc := service.NewTableService(pool, clientSvc, recentFilters, recentQueries)
	presetSvc := service.NewPresetService(pool, clientSvc)

	// Ensure app data tables on all client connections
	if err := presetSvc.EnsureTables(context.Background()); err != nil {
		log.Printf("Warning: failed to ensure app data tables: %v", err)
	}

	router := handler.NewRouter(tableSvc, presetSvc, clientSvc, lockSvc)
	srv := middleware.CORS(cfg.Server.CORSOrigin, router)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("DB Viewer server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, srv))
}
