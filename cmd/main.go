package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/bysoft-wallet/users/internal/app"
	"github.com/bysoft-wallet/users/internal/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()

	//init logger
	f, err := os.OpenFile("logs/app-log.json", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Failed to create logfile " + "app-log.json")
		panic(err)
	}
	defer f.Close()

	logger := &logrus.Logger{
		// Log into f file handler and on os.Stdout
		Out:   io.MultiWriter(f, os.Stdout),
		Level: logrus.DebugLevel,
		Formatter: &logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		},
	}

	// init db connection pool
	pool, err := pgxpool.New(ctx, os.Getenv("POSTGRES_URL"))
	if err != nil {
		logger.Errorf("pgxpool init error %v", err)
		os.Exit(1)
	}

	queryLog, err := strconv.ParseBool(os.Getenv("ENABLE_QUERY_LOG"))
	if err != nil {
		queryLog = false
	}

	if queryLog {
		addPoolTracer(pool, logger)
	}

	//init env variables
	JWTSecret := os.Getenv("JWT_SECRET")
	JWTAccessTTL, err := strconv.Atoi(os.Getenv("JWT_ACCESS_TTL"))
	JWTRefreshTTL, err := strconv.Atoi(os.Getenv("JWT_REFRESH_TTL"))
	if JWTSecret == "" || err != nil {
		logger.Errorf("JWT configuration must be provided %v", err)
		os.Exit(1)
	}

	maxSessions, err := strconv.Atoi(os.Getenv("MAX_USER_SESSIONS"))
	if err != nil {
		logger.Errorf("Max user sessions configuration must be provided %v", err)
		os.Exit(1)
	}

	accessHeader := os.Getenv("ACCESS_TOKEN_HEADER")
	if accessHeader == "" {
		logger.Errorf("ACCESS_TOKEN_HEADER must be provided %v", err)
		os.Exit(1)
	}

	//init application
	appConfig := app.Config{
		Ctx:             ctx,
		Logger:          logger,
		DbPool:          pool,
		JwtSecret:       JWTSecret,
		JwtAccessTTL:    JWTAccessTTL,
		JwtRefreshTTL:   JWTRefreshTTL,
		MaxUserSessions: maxSessions,
	}

	app, err := app.NewApplication(&appConfig)
	if err != nil {
		logger.Errorf("Application init error %v", err)
		os.Exit(1)
	}

	server := ports.NewHttpServer(app, accessHeader)
	server.Start()
}

type QueryTracer struct {
	logger *logrus.Logger
}

func (h QueryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return ctx
}

func (h QueryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {

}

func addPoolTracer(pool *pgxpool.Pool, logger *logrus.Logger) *pgxpool.Pool {
	config := pool.Config()
	config.ConnConfig.Tracer = QueryTracer{logger: logger}

	return pool
}
