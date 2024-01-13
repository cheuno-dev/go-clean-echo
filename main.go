package main

import (
	"flag"
	"log"
	"os"

	"github.com/zett-8/go-clean-echo/db"
	database "github.com/zett-8/go-clean-echo/db"
	_ "github.com/zett-8/go-clean-echo/docs"
	"github.com/zett-8/go-clean-echo/handlers"
	"github.com/zett-8/go-clean-echo/logger"
	"github.com/zett-8/go-clean-echo/middlewares"
	"github.com/zett-8/go-clean-echo/services"
	"github.com/zett-8/go-clean-echo/stores"
	"go.uber.org/zap"
)

var goEnv = os.Getenv("GO_ENV")

func initArgv(dbConfig *db.Config) {
	isDev := os.Getenv("GO_ENV") == "development"

	// -------------- DB --------------
	flag.StringVar(&dbConfig.PostgresURI, "postgresURI", os.Getenv("POSTGRES_URI"), "Specify a URI for Postgres")

	flag.BoolVar(&dbConfig.SeedTestData, "seedTestData", isDev, "Should seed test data")

	flag.Parse()
}

// @title Go clean echo API v1
// @version 1.0
// @description This is a sample server.
// @termsOfService http://swagger.io/terms/

// @host localhost:8888
// @BasePath /
// @schemes http
func main() {
	err := logger.New()

	var dbConfig db.Config
	initArgv(&dbConfig)

	if err != nil {
		log.Fatal(err)
	}

	db, err := database.New(&dbConfig)
	if err != nil {
		logger.Fatal("failed to connect to the database", zap.Error(err))
	}
	defer db.Close()

	e := handlers.Echo()

	s := stores.New(db)
	ss := services.New(s)
	h := handlers.New(ss)

	jwtCheck, err := middlewares.JwtMiddleware()
	if err != nil {
		logger.Fatal("failed to set JWT middleware", zap.Error(err))
	}

	handlers.SetDefault(e)
	handlers.SetApi(e, h, jwtCheck)

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8888"
	}

	logger.Fatal("failed to start server", zap.Error(e.Start(":"+PORT)))
}
