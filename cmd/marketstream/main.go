package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	_ "github.com/lib/pq"

	"marketstream/internal/adapters/driven/database"
	"marketstream/internal/adapters/driven/database/repository"
	"marketstream/internal/adapters/driven/exchange"
	"marketstream/internal/adapters/driven/redis"
	"marketstream/internal/adapters/driver/cli"
	"marketstream/internal/adapters/driver/web"
	"marketstream/internal/adapters/driver/web/handlers"
	"marketstream/internal/core/service"
	"marketstream/internal/utils"
)

func main() {
	ctx := context.Background()
	db := database.ConnectDB()
	defer db.Close()
	rdb := redis.NewRedis(ctx)
	defer rdb.Close()
	logger, logFile := utils.Logger()
	defer logFile.Close()
	exchange.ListenExchange("127.0.0.1:40101")
	baseHandler := handlers.NewBaseHandler(*logger)
	repositories := repository.New(db)
	services := service.New()
	_ = services
	_ = repositories
	handlers := handlers.New(*baseHandler)
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
