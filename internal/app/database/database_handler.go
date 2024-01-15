package database

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/chegde20121/Tigerhall-Kittens/pkg/config"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var user string
var password string
var host string
var port string
var ssl string
var timezone string
var dbConn *sql.DB
var once sync.Once

func init() {
	user = config.GetEnvVar("POSTGRES_USER")
	password = config.GetEnvVar("POSTGRES_PASSWORD")
	host = config.GetEnvVar("POSTGRES_HOST")
	port = config.GetEnvVar("POSTGRES_PORT")
	ssl = config.GetEnvVar("POSTGRES_SSL")
	timezone = config.GetEnvVar("POSTGRES_TIMEZONE")
}
func GetDSN(dbName string) string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s search_path=public", host, user, password, dbName, port, ssl, timezone)
}

func CreateDBConnection() (err error) {
	once.Do(func() {
		// Close the existing connection if open
		if dbConn != nil {
			CloseDB()
		}
		pgConn, err := GetConnection("postgres")
		if err != nil {
			log.Errorf("Error opening database connection: %v", err)
			return
		}
		var isMigrationUp bool
		isMigrationUp, err = strconv.ParseBool(config.GetEnvVar("MIGRATION_UP"))
		if err == nil {
			if isMigrationUp {
				err = runMigrations(pgConn, isMigrationUp, config.GetEnvVar("MIGRATION_FILES_DB"))
				if err != nil {
					log.Warn("Error in migration", err)
				}
				pgConn.Close()
			}
			dbConn, err = GetConnection(config.GetEnvVar("POSTGRES_DB"))
			if err != nil {
				log.Errorf("Error opening database connection: %v", err)
				return
			}
			err = runMigrations(dbConn, isMigrationUp, config.GetEnvVar("MIGRATION_FILES"))
			if err != nil {
				log.Warn("Error in migration", err)
				return
			}
		}
	})
	return err
}

func GetConnection(databaseName string) (dbConn *sql.DB, err error) {
	dbConn, err = sql.Open("postgres", GetDSN(databaseName))
	if err != nil {
		log.Errorf("Error opening database connection: %v", err)
		return
	}
	err = dbConn.Ping()
	if err != nil {
		log.Errorf("Error pinging database: %v", err)
	}
	dbConn.SetConnMaxIdleTime(time.Minute * 5)
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	dbConn.SetMaxIdleConns(10)
	// SetMaxOpenConns sets the maximum number of open connections to the database.
	dbConn.SetMaxOpenConns(100)
	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	dbConn.SetConnMaxLifetime(time.Hour)
	return
}

func GetDB() *sql.DB {
	return dbConn
}

// CloseDB closes the database connection.
func CloseDB() {
	if dbConn != nil {
		err := dbConn.Close()
		if err != nil {
			log.Println("Error closing the database:", err)
		} else {
			log.Println("Database connection closed")
		}
	}
}

func runMigrations(conn *sql.DB, up bool, migration_path string) (err error) {
	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		log.Warnf("error creating migration driver: %v", err)
		return
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", filepath.Join(migration_path)),
		"postgres", driver,
	)
	if err != nil {
		log.Errorf("error creating migration instance: %v", err)
		return
	}

	if up {
		err = m.Up()
		if err != nil && err != migrate.ErrNoChange {
			log.Warnf("error applying migrations: %v", err)
			return
		}
		log.Info("Migrations completed successfully")
	} else {
		err = m.Steps(-1)
		if err != nil && err != migrate.ErrNoChange {
			log.Errorf("error rolling back migrations: %v", err)
			return
		}
		fmt.Println("Migrations rolled back successfully")
	}
	return
}

func RollBack(tx *sql.Tx, err error) {
	if r := recover(); r != nil {
		log.Println("Rolling back transaction due to panic:", r)
		_ = tx.Rollback()
	} else if err != nil {
		log.Println("Rolling back transaction due to error:", err)
		_ = tx.Rollback()
	}
}
