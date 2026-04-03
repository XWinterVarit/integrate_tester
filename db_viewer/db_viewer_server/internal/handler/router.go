package handler

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/XWinterVarit/integrate_tester/db_viewer/db_viewer_server/internal/service"
)

const defaultQueryTimeout = 10 * time.Second

type statusRecorder struct {
	http.ResponseWriter
	code int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.code = code
	sr.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sr := &statusRecorder{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(sr, r)
		log.Printf("[REQUEST] %s %s %d %s", r.Method, r.URL.Path, sr.code, time.Since(start))
	})
}

func timeoutMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), defaultQueryTimeout)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewRouter(svc *service.TableService, presetSvc *service.PresetService, clientMgmtSvc *service.ClientService, lockSvc *service.LockService) http.Handler {
	mux := http.NewServeMux()

	clientH := NewClientHandler(svc)
	tableH := NewTableHandler(svc)
	queryH := NewQueryHandler(svc)
	exportH := NewExportHandler(svc)
	recentH := NewRecentHandler(svc)
	presetH := NewPresetHandler(presetSvc)
	clientMgmtH := NewClientMgmtHandler(clientMgmtSvc)
	lockH := NewLockHandler(lockSvc)

	// Client routes (existing — used by frontend for sidebar)
	mux.HandleFunc("GET /api/clients", clientH.List)
	mux.HandleFunc("GET /api/clients/{client}/tables", clientH.ListTables)

	// Client management routes
	mux.HandleFunc("GET /api/manage/clients", clientMgmtH.ListClients)
	mux.HandleFunc("POST /api/manage/clients", clientMgmtH.CreateClient)
	mux.HandleFunc("PUT /api/manage/clients/{name}", clientMgmtH.UpdateClient)
	mux.HandleFunc("DELETE /api/manage/clients/{name}", clientMgmtH.DeleteClient)
	mux.HandleFunc("POST /api/manage/clients/test-connection", clientMgmtH.TestConnection)
	mux.HandleFunc("GET /api/manage/clients/{name}/all-tables", clientMgmtH.ListAllTables)
	mux.HandleFunc("POST /api/manage/clients/list-tables", clientMgmtH.ListTablesFromConnection)

	// Lock routes
	mux.HandleFunc("POST /api/locks", lockH.Acquire)
	mux.HandleFunc("PUT /api/locks/{key}", lockH.Renew)
	mux.HandleFunc("DELETE /api/locks/{key}", lockH.Release)
	mux.HandleFunc("GET /api/locks/{key}", lockH.GetLock)

	// Table data routes
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/rows", tableH.GetRows)
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/columns", tableH.GetColumns)
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/constraints", tableH.GetConstraints)
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/indexes", tableH.GetIndexes)
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/size", tableH.GetSize)
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/count", tableH.GetRowCount)
	mux.HandleFunc("PUT /api/clients/{client}/tables/{table}/rows/update", tableH.UpdateCell)
	mux.HandleFunc("DELETE /api/clients/{client}/tables/{table}/rows/delete", tableH.DeleteRow)
	mux.HandleFunc("POST /api/clients/{client}/tables/{table}/rows/insert", tableH.InsertRow)
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/rows/delete-query", tableH.BuildDeleteQuery)
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/rows/update-query", tableH.BuildUpdateQuery)
	mux.HandleFunc("POST /api/clients/{client}/tables/{table}/rows/insert-query", tableH.BuildInsertQuery)
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/blob", tableH.DownloadBlob)
	mux.HandleFunc("POST /api/clients/{client}/tables/{table}/blob", tableH.UploadBlob)

	// Query routes
	mux.HandleFunc("POST /api/clients/{client}/tables/{table}/query", queryH.Execute)

	// Preset filter routes
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/preset-filters", presetH.ListFilters)
	mux.HandleFunc("POST /api/clients/{client}/tables/{table}/preset-filters", presetH.SaveFilter)
	mux.HandleFunc("DELETE /api/clients/{client}/tables/{table}/preset-filters/{name}", presetH.DeleteFilter)

	// Preset query routes
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/preset-queries", presetH.ListQueries)
	mux.HandleFunc("POST /api/clients/{client}/tables/{table}/preset-queries", presetH.SaveQuery)
	mux.HandleFunc("DELETE /api/clients/{client}/tables/{table}/preset-queries/{name}", presetH.DeleteQuery)
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/preset-queries/{name}/resolve", presetH.ResolveQuery)
	mux.HandleFunc("POST /api/clients/{client}/tables/{table}/validate-query", presetH.ValidateQuery)

	// Field description routes
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/field-descriptions", presetH.GetFieldDescriptions)
	mux.HandleFunc("PUT /api/clients/{client}/tables/{table}/field-descriptions", presetH.SaveFieldDescriptions)

	// Export routes
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/export", exportH.Export)

	// Recent usage routes
	mux.HandleFunc("POST /api/recent/filter", recentH.TouchFilter)
	mux.HandleFunc("POST /api/recent/query", recentH.TouchQuery)

	return loggingMiddleware(timeoutMiddleware(mux))
}
