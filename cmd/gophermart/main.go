package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/AlexTerra21/gophermart/internal/app/async"
	"github.com/AlexTerra21/gophermart/internal/app/config"
	"github.com/AlexTerra21/gophermart/internal/app/handlers"
	"github.com/AlexTerra21/gophermart/internal/app/logger"
	"github.com/AlexTerra21/gophermart/internal/app/storage"
)

// go build -o cmd/gophermart/gophermart.exe cmd/gophermart/*.go
// "host=localhost user=gophermart password=gophermart dbname=gophermart sslmode=disable"
// ./cmd/gophermart/gophermart.exe -a localhost:8081 -r http://localhost:8091 -l debug -d "postgresql://gophermart:gophermart@localhost/gophermart?sslmode=disable"
func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() (err error) {
	config := config.NewConfig()
	config.ParseFlags()
	config.Print()
	if err = logger.Initialize(config.GetLogLevel()); err != nil {
		return err
	}
	if err = storage.Init(config.GetDBConnectString()); err != nil {
		return err
	}
	defer storage.GetStorage().Close()
	// сигнальный канал для завершения горутин
	doneCh := make(chan struct{})
	// закрываем его при завершении программы
	defer close(doneCh)

	async.NewAsync(doneCh, storage.GetStorage(), config.GetAccrualAddress())

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info("Running server", logger.Field{Key: "address", Val: config.GetServerAddress()})
		err := http.ListenAndServe(config.GetServerAddress(), handlers.MainRouter(config))
		if err != nil {
			log.Fatal(err)
		}
	}()
	sig := <-signalCh
	doneCh <- struct{}{}
	logger.Info(fmt.Sprintf("Received signal: %v\n", sig))

	return nil
}
