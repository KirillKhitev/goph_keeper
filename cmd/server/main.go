package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/KirillKhitev/goph_keeper/internal/config"
	"github.com/KirillKhitev/goph_keeper/internal/store/pg"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Флаги сборки.
var (
	buildVersion string = "N/A" // Версия сборки
	buildDate    string = "N/A" // Дата сборки
	buildCommit  string = "N/A" // Комментарий сборки
)

func main() {
	printBuildInfo()
	if err := config.ConfigServer.Parse(); err != nil {
		panic(err)
	}

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	conn, err := sql.Open("pgx", config.ConfigServer.DBConnectionString)
	if err != nil {
		return err
	}

	ctx := context.Background()

	store, err := pg.NewStore(ctx, conn)
	if err != nil {
		return err
	}

	appInstance := newApp(store)

	go appInstance.StartServer()

	return appInstance.CatchTerminateSignal()
}

// printBuildInfo выводит в консоль информацию по сборке.
func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
