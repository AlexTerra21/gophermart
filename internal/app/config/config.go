package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/AlexTerra21/gophermart/internal/app/storage"
)

type Config struct {
	serverAddress   string
	logLevel        string
	dbConnectString string
	accrualAddress  string
	Storage         *storage.Storage
}

func NewConfig() *Config {
	conf := &Config{
		Storage: &storage.Storage{},
	}
	return conf
}

func (c *Config) InitStorage() (err error) {
	// logger.Log().Info(c.dbConnectString)
	err = c.Storage.New(c.dbConnectString)
	return
}

func (c *Config) GetServerAddress() string {
	return c.serverAddress
}

func (c *Config) GetLogLevel() string {
	return c.logLevel
}

func (c *Config) Print() {
	fmt.Printf("Server address: %s\n", c.serverAddress)
	fmt.Printf("Accrual URL: %s\n", c.accrualAddress)
	fmt.Printf("Log level: %s\n", c.logLevel)
	fmt.Printf("DB connection string: %s\n", c.dbConnectString)

}

func (c *Config) ParseFlags() {
	serverAddress := flag.String("a", ":8080", "address and port to run server")
	logLevel := flag.String("l", "info", "log level")
	dbConnectString := flag.String("d", "", "db connection string")
	accrualAddress := flag.String("r", ":8090", "accrual system address")

	flag.Parse()
	if serverAddressEnv := os.Getenv("RUN_ADDRESS"); serverAddressEnv != "" {
		serverAddress = &serverAddressEnv
	}
	if logLevelEnv := os.Getenv("LOG_LEVEL"); logLevelEnv != "" {
		logLevel = &logLevelEnv
	}
	if dbConnectStringEnv := os.Getenv("DATABASE_URI"); dbConnectStringEnv != "" {
		dbConnectString = &dbConnectStringEnv
	}
	if accrualAddressEnv := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); accrualAddressEnv != "" {
		accrualAddress = &accrualAddressEnv
	}
	c.serverAddress = *serverAddress
	c.logLevel = *logLevel
	c.dbConnectString = *dbConnectString
	c.accrualAddress = *accrualAddress
}
