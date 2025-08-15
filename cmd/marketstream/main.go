package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"marketstream/internal/adapters/driven/database"
	"marketstream/internal/adapters/driven/database/repository"
	"marketstream/internal/adapters/driven/exchange"
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

	// ---- источники и пары (жёстко) ----
	exchangeAddrs := []string{
		"exchange1:40101",
		"exchange2:40102",
		"exchange3:40103",
	}
	exchangeNames := []string{"exchange1", "exchange2", "exchange3"} // имена без портов — пригодятся агрегатору
	pairs := []string{"BTCUSDT", "DOGEUSDT", "TONUSDT", "SOLUSDT", "ETHUSDT"}

	// ---- DI (Service-комбайн) ----
	// ВАЖНО: сигнатура service.New должна быть New(repo, rdb, pairs, exchanges)
	// если у тебя пока New(repo, rdb, pairs) — добавь 4-й параметр в конструкторе Service.
	repos := repository.New(db)
	services := service.New(repos, rdb, pairs, exchangeNames)

	baseHandler := handlers.NewBaseHandler(logger)
	httpHandlers := handlers.New(baseHandler, services)

	// ---- HTTP ----
	mux := httpdrv.NewRouter(httpHandlers)
	httpServer := &http.Server{
		Addr:         cli.Port, // либо cfg-порт
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

	// ---- fan-out / fan-in ----
	resultCh := make(chan service.Tick, 4096) // общий канал (если пригодится дальше под метрики)

	var wg sync.WaitGroup

	for i, addr := range exchangeAddrs {
		exName := exchangeNames[i]
		rawCh := make(chan []byte, 1024)

		// listener с backoff-reconnect
		wg.Add(1)
		go func(name, a string, out chan<- []byte) {
			defer wg.Done()
			backoffs := []time.Duration{1 * time.Second, 5 * time.Second, 15 * time.Second, 30 * time.Second}
			attempt := 0
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				// блокирующий вызов — читает TCP и пишет строки в out; возвращается при разрыве
				exchange.Listen(ctx, a, out)

				// пауза перед реконнектом
				wait := backoffs[attempt]
				if attempt < len(backoffs)-1 {
					attempt++
				}
				select {
				case <-ctx.Done():
					return
				case <-time.After(wait):
				}
				_ = name // для логирования при желании
			}
		}(exName, addr, rawCh)

		// 5 воркеров на источник
		for w := 0; w < 5; w++ {
			wg.Add(1)
			go func(name string, in <-chan []byte) {
				defer wg.Done()
				services.Exchange.Worker(ctx, name, in, resultCh)
			}(exName, rawCh)
		}
	}

	// ---- Aggregator (каждую секунду собирает окна 60s, раз в минуту пишет батч в Postgres) ----
	go services.Aggregator.Run(ctx)

	// ---- потребитель fan-in (опционально — счётчики/лог) ----
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case t := <-resultCh:
				_ = t // здесь можно считать метрики, если надо
			}
		}
	}()

	// ---- graceful shutdown ----
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}

	wg.Wait()
	log.Println("graceful shutdown complete")
}
