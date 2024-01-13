package db

import (
	"database/sql"
	"embed"
	"io/ioutil"

	"github.com/DATA-DOG/go-sqlmock"
	// Importing the "github.com/lib/pq" package for its side effects.
	// This package registers the PostgreSQL driver with the database/sql package.
	// The blank identifier (_) is used to import the package solely for its side effects.
	// It is a common practice to import packages solely for their side effects.
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/zett-8/go-clean-echo/logger"
	"go.uber.org/zap"
)

// Config represents the configuration for the database.
type Config struct {
	// PostgresURI is the postgres connection string.
	PostgresURI string

	// PostgresMaxIdleConnections is the maximum number of connections in the idle connection pool.
	PostgresMaxIdleConnections int

	// PostgresMaxOpenConnections is the maximum number of open connections to the database.
	PostgresMaxOpenConnections int

	// SeedTestData is a flag to seed the database with test data.
	SeedTestData bool
}

//go:embed migrations/*.sql
var migrationsFS embed.FS

// New creates a new database connection based on the provided configuration.
func New(dbConfig *Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", dbConfig.PostgresURI)

	logger.Info("connecting to the database", zap.String("postgresURI", dbConfig.PostgresURI))

	if err != nil {
		return nil, err
	}

	migrations := &migrate.EmbedFileSystemMigrationSource{
		FileSystem: migrationsFS,
		Root:       "migrations",
	}

	if _, err := migrate.Exec(db, "postgres", migrations, migrate.Up); err != nil {
		return nil, err
	}

	if dbConfig.SeedTestData {
		seedFiles, err := ioutil.ReadDir("db/seed/")
		if err != nil {
			logger.Error("failed to read seed files", zap.Error(err))
		}

		for _, f := range seedFiles {
			c, err := ioutil.ReadFile("db/seed/" + f.Name())
			if err != nil {
				logger.Error("failed to read seed file", zap.Error(err))
			}

			sqlCode := string(c)

			_, err = db.Exec(sqlCode)
			if err != nil {
				logger.Error("failed to seed database", zap.Error(err))
			}
		}

		logger.Info("seeded database")
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Mock creates a mock database connection and returns the database and mock object.
func Mock() (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))

	if err != nil {
		logger.Fatal("failed to create mock db", zap.Error(err))
	}

	return db, mock
}
