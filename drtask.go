package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
)



func main() {
	// Debugging
	// - set DEBUG (`export DEBUG=1`) and `tail -f messages.log` in another terminal
	var dump *os.File
	if _, ok := os.LookupEnv("DEBUG"); ok {
		var err error
		dump, err = os.OpenFile("messages.log", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			os.Exit(1)
		}
	}

	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}

	app := &cli.App{
		Name:  "drtask",
		Usage: "A simple terminal based task organiser with AI support.",
		Action: func(*cli.Context) error {

			initialModel := initialModel()
			initialModel.dump = dump
			p := tea.NewProgram(&initialModel)

			if _, err := p.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
