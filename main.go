// @title TigerHall Kittens
// @version 1.0
// @description Users can use a fictional mobile app to add sightings of tigers in the wild.
//
//	For this to work, our API must expose a list of tigers and their recent sightings.
//
// @host localhost:8888
// @securityDefinitions.apikey Authorization
// @in header
// @name Authorization
package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/chegde20121/Tigerhall-Kittens/internal/app/database"
	"github.com/chegde20121/Tigerhall-Kittens/internal/app/handlers"
	"github.com/gorilla/mux"
	"github.com/natefinch/lumberjack"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
)

func main() {
	setupLogger()
	log.Info("starting application")
	if dberr := database.CreateDBConnection(); dberr != nil {
		log.Error("Error occurred while creating the database connection")
	}
	defer database.CloseDB()
	serveMux := mux.NewRouter()
	handlers.RegisterApiHandlers(serveMux)
	server := &http.Server{
		Addr:        ":8888",
		Handler:     serveMux,
		IdleTimeout: 120 * time.Second,
	}
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Error("failed to listen and serve,", err)
		}
	}()
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, syscall.SIGTERM)
	sig := <-sigChan
	log.Info("Recieved terminate, gracefull shutdown", sig)
}

func setupLogger() {
	// Log file configuration
	logFile := &lumberjack.Logger{
		Filename:   "./app.log", // Log file path
		MaxSize:    50,          // Max size in megabytes before rotation
		MaxBackups: 3,           // Max number of old log files to keep
		MaxAge:     30,          // Max number of days to retain log files
		Compress:   true,        // Whether to compress the rotated log files
	}

	// Logrus configuration
	log.SetOutput(os.Stdout) // Direct logs to Stdout

	// JSON formatter for structured logs
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	})

	// Log level configuration (you can change this based on your environment)
	log.SetLevel(log.InfoLevel)

	// Log4j-style log format for console output
	hook := lfshook.NewHook(
		lfshook.WriterMap{
			log.DebugLevel: logFile,
			log.InfoLevel:  logFile,
			log.WarnLevel:  logFile,
			log.ErrorLevel: logFile,
			log.FatalLevel: logFile,
			log.PanicLevel: logFile,
		},
		&log.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				return "", fmt.Sprintf("[%s:%d]", filepath.Base(f.File), f.Line)
			},
		},
	)
	log.AddHook(hook)
}
