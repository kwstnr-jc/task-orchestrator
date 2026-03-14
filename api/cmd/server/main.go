package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/kwstnr-jc/task-orchestrator/api/internal/auth"
	"github.com/kwstnr-jc/task-orchestrator/api/internal/config"
	"github.com/kwstnr-jc/task-orchestrator/api/internal/db"
	"github.com/kwstnr-jc/task-orchestrator/api/internal/handler"
	"github.com/kwstnr-jc/task-orchestrator/api/internal/middleware"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	store := db.NewStore(pool)
	authHandler := auth.NewHandler(cfg, store)
	taskHandler := handler.NewTaskHandler(store)
	projectHandler := handler.NewProjectHandler(store)
	enqueueHandler := handler.NewEnqueueHandler(store, cfg)

	r := chi.NewRouter()

	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(chimw.RequestID)
	r.Use(middleware.SecureHeaders)
	r.Use(middleware.CORS(cfg.CORSOrigin))

	// Health check (unauthenticated, for deployment verification)
	r.Get("/api/health", handler.HealthCheck)

	// Public auth routes
	r.Route("/api/auth", func(r chi.Router) {
		r.Get("/login", authHandler.LoginRedirect)
		r.Get("/callback", authHandler.Callback)
		r.Get("/logout", authHandler.Logout)
		r.With(authHandler.AuthMiddleware).Get("/me", authHandler.Me)
	})

	// Protected task routes
	// Protected project routes
	r.Route("/api/projects", func(r chi.Router) {
		r.Use(authHandler.AuthMiddleware)
		r.Get("/", projectHandler.List)
		r.Post("/", projectHandler.Create)
		r.Get("/{id}", projectHandler.Get)
		r.Put("/{id}", projectHandler.Update)
		r.Delete("/{id}", projectHandler.Delete)
	})

	r.Route("/api/tasks", func(r chi.Router) {
		r.Use(authHandler.AuthMiddleware)
		r.Get("/", taskHandler.List)
		r.Post("/", taskHandler.Create)
		r.Get("/{id}", taskHandler.Get)
		r.Put("/{id}", taskHandler.Update)
		r.Delete("/{id}", taskHandler.Delete)
		r.Post("/{id}/enqueue", enqueueHandler.Enqueue)
	})

	// SPA fallback
	spaHandler := spaFileServer("./dist")
	r.NotFound(spaHandler.ServeHTTP)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-done
	log.Println("Shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}
	log.Println("Server stopped")
}

func spaFileServer(dir string) http.Handler {
	fsys := http.Dir(dir)
	fileServer := http.FileServer(fsys)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := fsys.Open(r.URL.Path)
		if err != nil {
			// SPA fallback: serve index.html for unknown routes
			r.URL.Path = "/"
		} else {
			f.Close()
		}
		fileServer.ServeHTTP(w, r)
	})
}
