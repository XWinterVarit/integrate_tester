package handler

import (
	"net/http"

	"github.com/XWinterVarit/integrate_tester/db_viewer/backend/internal/service"
)

func NewRouter(svc *service.TableService) *http.ServeMux {
	mux := http.NewServeMux()

	clientH := NewClientHandler(svc)
	tableH := NewTableHandler(svc)
	queryH := NewQueryHandler(svc)
	exportH := NewExportHandler(svc)
	recentH := NewRecentHandler(svc)

	// Client routes
	mux.HandleFunc("GET /api/clients", clientH.List)
	mux.HandleFunc("GET /api/clients/{client}/tables", clientH.ListTables)

	// Table data routes
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/rows", tableH.GetRows)
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/columns", tableH.GetColumns)
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/constraints", tableH.GetConstraints)
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/indexes", tableH.GetIndexes)
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/size", tableH.GetSize)
	mux.HandleFunc("PUT /api/clients/{client}/tables/{table}/rows/update", tableH.UpdateCell)
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/blob", tableH.DownloadBlob)

	// Query routes
	mux.HandleFunc("POST /api/clients/{client}/tables/{table}/query", queryH.Execute)
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/filters", queryH.GetFilters)
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/preset-queries", queryH.GetPresetQueries)
	mux.HandleFunc("POST /api/clients/{client}/tables/{table}/preset-queries/{index}/resolve", queryH.ResolvePresetQuery)

	// Export routes
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/export", exportH.Export)

	// Recent usage routes
	mux.HandleFunc("POST /api/recent/filter", recentH.TouchFilter)
	mux.HandleFunc("POST /api/recent/query", recentH.TouchQuery)

	return mux
}
