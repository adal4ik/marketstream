package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"marketstream/internal/adapters/driven/database"
	"marketstream/internal/adapters/driven/database/repository"
	redisx "marketstream/internal/adapters/driven/redis"
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
	if logger == nil {
		// Fallback на стандартный вывод, чтобы не паниковать
		logger = slog.Default()
	}
	if logFile != nil {
		defer logFile.Close()
	}

	cfg := config.Load()
	logger.Info("config loaded")

	db, err := database.ConnectDB(ctx, cfg.Database)
	if err != nil {
		logger.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("postgres connected")

	rdb, err := redisx.NewRedis(ctx, cfg.Redis)
	if err != nil {
		logger.Warn("redis connect failed", "err", err)
		rdb = nil
	} else {
		logger.Info("redis connected")
	}
	defer func() {
		if rdb != nil {
			_ = rdb.Close()
		}
	}()

	// ---- источники/пары ----
	exchangeAddrs := []string{
		"exchange1:40101",
		"exchange2:40102",
		"exchange3:40103",
	}
	exchangeNames := []string{"exchange1", "exchange2", "exchange3"}
	pairs := []string{"BTCUSDT", "DOGEUSDT", "TONUSDT", "SOLUSDT", "ETHUSDT"}

	// fan-in
	resultCh := make(chan service.Tick, 4096)

	// ---- DI ----
	repos := repository.New(db)
	mode := service.ParseMode(os.Getenv("MODE")) // "live" (default) | "test"
	svcs := service.New(repos, rdb, pairs, exchangeNames, resultCh, mode, db)

	// ---- старт начального режима ----
	if mode == service.ModeLive {
		svcs.ModeService.SwitchToLive(ctx, exchangeAddrs)
		logger.Info("mode live started")
	} else {
		svcs.ModeService.SwitchToTest(ctx, 3, 5)
		logger.Info("mode test started")
	}

	// ---- агрегатор + consumer фан-ина ----
	go svcs.Aggregator.Run(ctx)
	collector := service.NewCollector(logger)
	go collector.Run(ctx, resultCh)

	// ---- HTTP ----
	baseHandler := handlers.NewBaseHandler(logger)
	httpHandlers := handlers.New(baseHandler, svcs, db, rdb)
	mux := httpdrv.NewRouter(httpHandlers)

	httpServer := &http.Server{
		Addr:         cli.Port, // или из cfg
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("http server started", "addr", cli.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server error", "err", err)
			os.Exit(1)
		}
	}()

	// ---- graceful shutdown ----
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// аккуратно гасим источники
	if svcs.Sources != nil {
		svcs.Sources.StopAll()
	}

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown error", "err", err)
	}
	logger.Info("graceful shutdown complete")
}
