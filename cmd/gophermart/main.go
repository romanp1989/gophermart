package main

import (
	"github.com/romanp1989/gophermart/cli/migrate"
	app2 "github.com/romanp1989/gophermart/internal/app"
	"github.com/romanp1989/gophermart/internal/config"
	"log"
	"os"
)

func main() {
	application, err := app2.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	err = InitDB()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(application.Run())
}

func InitDB() error {
	mCmd, err := migrate.NewMigrateCmd(&migrate.Config{Dsn: config.Options.FlagDBDsn})
	if err != nil {
		return err
	}

	if err = mCmd.Up(); err != nil {
		return err
	}

	return nil
}
