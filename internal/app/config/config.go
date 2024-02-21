package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	serverAddress   string
	logLevel        string
	dbConnectString string
	accrualAddress  string
}

func NewConfig() *Config {
	conf := &Config{}
	return conf
}

func (c *Config) GetServerAddress() string {
	return c.serverAddress
}

func (c *Config) GetAccrualAddress() string {
	return c.accrualAddress
}

func (c *Config) GetDBConnectString() string {
	return c.dbConnectString
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

// Для тестов
func (c *Config) SetServerAddress(addr string) {
	c.serverAddress = addr
}
func (c *Config) SetAccrualAddress(addr string) {
	c.accrualAddress = addr
}
func (c *Config) SetDBConnectionString(dbString string) {
	c.dbConnectString = dbString
}
func (c *Config) SetLogLevel(level string) {
	c.logLevel = level
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
