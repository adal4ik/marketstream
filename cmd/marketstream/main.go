package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"marketstream/internal/adapters/driven/database"
	"marketstream/internal/adapters/driven/database/repository"
	"marketstream/internal/adapters/driven/redis"
	"marketstream/internal/adapters/driver/cli"
	httpdrv "marketstream/internal/adapters/driver/http"
	"marketstream/internal/adapters/driver/http/handlers"
	"marketstream/internal/config"
	"marketstream/internal/core/service"
	"marketstream/internal/utils"
)

func main() {
	// ---- контекст с сигналами ----
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ---- логгер / конфиг / коннекты ----
	logger, logFile := utils.Logger()
	defer logFile.Close()

	cfg := config.Load()
	logger.Info("config loaded")

	db := database.ConnectDB(cfg.Database)
	defer db.Close()
	logger.Info("postgres connected")

	rdb := redis.NewRedis(ctx, cfg.Redis)
	defer rdb.Close()
	logger.Info("redis connected")

	// ---- источники/пары ----
	exchangeAddrs := []string{
		"exchange1:40101",
		"exchange2:40102",
		"exchange3:40103",
	}
	exchangeNames := []string{"exchange1", "exchange2", "exchange3"}
	pairs := []string{"BTCUSDT", "DOGEUSDT", "TONUSDT", "SOLUSDT", "ETHUSDT"}

	// fan-in (если нужно считать метрики — можно подписаться)
	resultCh := make(chan service.Tick, 4096)

	// ---- DI ----
	repos := repository.New(db)
	mode := service.ParseMode(os.Getenv("MODE")) // "live" (default) | "test"

	// New(repo, rdb, pairs, exchanges, resultCh, initialMode)
	svcs := service.New(repos, rdb, pairs, exchangeNames, resultCh, mode, db)

	// ---- старт начального режима ----
	if mode == service.ModeLive {
		svcs.ModeService.SwitchToLive(ctx, exchangeAddrs)
		logger.Info("mode live started")
	} else {
		svcs.ModeService.SwitchToTest(ctx, 3, 5) // 3 тест-«биржи», 5 tps/пара
		logger.Info("mode test started")
	}

	// ---- агрегатор ----
	go svcs.Aggregator.Run(ctx)

	// ---- HTTP ----
	baseHandler := handlers.NewBaseHandler(logger)
	httpHandlers := handlers.New(baseHandler, svcs, db, rdb) // у тебя уже так
	mux := httpdrv.NewRouter(httpHandlers)

	httpServer := &http.Server{
		Addr:         cli.Port, // или порт из cfg
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("HTTP server: http://localhost" + cli.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	// ---- graceful shutdown ----
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// аккуратно гасим источники (live/test)
	if svcs.Sources != nil {
		svcs.Sources.StopAll()
	}

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}

	log.Println("graceful shutdown complete")
}
