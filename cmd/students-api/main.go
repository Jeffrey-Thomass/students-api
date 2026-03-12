package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jeffrey-Thomass/students-api/internal/config"
	"github.com/Jeffrey-Thomass/students-api/internal/http/handlers/student"
)

func main() {

	// load config

	cfg := config.MustLoad()
	fmt.Println(cfg)

	// database setup
	// setup router

	router := http.NewServeMux()

	router.HandleFunc("GET /api/students", student.New())

	// setup server
	server := http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	slog.Info("server started", slog.String("address", cfg.Addr))
	fmt.Println("Server Started at: ", cfg.HttpServer.Addr)

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()

	<-done

	slog.Info("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // this clean up the timer

	err := server.Shutdown(ctx)
	if err != nil {
		slog.Error("failed ot shutdown server", slog.String("error", err.Error()))
	}

	slog.Info("server shut down successfully")
}
