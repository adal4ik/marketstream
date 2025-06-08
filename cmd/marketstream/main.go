package main

import (
	"context"
	"log"
	"marketstream/internal/adapters/driven/database"
	"marketstream/internal/adapters/driven/database/repository"
	"marketstream/internal/adapters/driven/redis"
	"marketstream/internal/adapters/driver/cli"
	"marketstream/internal/adapters/driver/exchange"
	"marketstream/internal/adapters/driver/web"
	"marketstream/internal/adapters/driver/web/handlers"
	"marketstream/internal/config"
	"marketstream/internal/core/service"
	"marketstream/internal/utils"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()
	db := database.ConnectDB(cfg.Database)
	defer db.Close()
	rdb := redis.NewRedis(ctx, cfg.Redis)
	logger, logFile := utils.Logger()
	defer logFile.Close()
	// // Distributors

	baseHandler := handlers.NewBaseHandler(*logger)
	repositories := repository.New(db)
	services := service.New(*repositories, rdb)
	ExchangeHandler := exchange.NewExchageHandler(services.ExchangeService)
	// Starting Listeners
	go ExchangeHandler.ListenExchange("exchange1:40101", ctx)
	go ExchangeHandler.ListenExchange("exchange2:40102", ctx)
	go ExchangeHandler.ListenExchange("exchange3:40103", ctx)

	handlers := handlers.New(*baseHandler, *services)
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	mux := web.NewRouter(*handlers)
	httpServer := &http.Server{
		Addr:    cli.Port,
		Handler: mux,
	}
	go func() {
		log.Println("Server is running on port: http://localhost" + cli.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %s", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()
}
