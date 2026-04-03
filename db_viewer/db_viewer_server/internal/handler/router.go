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

func NewRouter(svc *service.TableService) http.Handler {
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
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/filters", queryH.GetFilters)
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/preset-queries", queryH.GetPresetQueries)
	mux.HandleFunc("POST /api/clients/{client}/tables/{table}/preset-queries/{index}/resolve", queryH.ResolvePresetQuery)

	// Export routes
	mux.HandleFunc("GET /api/clients/{client}/tables/{table}/export", exportH.Export)

	// Recent usage routes
	mux.HandleFunc("POST /api/recent/filter", recentH.TouchFilter)
	mux.HandleFunc("POST /api/recent/query", recentH.TouchQuery)

	return loggingMiddleware(timeoutMiddleware(mux))
}
