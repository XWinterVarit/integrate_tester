package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/sijms/go-ora/v2"

	"github.com/XWinterVarit/integrate_tester/db_viewer/backend/internal/config"
	"github.com/XWinterVarit/integrate_tester/db_viewer/backend/internal/handler"
	"github.com/XWinterVarit/integrate_tester/db_viewer/backend/internal/middleware"
	"github.com/XWinterVarit/integrate_tester/db_viewer/backend/internal/model"
	"github.com/XWinterVarit/integrate_tester/db_viewer/backend/internal/repository"
	"github.com/XWinterVarit/integrate_tester/db_viewer/backend/internal/service"
	"github.com/XWinterVarit/integrate_tester/db_viewer/backend/internal/tracker"
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

	repos := make(map[string]*repository.OracleRepository)
	clientConfigs := make(map[string]model.ClientConfig)

	for _, c := range cfg.Clients {
		connStr := fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
			c.User, c.Password, c.Host, c.Port, c.Service)

		db, err := sql.Open("oracle", connStr)
		if err != nil {
			log.Printf("Warning: failed to open connection for client %s: %v", c.Name, err)
			continue
		}
		db.SetMaxOpenConns(10)
		db.SetMaxIdleConns(5)

		repos[c.Name] = repository.NewOracle(db, c.Schema)
		clientConfigs[c.Name] = c
	}

	recentFilters := tracker.New()
	recentQueries := tracker.New()

	svc := service.NewTableService(repos, clientConfigs, recentFilters, recentQueries)
	router := handler.NewRouter(svc)
	srv := middleware.CORS(cfg.Server.CORSOrigin, router)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("DB Viewer server starting on %s", addr)
	log.Fatal(http.ListenAndServe(addr, srv))
}
