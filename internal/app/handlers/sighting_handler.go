package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/chegde20121/Tigerhall-Kittens/internal/app/database"
	"github.com/chegde20121/Tigerhall-Kittens/internal/app/service"
	log "github.com/sirupsen/logrus"
)

type SightingHandler struct {
	logger *log.Logger
}

func NewSightingHandler(logger *log.Logger) *SightingHandler {
	return &SightingHandler{logger: logger}
}

func (sh *SightingHandler) CreateSight(rw http.ResponseWriter, req *http.Request) {
	sightingSvc := service.NewSightingService(database.GetDB(), sh.logger)
	ctx, cancel := context.WithTimeout(req.Context(), 120*time.Second)
	defer cancel()
	sightingSvc.CreateSight(ctx, rw, req)
}

func (sh *SightingHandler) ListAllSightings(rw http.ResponseWriter, req *http.Request) {
	sightingSvc := service.NewSightingService(database.GetDB(), sh.logger)
	ctx, cancel := context.WithTimeout(req.Context(), 120*time.Second)
	defer cancel()
	sightingSvc.ListAllSightings(ctx, rw, req)
}
