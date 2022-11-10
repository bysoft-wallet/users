package main

import (
	"context"
	"log"
	"os"

	"github.com/bysoft-wallet/users/internal/app"
	"github.com/bysoft-wallet/users/internal/ports"
)

func main() {
	ctx := context.Background()
	app, err := app.NewApplication(ctx)
	if err != nil {
		log.Printf("Application creation error %s", err)
		os.Exit(1)
	}

	accessHeader := os.Getenv("ACCESS_TOKEN_HEADER")
	
	server := ports.NewHttpServer(app, accessHeader)
	server.Start()
}
