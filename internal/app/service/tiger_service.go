package service

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chegde20121/Tigerhall-Kittens/internal/app/database/repositories"
	"github.com/chegde20121/Tigerhall-Kittens/internal/app/models"
	"github.com/chegde20121/Tigerhall-Kittens/pkg/config"
	"github.com/sirupsen/logrus"
)

type TigerService struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewTigerService(db *sql.DB, logger *logrus.Logger) *TigerService {
	return &TigerService{db: db, logger: logger}
}

// CreateTiger godoc
// @Summary Create a new tiger
// @Description Create a new tiger using either JSON or multipart form data
// @Tags Tiger
// @Accept json
// @Param tiger body models.CreateTigerRequest true "Create Tiger"
// @Success 201 {object} models.CreateTigerRequest
// @Failure 400 {object} models.ErrorResponse
// @Security Authorization
// @Router /api/v1/createTigers [post]
func (t *TigerService) CreateTiger(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	contentType := req.Header.Get("Content-Type")
	if contentType == "application/json" {
		t.CreateTigerFromJson(ctx, rw, req)
	} else if strings.Contains(contentType, "multipart/form-data") {
		t.CreateTigerFromMultiPartForm(ctx, rw, req)
	} else {
		t.logger.Error("Unsupported Content-Type")
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Unsupported Content-Type. Please try again.", Status: http.StatusInternalServerError})

	}
}

func (t *TigerService) CreateTigerFromMultiPartForm(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	tiger, sightings, err := ParseFormData(req)
	if err != nil {
		t.logger.Error("Failed to parse the request body")
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to parse the request payload. Please try again.", Status: http.StatusInternalServerError})
		return
	}
	tigerRepo := repositories.NewTigerRepository(t.db, t.logger)
	id, err := tigerRepo.CreateTiger(tiger)
	if err != nil {
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to create tiger. Please try again", Status: http.StatusInternalServerError})
		return
	}
	tiger.ID = id
	if len(sightings.ImageBlob) > 0 {
		userRepo := repositories.NewUserRepository(t.db, t.logger)
		var user *models.User
		username, err := getUsernameFromToken(req)
		if err == nil {
			user, err = userRepo.GetUserByUserName(username)
			if err != nil {
				t.logger.Error("Failed to fetch user details")
				models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to create sighting. Please try again.", Status: http.StatusInternalServerError})
				return
			}
		} else {
			t.logger.Error("username is not present in context", username)
			models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to create sighting. Please try again.", Status: http.StatusInternalServerError})
			return
		}
		sightings.TigerID = tiger.ID
		sightings.User = user
		sightingRepo := repositories.NewSightingRepository(t.db, t.logger)
		id, err := sightingRepo.CreateSight(sightings)
		if err != nil {
			t.logger.Error("Failed to create sightings.", sightings.TigerName)
			models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to create sightings. Please try again.", Status: http.StatusInternalServerError})
			return
		}
		sightings.ID = id
		sightings.Image = base64.StdEncoding.EncodeToString(sightings.ImageBlob)
	}
	t.logger.Info("tiger created successfully")
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	json.NewEncoder(rw).Encode(tiger)
}

func ParseFormData(r *http.Request) (*models.Tiger, *models.Sighting, error) {
	latitude, err := strconv.ParseFloat(r.FormValue("last_seen_coordinates_lat"), 64)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid last_seen_coordinates_lat: %v", err)
	}
	longitude, err := strconv.ParseFloat(r.FormValue("last_seen_coordinates_lon"), 64)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid last_seen_coordinates_lon: %v", err)
	}
	timestampUnix, err := strconv.ParseInt(r.FormValue("last_seen_timestamp"), 10, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid last_seen_timestamp: %v", err)
	}
	dobParsed := models.CustomTime{}
	dobParsed.UnmarshalJSON([]byte(r.FormValue("date_of_birth")))
	timestamp := models.UnixTime{
		Time: time.Unix(timestampUnix, 0),
	}
	image, _, err := r.FormFile("image")
	var imageBlob []byte
	if err != nil && !strings.Contains(err.Error(), "no such file") {
		return nil, nil, fmt.Errorf("error retrieving image: %v", err)
	} else if err == nil {
		defer image.Close()
		imageBlob, err = resizeImage(image)
		if err != nil {
			return nil, nil, fmt.Errorf("error resizing image: %v", err)
		}
	}

	tiger := &models.Tiger{
		Name:        r.FormValue("name"),
		DateOfBirth: dobParsed,
		LastSeenAt:  timestamp,
		LastCoordinates: struct {
			Latitude  float64 "json:\"last_seen_coordinates_lat\""
			Longitude float64 "json:\"last_seen_coordinates_lon\""
		}{
			Latitude:  latitude,
			Longitude: longitude,
		},
	}
	s := &models.Sighting{
		Timestamp: timestamp,
		LastCoordinates: struct {
			Latitude  float64 `json:"last_seen_coordinates_lat"`
			Longitude float64 `json:"last_seen_coordinates_lon"`
		}{
			Latitude:  latitude,
			Longitude: longitude,
		},
		ImageBlob: imageBlob,
	}

	return tiger, s, nil
}

func (t *TigerService) CreateTigerFromJson(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	t.logger.Info("creating new tiger")
	tiger := &models.Tiger{}
	err := tiger.FormJson(req.Body)
	if err != nil {
		t.logger.Error(err)
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to create tiger. Invalid JSON format", Status: http.StatusBadRequest})
		return
	}
	tigerRepo := repositories.NewTigerRepository(t.db, t.logger)
	id, err := tigerRepo.CreateTiger(tiger)
	if err != nil {
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to create tiger. Please try again", Status: http.StatusInternalServerError})
		return
	}
	tiger.ID = id
	t.logger.Info("tiger created successfully")
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	json.NewEncoder(rw).Encode(tiger)
}

// ListAllTigers godoc
// @Summary List all tigers
// @Description Retrieve a list of tigers with optional pagination.
// @Tags Tiger
// @Accept json
// @Produce json
// @Param pageSize query int false "Number of tigers per page"
// @Param offset query int false "Pagination offset"
// @Success 200 {object} models.TigersResponse
// @Failure 400 {object} models.ErrorResponse "Invalid input format"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Router /api/v1/listTigers [get]
func (t *TigerService) ListAllTigers(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	pageSizeStr := req.URL.Query().Get("pageSize")
	offsetStr := req.URL.Query().Get("offset")
	if pageSizeStr == "" {
		pageSizeStr = config.GetEnvVar("PAGE_SIZE")
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		t.logger.Errorf("Failed to parse PageSize: %v", err)
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to parse the request payload. Please try again.", Status: http.StatusInternalServerError})
		return
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil && offsetStr != "" {
		t.logger.Errorf("Failed to parse PageSize: %v", err)
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to parse the request payload. Please try again.", Status: http.StatusInternalServerError})
		return
	}
	tigerRepo := repositories.NewTigerRepository(t.db, t.logger)
	tigerRepo.CacheMutex = sync.RWMutex{}
	tigerRepo.CacheExpiry = 15 * time.Minute
	response, err := tigerRepo.ListTigers(pageSize, offset)
	if err != nil {
		t.logger.Errorf("Failed to fetch all tigers: %v", err)
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to fetch the tigers. Please try again.", Status: http.StatusInternalServerError})
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)
}
