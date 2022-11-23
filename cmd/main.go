package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/bysoft-wallet/users/internal/app"
	"github.com/bysoft-wallet/users/internal/ports"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()
	app, err := app.NewApplication(ctx)
	if err != nil {
		log.Printf("Application creation error %s", err)
		os.Exit(1)
	}

	accessHeader := os.Getenv("ACCESS_TOKEN_HEADER")
	f, err := os.OpenFile("app-log.json", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Failed to create logfile" + "app-log.json")
		panic(err)
	}
	defer f.Close()

	log := &logrus.Logger{
		// Log into f file handler and on os.Stdout
		Out:   io.MultiWriter(f, os.Stdout),
		Level: logrus.DebugLevel,
		Formatter: &logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		},
	}
	
	server := ports.NewHttpServer(app, accessHeader, log)
	server.Start()
}
