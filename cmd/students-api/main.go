package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Jeffrey-Thomass/students-api/internal/config"
	"github.com/Jeffrey-Thomass/students-api/internal/http/handlers/student"
	"github.com/Jeffrey-Thomass/students-api/internal/storage/sqlite"
)

func main() {

	// load config

	cfg := config.MustLoad()
	fmt.Println(cfg)

	// database setup

	storage, err := sqlite.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("storage initialized", slog.String("env", cfg.Env), slog.String("version", "1.0.0"))

	// setup router

	router := http.NewServeMux()

	router.HandleFunc("POST /api/students", student.New(storage))
	router.HandleFunc("GET /api/students/{id}", student.GetById(storage))
	router.HandleFunc("GET /api/students", student.GetList(storage))

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

	err2 := server.Shutdown(ctx)
	if err2 != nil {
		slog.Error("failed ot shutdown server", slog.String("error", err2.Error()))
	}

	slog.Info("server shut down successfully")
}
