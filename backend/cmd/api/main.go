package main

import (
	"context"
	"errors"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nunutech40/affilatorshopee/internal/config"
	"github.com/nunutech40/affilatorshopee/internal/db"
	"github.com/nunutech40/affilatorshopee/internal/handler"
	"github.com/nunutech40/affilatorshopee/internal/repository"
	"github.com/nunutech40/affilatorshopee/internal/service"
	"github.com/nunutech40/affilatorshopee/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "backend/internal/db/migrations"
	}
	if !filepath.IsAbs(migrationsPath) {
		migrationsPath, _ = filepath.Abs(migrationsPath)
	}
	if err := db.RunMigrations(cfg.DatabaseURL, migrationsPath); err != nil {
		log.Fatal(err)
	}

	productRepo := repository.NewProductRepository(database)
	nicheRepo := repository.NewNicheRepository(database)
	clickRepo := repository.NewClickRepository(database)
	commissionRepo := repository.NewCommissionRepository(database)
	postLogRepo := repository.NewPostLogRepository(database)
	variationRepo := repository.NewCaptionVariationRepository(database)
	mediaRepo := repository.NewMediaRepository(database)
	contentRepo := repository.NewContentRepository(database)
	productService := service.NewProductService(productRepo)
	shareService := service.NewShareService()
	captionService := service.NewCaptionService(shareService)
	variationService := service.NewCaptionVariationService(variationRepo)
	postLogService := service.NewPostLogService(postLogRepo)
	aiService := service.NewAIService(cfg.AIAPIKey, cfg.OpenRouterModel, cfg.AIBaseURL)
	aiService.ConfigureProviders(cfg.NineRouterAPIKey, cfg.NineRouterBaseURL, cfg.OpenCodeAPIKey)
	aiService.ConfigureCodexBridge(cfg.CodexBridgeURL, cfg.CodexBridgeToken)
	localStorage, err := storage.NewLocalStorage(cfg.StoragePath)
	if err != nil {
		log.Fatal(err)
	}
	mediaService := service.NewMediaService(localStorage, mediaRepo)
	xService := service.NewXService(productRepo, mediaService)

	r := chi.NewRouter()
	r.Use(handler.CORS(cfg.CORSAllowedOrigins), handler.RequestLimit)
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	products := handler.NewProductHandler(productService, mediaService)
	niches := handler.NewNicheHandler(nicheRepo)
	content := handler.NewContentHandler(contentRepo)
	clicks := handler.NewClickHandler(clickRepo)
	commissions := handler.NewCommissionHandler(commissionRepo)
	analytics := handler.NewAnalyticsHandler(database, clickRepo, commissionRepo, productRepo)
	ai := handler.NewAIHandler(productService, aiService)
	captions := handler.NewCaptionHandler(productService, captionService, variationService)
	postLogs := handler.NewPostLogHandler(postLogService)
	share := handler.NewShareHandler(shareService)
	media := handler.NewMediaHandler(mediaService)
	xHandler := handler.NewXHandler(xService)

	r.Route("/api", func(api chi.Router) {
		api.Get("/ai/models", ai.Models)
		api.Get("/niches", niches.List)
		api.Post("/niches", niches.Create)
		api.Delete("/niches/{id}", niches.Delete)
		api.Get("/content-niches", content.ListNiches)
		api.Post("/content-niches", content.CreateNiche)
		api.Delete("/content-niches/{id}", content.DeleteNiche)
		api.Get("/content-items", content.List)
		api.Post("/content-items", content.Create)
		api.Get("/products", products.List)
		api.Post("/products", products.Create)
		api.Put("/products/{id}/niches", niches.ReplaceProduct)
		api.Post("/analytics/clicks/import", clicks.ImportCSV)
		api.Post("/analytics/commissions/import", commissions.ImportCSV)
		api.Get("/analytics/commissions/sold", commissions.ListSold)
		api.Get("/analytics/commissions/events", commissions.ListEvents)
		api.Get("/analytics/commissions/summary", commissions.GetSummary)
		api.Post("/analytics/reset", analytics.Reset)
		api.Post("/products/import/x", xHandler.Import)
		api.Post("/ai/reformat", ai.Reformat)
		api.Post("/captions/generate", captions.Generate)
		api.Post("/captions/variations", captions.GenerateVariations)
		api.Get("/share/x", share.X)
		api.Post("/post-logs", postLogs.Create)
		api.Get("/post-logs", postLogs.ListAll)
		api.Get("/post-logs/{productID}", postLogs.List)
		api.Get("/products/{id}", products.Get)
		api.Patch("/products/{id}", products.Patch)
		api.Delete("/products/{id}", products.Delete)
		api.Get("/products/{id}/caption-variations", captions.ListVariations)
		api.Patch("/caption-variations/{variationID}", captions.PatchVariation)
		api.Delete("/caption-variations/{variationID}", captions.DeleteVariation)
		api.Get("/products/{id}/media", media.List)
		api.Post("/products/{id}/media", media.Add)
		api.Get("/products/{id}/media/download", media.Download)
		api.Get("/products/{id}/media/{mediaID}/file", media.File)
		api.Delete("/products/{id}/media/{mediaID}", media.Delete)
	})

	registerStatic(r)
	server := &http.Server{Addr: ":" + cfg.Port, Handler: r, ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 4 * time.Minute, IdleTimeout: 60 * time.Second}
	go func() {
		log.Printf("server listening on :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}

func registerStatic(r chi.Router) {
	dist, _ := filepath.Abs("frontend/dist")
	r.Get("/*", func(w http.ResponseWriter, req *http.Request) {
		relative := strings.TrimPrefix(filepath.Clean("/"+req.URL.Path), "/")
		path := filepath.Join(dist, filepath.FromSlash(relative))
		if relative == "" || !strings.HasPrefix(path, dist+string(os.PathSeparator)) || !fileExists(path) {
			path = filepath.Join(dist, "index.html")
		}
		file, err := os.Open(path)
		if err != nil {
			http.NotFound(w, req)
			return
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil || info.IsDir() {
			http.NotFound(w, req)
			return
		}
		if contentType := mime.TypeByExtension(filepath.Ext(path)); contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		http.ServeContent(w, req, info.Name(), info.ModTime(), file)
	})
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
