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
	"github.com/jiin/stale/internal/service/email"
	"github.com/jiin/stale/internal/service/scheduler"
	"github.com/jiin/stale/ui"
	"github.com/jmoiron/sqlx"
)

func NewRouter(
	db *sqlx.DB,
	scheduler *scheduler.Scheduler,
	emailService *email.Service,
) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))
	r.Use(apimiddleware.SecurityHeaders())
	r.Use(apimiddleware.AuditLog())

	// CORS configuration
	corsConfig := apimiddleware.DefaultCORSConfig()
	r.Use(apimiddleware.CORS(corsConfig))

	// Authentication
	authConfig := apimiddleware.DefaultAuthConfig()
	r.Use(apimiddleware.Auth(authConfig))

	// Rate limiting: 100 requests per second per client
	rateLimiter := apimiddleware.NewRateLimiter(100, time.Second)
	r.Use(rateLimiter.Handler)

	// Repositories
	sourceRepo := repository.NewSourceRepository(db)
	repoRepo := repository.NewRepoRepository(db)
	depRepo := repository.NewDependencyRepository(db)
	scanRepo := repository.NewScanRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	ignoredRepo := repository.NewIgnoredRepository(db)

	// Handlers
	healthHandler := handler.NewHealthHandler(db)
	sourceHandler := handler.NewSourceHandler(sourceRepo, repoRepo, depRepo)
	repoHandler := handler.NewRepoHandler(repoRepo, depRepo)
	depHandler := handler.NewDependencyHandler(depRepo)
	scanHandler := handler.NewScanHandler(scanRepo, scheduler)
	settingsHandler := handler.NewSettingsHandler(settingsRepo, scheduler, emailService)
	ignoredHandler := handler.NewIgnoredHandler(ignoredRepo)

	// Register cache invalidation callback for scan completion
	scheduler.OnScanComplete(depHandler.ClearCache)

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
			r.Post("/bulk-delete", repoHandler.BulkDelete)
			r.Get("/{id}", repoHandler.Get)
			r.Get("/{id}/dependencies", repoHandler.GetDependencies)
			r.Delete("/{id}", repoHandler.Delete)
		})

		r.Route("/dependencies", func(r chi.Router) {
			r.Get("/", depHandler.List)
			r.Get("/paginated", depHandler.ListPaginated)
			r.Get("/upgradable", depHandler.GetUpgradable)
			r.Get("/stats", depHandler.GetStats)
			r.Get("/repos", depHandler.GetRepositoryNames)
			r.Get("/packages", depHandler.GetPackageNames)
			r.Get("/filter-options", depHandler.GetFilterOptions)
			r.Get("/export", depHandler.ExportCSV)
		})

		r.Route("/scans", func(r chi.Router) {
			r.Post("/", scanHandler.TriggerScan)
			r.Get("/", scanHandler.List)
			r.Get("/running", scanHandler.GetRunning)
			r.Get("/{id}", scanHandler.Get)
			r.Post("/{id}/cancel", scanHandler.Cancel)
		})

		r.Route("/settings", func(r chi.Router) {
			r.Get("/", settingsHandler.Get)
			r.Put("/", settingsHandler.Update)
			r.Post("/test-email", settingsHandler.TestEmail)
			r.Get("/next-scan", settingsHandler.GetNextScan)
		})

		r.Route("/ignored", func(r chi.Router) {
			r.Get("/", ignoredHandler.List)
			r.Post("/", ignoredHandler.Create)
			r.Post("/bulk", ignoredHandler.BulkCreate)
			r.Post("/bulk-delete", ignoredHandler.BulkDelete)
			r.Delete("/{id}", ignoredHandler.Delete)
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

		// Set cache headers based on file type
		switch ext {
		case ".js", ".css":
			// Immutable assets with hash in filename - cache for 1 year
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		case ".html":
			// HTML should be revalidated
			w.Header().Set("Cache-Control", "public, max-age=0, must-revalidate")
		case ".svg", ".png", ".jpg", ".ico":
			// Images - cache for 1 week
			w.Header().Set("Cache-Control", "public, max-age=604800")
		}

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
