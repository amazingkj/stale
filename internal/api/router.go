package api

import (
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jiin/stale/internal/api/handler"
	apimiddleware "github.com/jiin/stale/internal/api/middleware"
	"github.com/jiin/stale/internal/repository"
	"github.com/jiin/stale/internal/service/scheduler"
	"github.com/jiin/stale/ui"
	"github.com/jmoiron/sqlx"
)

func NewRouter(
	db *sqlx.DB,
	scheduler *scheduler.Scheduler,
) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// CORS configuration
	corsConfig := apimiddleware.DefaultCORSConfig()
	r.Use(apimiddleware.CORS(corsConfig))

	// Rate limiting: 100 requests per second per client
	rateLimiter := apimiddleware.NewRateLimiter(100, time.Second)
	r.Use(rateLimiter.Handler)

	// Repositories
	sourceRepo := repository.NewSourceRepository(db)
	repoRepo := repository.NewRepoRepository(db)
	depRepo := repository.NewDependencyRepository(db)
	scanRepo := repository.NewScanRepository(db)

	// Handlers
	healthHandler := handler.NewHealthHandler()
	sourceHandler := handler.NewSourceHandler(sourceRepo)
	repoHandler := handler.NewRepoHandler(repoRepo, depRepo)
	depHandler := handler.NewDependencyHandler(depRepo)
	scanHandler := handler.NewScanHandler(scanRepo, scheduler)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(jsonContentType)

		r.Get("/health", healthHandler.Check)

		r.Route("/sources", func(r chi.Router) {
			r.Get("/", sourceHandler.List)
			r.Post("/", sourceHandler.Create)
			r.Get("/{id}", sourceHandler.Get)
			r.Put("/{id}", sourceHandler.Update)
			r.Delete("/{id}", sourceHandler.Delete)
		})

		r.Route("/repositories", func(r chi.Router) {
			r.Get("/", repoHandler.List)
			r.Get("/{id}", repoHandler.Get)
			r.Get("/{id}/dependencies", repoHandler.GetDependencies)
			r.Delete("/{id}", repoHandler.Delete)
		})

		r.Route("/dependencies", func(r chi.Router) {
			r.Get("/", depHandler.List)
			r.Get("/paginated", depHandler.ListPaginated)
			r.Get("/upgradable", depHandler.GetUpgradable)
			r.Get("/stats", depHandler.GetStats)
			r.Get("/export", depHandler.ExportCSV)
		})

		r.Route("/scans", func(r chi.Router) {
			r.Post("/", scanHandler.TriggerScan)
			r.Get("/", scanHandler.List)
			r.Get("/running", scanHandler.GetRunning)
			r.Get("/{id}", scanHandler.Get)
		})
	})

	// Serve embedded frontend
	r.Get("/*", spaHandler())

	return r
}

func spaHandler() http.HandlerFunc {
	distFS, err := fs.Sub(ui.Dist, "dist")
	if err != nil {
		return func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "frontend not built", http.StatusNotFound)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// For SPA routes (no extension), serve index.html
		if !strings.Contains(path, ".") {
			path = "/index.html"
		}

		// Read file from embedded FS
		file, err := fs.ReadFile(distFS, strings.TrimPrefix(path, "/"))
		if err != nil {
			// File not found, serve index.html for SPA
			file, _ = fs.ReadFile(distFS, "index.html")
			path = "/index.html"
		}

		// Set content type based on extension
		ext := filepath.Ext(path)
		contentType := mimeTypes[ext]
		if contentType == "" {
			contentType = "application/octet-stream"
		}
		w.Header().Set("Content-Type", contentType)
		w.Write(file)
	}
}

var mimeTypes = map[string]string{
	".html": "text/html; charset=utf-8",
	".css":  "text/css; charset=utf-8",
	".js":   "application/javascript; charset=utf-8",
	".json": "application/json",
	".svg":  "image/svg+xml",
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".ico":  "image/x-icon",
}

func jsonContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
