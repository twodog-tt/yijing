package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wangxintong/yijing/backend/internal/config"
	"github.com/wangxintong/yijing/backend/internal/db"
	"github.com/wangxintong/yijing/backend/internal/handler"
	"github.com/wangxintong/yijing/backend/internal/middleware"
	"github.com/wangxintong/yijing/backend/internal/middleware/ratelimit"
	"github.com/wangxintong/yijing/backend/internal/repository"
	"github.com/wangxintong/yijing/backend/internal/service/ai"
	"github.com/wangxintong/yijing/backend/internal/service/category"
	"github.com/wangxintong/yijing/backend/internal/service/dailyfortune"
	"github.com/wangxintong/yijing/backend/internal/service/divination"
	"github.com/wangxintong/yijing/backend/internal/service/interpretation"
	"github.com/wangxintong/yijing/backend/internal/service/sensitive"
	"github.com/wangxintong/yijing/backend/internal/service/session"
	"github.com/wangxintong/yijing/backend/internal/service/unlock"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	mysqlDB, err := db.Connect(cfg.DSN())
	if err != nil {
		log.Fatalf("connect mysql: %v", err)
	}
	defer mysqlDB.Close()
	log.Println("mysql connected")
	log.Printf("app env: %s", cfg.AppEnv)
	log.Printf("ai provider configured: %s", cfg.AIProvider)
	log.Printf("debug routes enabled: %v", cfg.EnableDebugRoutes)
	log.Printf("rate limit enabled: %v (per minute: %d)", cfg.EnableRateLimit, cfg.RateLimitPerMinute)
	log.Printf("cors allowed origins: %v", cfg.CORSAllowedOrigins)

	sessionRepo := repository.NewSessionRepository(mysqlDB)
	categoryRepo := repository.NewCategoryRepository(mysqlDB)
	hexagramRepo := repository.NewHexagramRepository(mysqlDB)
	sensitiveRepo := repository.NewSensitiveRepository(mysqlDB)
	divinationRepo := repository.NewDivinationRepository(mysqlDB)
	interpretationRepo := repository.NewInterpretationRepository(mysqlDB)
	unlockRepo := repository.NewUnlockRepository(mysqlDB)
	aiLogRepo := repository.NewAILogRepository(mysqlDB)
	dailyFortuneRepo := repository.NewDailyFortuneRepository(mysqlDB)

	aiRouter := ai.NewRouter(cfg, aiLogRepo)
	sessionSvc := session.NewService(sessionRepo)
	categorySvc := category.NewService(categoryRepo)
	sensitiveSvc := sensitive.NewService(sensitiveRepo)
	interpretationSvc := interpretation.NewService(interpretationRepo, aiRouter)
	unlockSvc := unlock.NewService(divinationRepo, sessionRepo, unlockRepo, interpretationSvc, hexagramRepo, categoryRepo)
	divinationSvc := divination.NewService(divinationRepo, hexagramRepo, categoryRepo, sessionRepo, sensitiveSvc, interpretationSvc)
	dailyFortuneSvc := dailyfortune.NewService(dailyFortuneRepo, sessionSvc, divinationSvc, interpretationSvc)

	healthHandler := &handler.HealthHandler{DB: mysqlDB}
	sessionHandler := handler.NewSessionHandler(sessionSvc)
	categoryHandler := handler.NewCategoryHandler(categorySvc)
	divinationHandler := handler.NewDivinationHandler(divinationSvc)
	interpretationHandler := handler.NewInterpretationHandler(interpretationSvc, unlockSvc)
	unlockHandler := handler.NewUnlockHandler(unlockSvc)
	dailyFortuneHandler := handler.NewDailyFortuneHandler(dailyFortuneSvc)
	debugHandler := handler.NewDebugHandler(aiLogRepo, aiRouter)

	limiter := ratelimit.NewLimiter(cfg.RateLimitPerMinute)
	rateLimit := ratelimit.Middleware(limiter, cfg.EnableRateLimit)

	mux := http.NewServeMux()
	mux.Handle("GET /health", healthHandler)
	mux.Handle("GET /api/v1/health", healthHandler)
	mux.HandleFunc("POST /api/v1/sessions", sessionHandler.Create)
	mux.HandleFunc("GET /api/v1/categories", categoryHandler.List)
	mux.HandleFunc("POST /api/v1/divinations", rateLimit(divinationHandler.Create))
	mux.HandleFunc("POST /api/v1/daily-fortune/today", dailyFortuneHandler.Today)
	mux.HandleFunc("GET /api/v1/divinations/{id}/interpretation/free", interpretationHandler.GetFree)
	mux.HandleFunc("GET /api/v1/divinations/{id}/interpretation/full", interpretationHandler.GetFull)
	mux.HandleFunc("POST /api/v1/divinations/{id}/unlock", rateLimit(unlockHandler.Unlock))
	mux.HandleFunc("GET /api/v1/divinations/{id}", divinationHandler.Get)
	mux.HandleFunc("GET /api/v1/divinations", divinationHandler.List)

	if cfg.EnableDebugRoutes {
		log.Println("registering debug routes under /api/v1/debug/*")
		mux.HandleFunc("GET /api/v1/debug/ai-logs", debugHandler.ListAILogs)
		mux.HandleFunc("GET /api/v1/debug/ai-health", debugHandler.AIHealth)
		mux.HandleFunc("GET /api/v1/debug/ai-stats", debugHandler.AIStats)
	}

	writeTimeout := 15 * time.Second
	if cfg.AIProvider == "deepseek" {
		writeTimeout = time.Duration(cfg.DeepSeekTimeoutSeconds+15) * time.Second
	}

	handlerChain := middleware.CORS(cfg.CORSAllowedOrigins)(mux)

	server := &http.Server{
		Addr:         ":" + cfg.BackendPort,
		Handler:      handlerChain,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: writeTimeout,
	}

	go func() {
		log.Printf("backend listening on http://localhost:%s", cfg.BackendPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
}
