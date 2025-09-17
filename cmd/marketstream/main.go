package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"marketstream/internal/adapters/driven/database"
	"marketstream/internal/adapters/driven/database/repository"
	"marketstream/internal/adapters/driver/cli"
	httpdrv "marketstream/internal/adapters/driver/http"
	"marketstream/internal/adapters/driver/http/handlers"
	"marketstream/internal/config"
	"marketstream/internal/core/service"
	"marketstream/internal/utils"

	_ "github.com/jackc/pgx/v5/stdlib"

	redisx "marketstream/internal/adapters/driven/redis"
)

func main() {
	// ---- контекст с сигналами ----
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ---- логгер / конфиг / коннекты ----
	logger, logFile := utils.Logger()
	if logger == nil {
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
	liveAddrs := []string{
		"exchange1:40101",
		"exchange2:40102",
		"exchange3:40103",
	}
	pairs := []string{"BTCUSDT", "DOGEUSDT", "TONUSDT", "SOLUSDT", "ETHUSDT"}

	// fan-in канал (общий для всех источников)
	resultCh := make(chan service.Tick, 4096)

	// ---- DI ----
	repos := repository.New(db)
	mode := service.ParseMode(os.Getenv("MODE")) // "live" (default) | "test"

	svcs := service.New(
		logger,
		repos,
		rdb,
		pairs,
		liveAddrs,
		resultCh,
		mode,
		db,
	)

	// ---- старт начального режима ----
	if mode == service.ModeLive {
		svcs.ModeService.SwitchToLive(ctx, liveAddrs)
		logger.Info("mode live started")
	} else {
		// num=3 генераторов, hz=5 тиков/сек; пары те же
		svcs.ModeService.SwitchToTest(ctx, 3, 5, pairs)
		logger.Info("mode test started")
	}
	go svcs.Aggregator.Run(ctx)

	// ---- постоянный дренаж fan-in ----
	// Всегда держим потребителя, чтобы тикеры не забивали буфер и не вешали отправителей.
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-resultCh:
				// no-op: просто дренаж; тут можно считать метрики
			}
		}
	}()

	// ---- HTTP ----
	baseHandler := handlers.NewBaseHandler(logger)
	// обновлённая сигнатура: добавлены pairs и liveAddrs
	httpHandlers := handlers.New(baseHandler, svcs, db, rdb, pairs, liveAddrs, ctx)
	mux := httpdrv.NewRouter(httpHandlers)

	httpServer := &http.Server{
		Addr:         cli.Port, // можно заменить на cfg.App.Port
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

	// аккуратно гасим источники (если поле доступно)

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown error", "err", err)
	}
	logger.Info("graceful shutdown complete")
}
