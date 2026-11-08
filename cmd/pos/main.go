package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gkamil110/pos-go/internal/bootstrap"
	"github.com/gkamil110/pos-go/internal/demo"
)

func main() {
	dataFile := flag.String("data", "data/pos.json", "файл для хранения данных")
	csvFile := flag.String("export", "results.csv", "файл для сохранения результатов")
	scoreMode := flag.String("policy", "weighted", "способ подсчета: weighted или equal")
	demoEnabled := flag.Bool("demo", false, "запустить пример работы программы")
	flag.Parse()

	app, err := bootstrap.New(
		*dataFile,
		*csvFile,
		*scoreMode,
		os.Stdin,
		os.Stdout,
		*demoEnabled,
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Не удалось запустить программу:", err)
		os.Exit(1)
	}

	if *demoEnabled {
		err = demo.Run(app.Facade, os.Stdout, *csvFile)
	} else {
		err = app.Console.Run()
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка при работе программы:", err)
		os.Exit(1)
	}
}
