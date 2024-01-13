package main

import (
	"flag"
	"log"
	"os"

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

var envName = os.Getenv("MY_NAME_ENV_VAR")
var postgresDbName string

func initArgv() {
	var postgresDbName string

	flag.StringVar(&postgresDbName, "postgresDbName", os.Getenv("postgresDbName"), "Specify a postgresDbName")

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
	initArgv()

	err := logger.New()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.New(goEnv == "development")
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
