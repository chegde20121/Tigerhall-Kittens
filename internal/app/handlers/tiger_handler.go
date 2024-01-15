package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/chegde20121/Tigerhall-Kittens/internal/app/database"
	"github.com/chegde20121/Tigerhall-Kittens/internal/app/service"
	log "github.com/sirupsen/logrus"
)

type TigerHandler struct {
	logger *log.Logger
}

func NewTigerHanlder(logger *log.Logger) *TigerHandler {
	return &TigerHandler{logger: logger}
}

func (t *TigerHandler) CreateTiger(rw http.ResponseWriter, req *http.Request) {
	tigerService := service.NewTigerService(database.GetDB(), t.logger)
	ctx, cancel := context.WithTimeout(req.Context(), 120*time.Second)
	defer cancel()
	tigerService.CreateTiger(ctx, rw, req)
}

func (t *TigerHandler) ListTigers(rw http.ResponseWriter, req *http.Request) {
	tigerService := service.NewTigerService(database.GetDB(), t.logger)
	ctx, cancel := context.WithTimeout(req.Context(), 120*time.Second)
	defer cancel()
	tigerService.ListAllTigers(ctx, rw, req)
}
