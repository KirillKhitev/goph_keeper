// Пакет config для конфигурирования сервера и агента.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// ParamsServer - параметры сервера.
type ParamsServer struct {
	AddrRun            string `json:"addr_run"`
	DBConnectionString string `json:"database_dsn"`
	MasterKey          string `json:"master_key"`
	MigrationDB        string `json:"migration_db"`
	MigrationUser      string `json:"migration_user"`
	MigrationPassword  string `json:"migration_password"`
}

// ConfigServer хранит параметры запуска сервера.
var ConfigServer ParamsServer = ParamsServer{}

// DefaultServerConfigPath - путь к конфигу сервера.
const DefaultServerConfigPath = "server_config.json"

// Parse разбирает аргументы запуска сервера в переменнную ConfigServer.
func (f *ParamsServer) Parse() error {
	c := &ParamsServer{}
	data, err := os.ReadFile(DefaultServerConfigPath)

	if err != nil {
		fmt.Println(err)
	} else {
		err = json.Unmarshal(data, c)

		if err != nil {
			return err
		}
	}

	flag.StringVar(&f.AddrRun, "a", c.AddrRun, "address and port to run API")
	flag.StringVar(&f.DBConnectionString, "d", c.DBConnectionString, "string for connection to DB, format 'host=%s port=%s user=%s password=%s dbname=%s sslmode=%s'")
	flag.StringVar(&f.MasterKey, "mk", c.MasterKey, "master key server")
	flag.StringVar(&f.MigrationDB, "mdb", c.MigrationDB, "db for migrations")
	flag.StringVar(&f.MigrationUser, "mu", c.MigrationUser, "user for migrations")
	flag.StringVar(&f.MigrationPassword, "mp", c.MigrationPassword, "password for migrations")
	flag.Parse()

	if envRunAddr := os.Getenv(`RUN_ADDRESS`); envRunAddr != `` {
		f.AddrRun = envRunAddr
	}

	if envDBConnectionString := os.Getenv("DATABASE_URI"); envDBConnectionString != "" {
		f.DBConnectionString = envDBConnectionString
	}

	if envMasterKey := os.Getenv("MASTER_KEY"); envMasterKey != "" {
		f.MasterKey = envMasterKey
	}

	if f.AddrRun == "" {
		return fmt.Errorf("missing required address to run API")
	}

	if f.DBConnectionString == "" {
		return fmt.Errorf("missing required string for connection to DB")
	}

	if f.MasterKey == "" {
		return fmt.Errorf("missing required master key server")
	}

	return nil
}

// ParamsClient - параметры сервера.
type ParamsClient struct {
	AddrServer string `json:"addr_server"`
}

// ConfigClient хранит параметры запуска агента.
var ConfigClient ParamsClient = ParamsClient{}

// DefaultClientConfigPath - путь к конфигу агента.
const DefaultClientConfigPath = "client_config.json"

// Parse разбирает аргументы запуска агента в переменнную ConfigClient.
func (f *ParamsClient) Parse() error {
	c := &ParamsClient{}
	data, err := os.ReadFile(DefaultClientConfigPath)

	if err != nil {
		fmt.Println(err)
		c.AddrServer = "localhost:8080"
	} else {
		err = json.Unmarshal(data, c)

		if err != nil {
			return err
		}
	}

	flag.StringVar(&f.AddrServer, "a", c.AddrServer, "address and port server")
	flag.Parse()

	if envRunAddr := os.Getenv(`RUN_ADDRESS`); envRunAddr != `` {
		f.AddrServer = envRunAddr
	}

	if f.AddrServer == "" {
		return fmt.Errorf("missing required address to run API")
	}

	return nil
}
