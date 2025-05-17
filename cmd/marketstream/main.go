package main

import (
	"context"
	"os"
	"os/signal"

	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	// var wg sync.WaitGroup
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	// 	<-ctx.Done()
	// 	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// 	defer cancel()
	// 	if err := httpServer.Shutdown(shutdownCtx); err != nil {
	// 		fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
	// 	}
	// }()
	// wg.Wait()
}
