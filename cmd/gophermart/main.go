package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/AlexTerra21/gophermart/internal/app/config"
	"github.com/AlexTerra21/gophermart/internal/app/handlers"
	"github.com/AlexTerra21/gophermart/internal/app/logger"
)

// go build -o cmd/gophermart/gophermart.exe cmd/gophermart/*.go
// "host=localhost user=gophermart password=gophermart dbname=gophermart sslmode=disable"
// ./cmd/gophermart/gophermart.exe -a :8081 -r :8091 -l debug -d "postgresql://gophermart:gophermart@localhost/gophermart?sslmode=disable"
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
	if err = config.InitStorage(); err != nil {
		return err
	}
	defer config.Storage.Close()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Log().Info("Running server", zap.String("address", config.GetServerAddress()))
		err := http.ListenAndServe(config.GetServerAddress(), handlers.MainRouter(config))
		if err != nil {
			log.Fatal(err)
		}
	}()
	sig := <-signalCh
	logger.Log().Sugar().Infof("Received signal: %v\n", sig)

	return nil
}
