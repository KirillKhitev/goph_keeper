// Агент для отправки/отображения приватных данных пользователя.
package main

import (
	"fmt"
	"github.com/KirillKhitev/goph_keeper/internal/agent"
	"github.com/KirillKhitev/goph_keeper/internal/config"
	tea "github.com/charmbracelet/bubbletea"
	"log"
	"os"
)

// Флаги сборки.
var (
	buildVersion string = "N/A" // Версия сборки
	buildDate    string = "N/A" // Дата сборки
	buildCommit  string = "N/A" // Комментарий сборки
)

func main() {
	printBuildInfo()
	if err := config.ConfigClient.Parse(); err != nil {
		panic(err)
	}

	if err := run(); err != nil {
		panic(err)
	}
}

// run запускает приложение агента.
func run() error {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	app, err := agent.NewAgent()
	if err != nil {
		log.Fatal(err)
	}

	p := tea.NewProgram(app, tea.WithAltScreen())

	_, err = p.Run()
	if err != nil {
		return err
	}

	return app.CatchTerminateSignal()
}

// printBuildInfo выводит в консоль информацию по сборке.
func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
