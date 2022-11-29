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
		Out:   io.MultiWriter(f, os.Stdout),
		Level: logrus.DebugLevel,
		Formatter: &logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		},
	}

	// init db connection pool
	queryLog, err := strconv.ParseBool(os.Getenv("ENABLE_QUERY_LOG"))
	if err != nil {
		queryLog = false
	}

	postgresDB := os.Getenv("POSTGRES_DB")
	if postgresDB == "" {
		fmt.Println("POSTGRES_DB configuration must be provided")
		os.Exit(1)
	}

	postgresUser := os.Getenv("POSTGRES_USER")
	if postgresUser == "" {
		fmt.Println("POSTGRES_USER configuration must be provided")
		os.Exit(1)
	}

	postgresPass := os.Getenv("POSTGRES_PASSWORD")
	if postgresPass == "" {
		fmt.Println("POSTGRES_PASSWORD configuration must be provided")
		os.Exit(1)
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		fmt.Println("DB_HOST configuration must be provided")
		os.Exit(1)
	}

	postgresString := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=disable", postgresUser, postgresPass, dbHost, postgresDB)

	var pool *pgxpool.Pool

	if queryLog {
		//init query  logger
		queryf, err := os.OpenFile("logs/query-log.json", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			logger.Errorf("Failed to create logfile query-log.json %v", err)
			os.Exit(1)
		}
		defer queryf.Close()

		qlogger := &logrus.Logger{
			// Log into f file handler and on os.Stdout
			Out:   io.MultiWriter(queryf, os.Stdout),
			Level: logrus.DebugLevel,
			Formatter: &logrus.JSONFormatter{
				TimestampFormat: time.RFC3339Nano,
			},
		}

		config, err := pgxpool.ParseConfig(postgresString)
		if err != nil {
			logger.Errorf("pgxpool init error %v", err)
			os.Exit(1)
		}

		config.ConnConfig.Tracer = &QueryTracer{logger: qlogger}
		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err != nil {
			logger.Errorf("pgxpool init error %v", err)
			os.Exit(1)
		}
	} else {
		pool, err = pgxpool.New(ctx, postgresString)
		if err != nil {
			logger.Errorf("pgxpool init error %v", err)
			os.Exit(1)
		}
	}

	//init env variables
	JWTSecret := os.Getenv("JWT_SECRET")
	if JWTSecret == "" {
		logger.Errorf("JWT configuration must be provided %v", err)
		os.Exit(1)
	}

	JWTAccessTTL, err := strconv.Atoi(os.Getenv("JWT_ACCESS_TTL"))
	if err != nil {
		logger.Errorf("JWT configuration must be provided %v", err)
		os.Exit(1)
	}

	JWTRefreshTTL, err := strconv.Atoi(os.Getenv("JWT_REFRESH_TTL"))
	if err != nil {
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
	h.logger.Info("Query log query start", data)
	return ctx
}

func (h QueryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	h.logger.Info("Query log query end", data)
}
