package main

import (
	"fmt"
	"os"

	"vyshkat/internal/app"
	"vyshkat/internal/console"
)

func main() {
	fleet, err := app.BuildFleet()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка запуска:", err)
		os.Exit(1)
	}
	if err := console.New(fleet, os.Stdin, os.Stdout).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка приложения:", err)
		os.Exit(1)
	}
}
