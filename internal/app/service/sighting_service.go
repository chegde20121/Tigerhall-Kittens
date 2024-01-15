package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chegde20121/Tigerhall-Kittens/internal/app/database/repositories"
	"github.com/chegde20121/Tigerhall-Kittens/internal/app/models"
	"github.com/chegde20121/Tigerhall-Kittens/pkg/config"
	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
	log "github.com/sirupsen/logrus"
)

type SightingService struct {
	db     *sql.DB
	logger *log.Logger
}

func NewSightingService(db *sql.DB, logger *log.Logger) *SightingService {
	return &SightingService{db: db, logger: logger}
}

// CreateSight godoc
// @Summary Create a new tiger sighting
// @Description Create a new tiger sighting with the provided information.
// @Tags Sighting
// @Accept multipart/form-data
// @Produce json
// @Param tiger_name formData string true "Name of the tiger"
// @Param last_seen_timestamp formData string true "Timestamp of the sighting (unix epoch utc format example:1705147765)"
// @Param last_seen_coordinates_lat formData number true "Latitude of the sighting coordinates"
// @Param last_seen_coordinates_lon formData number true "Longitude of the sighting coordinates"
// @Param image formData file true "Image of the sighting"
// @Success 201 {object} models.SightingsResponse
// @Failure 400 {object} models.ErrorResponse "Invalid input format"
// @Failure 401 {object} models.ErrorResponse "Unauthorized"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Security Authorization
// @Router /api/v1/createSights [post]
func (s *SightingService) CreateSight(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	//Read request and parse payload
	contentType := req.Header.Get("Content-Type")
	if !strings.Contains(contentType, "multipart/form-data") {
		s.logger.Error("Unsupported Content-Type")
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Unsupported Content-Type. Please try again.", Status: http.StatusInternalServerError})
		return
	}
	sightings, err := ParseSightingFormData(req)
	if err != nil {
		s.logger.Error("Failed to parse the request body")
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to parse the request payload. Please try again.", Status: http.StatusInternalServerError})
		return
	}
	//Fetch tiger id
	tigerRepo := repositories.NewTigerRepository(s.db, s.logger)
	tigerId, err := tigerRepo.GetTigerIDByName(sightings.TigerName)
	if err != nil {
		s.logger.Error("Unsupported Content-Type")
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to fetch tiger details. Please try again.", Status: http.StatusInternalServerError})
		return
	}
	if tigerId < 1 {
		s.logger.Error("No tigers with given name", sightings.TigerName)
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to fetch tiger details. Please try again.", Status: http.StatusInternalServerError})
		return
	}
	sightings.TigerID = tigerId
	//Fetch User details
	userRepo := repositories.NewUserRepository(s.db, s.logger)
	var user *models.User
	username, err := getUsernameFromToken(req)
	if err == nil {
		user, err = userRepo.GetUserByUserName(username)
		if err != nil {
			s.logger.Error("Failed to fetch user details")
			models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to create sighting. Please try again.", Status: http.StatusInternalServerError})
			return
		}
	} else {
		s.logger.Error("username is not present in context", sightings.TigerName)
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to create sighting. Please try again.", Status: http.StatusInternalServerError})
		return
	}
	sightings.User = user

	//Get Previous co-ordinates

	sightingRepo := repositories.NewSightingRepository(s.db, s.logger)
	coordinates, err := sightingRepo.GetPreviousSightingCoordinates(sightings.TigerID)
	if err != nil {
		s.logger.Error("Failed to fetch previous sightings.", sightings.TigerName)
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to fetch previous sightings. Please try again.", Status: http.StatusInternalServerError})
		return
	}
	if IsWithinRange(sightings.LastCoordinates.Latitude, sightings.LastCoordinates.Longitude, coordinates.Latitude, coordinates.Longitude) {
		id, err := sightingRepo.CreateSight(sightings)
		if err != nil {
			s.logger.Error("Failed to create sightings.", sightings.TigerName)
			models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to create sightings. Please try again.", Status: http.StatusInternalServerError})
			return
		}
		sightings.ID = id
		sightings.Image = base64.StdEncoding.EncodeToString(sightings.ImageBlob)
	} else {
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Cannot submit new sighting. A tiger has been sighted within 5 kilometers recently.", Status: http.StatusInternalServerError})
		return
	}
	messagequeue := GetMessagingQueue()
	notifier := NewNotifier(s.db, s.logger)
	_ = notifier.RegisterSightingSubscriber()
	s.logger.Info("Sighting created successfully")
	users, err := userRepo.GetUsersByTigerId(sightings.TigerID, int(user.ID))
	if err != nil {
		s.logger.Error("Failed to create sightings.", sightings.TigerName)
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to fetch users. Please try again.", Status: http.StatusInternalServerError})
		return
	}
	sightingNotification := SightingNotification{
		TigerName:     sightings.TigerName,
		TigerLocation: fmt.Sprintf("Latitude:%v,Longitude:%v", sightings.LastCoordinates.Latitude, sightings.LastCoordinates.Longitude),
		SightingTime:  sightings.Timestamp.Format("2006-01-02,15:04:05"),
	}
	useremails := []string{}
	for i := range users {
		useremails = append(useremails, users[i].Email)
	}
	sightingNotification.userEmails = useremails
	messagequeue.Publish(sightingNotification)
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	json.NewEncoder(rw).Encode(sightings)

}

func IsWithinRange(lat, lon, prevLat, prevLon float64) bool {
	distance := haversineDistance(lat, lon, prevLat, prevLon)
	return distance >= 5.0
}

func ParseSightingFormData(r *http.Request) (*models.Sighting, error) {
	// err := r.ParseMultipartForm(600 << 20) // 10 MB limit for the form data
	// if err != nil {
	// 	return nil, fmt.Errorf("error parsing form data: %v", err)
	// }

	latitude, err := strconv.ParseFloat(r.FormValue("last_seen_coordinates_lat"), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid last_seen_coordinates_lat: %v", err)
	}

	longitude, err := strconv.ParseFloat(r.FormValue("last_seen_coordinates_lon"), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid last_seen_coordinates_lon: %v", err)
	}

	timestampUnix, err := strconv.ParseInt(r.FormValue("last_seen_timestamp"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid last_seen_timestamp: %v", err)
	}

	timestamp := models.UnixTime{
		Time: time.Unix(timestampUnix, 0),
	}

	image, _, err := r.FormFile("image")
	if err != nil {
		return nil, fmt.Errorf("error retrieving image: %v", err)
	}
	defer image.Close()

	imageBlob, err := resizeImage(image)
	if err != nil {
		return nil, fmt.Errorf("error resizing image: %v", err)
	}

	s := &models.Sighting{
		TigerName: r.FormValue("tiger_name"),
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

	return s, nil
}

func resizeImage(imageFile io.Reader) ([]byte, error) {
	img, _, err := image.Decode(imageFile)
	if err != nil {
		return nil, fmt.Errorf("error decoding image: %v", err)
	}

	resizedImg := resize.Resize(250, 200, img, resize.Lanczos3)

	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, resizedImg, nil); err != nil {
		return nil, fmt.Errorf("error encoding image: %v", err)
	}

	return buffer.Bytes(), nil
}

// ListAllSightings godoc
// @Summary List all sightings
// @Description Get a paginated list of all sightings
// @Tags Sighting
// @Accept json
// @Produce json
// @ID list-tiger-sightings
// @Param id path int true "Tiger ID"
// @Param pageSize query int false "Number of sightings to retrieve per page"
// @Param offset query int false "Offset for paginating the list"
// @Success 200 {object} models.SightingsResponse "List of sightings"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /api/v1/tigers/:id/listSightings [get]
func (s *SightingService) ListAllSightings(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	tigerID, err := strconv.Atoi(vars["id"])
	if err != nil {
		s.logger.Error("Failed to parse tiger id")
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to parse the request payload. Please try again.", Status: http.StatusInternalServerError})
		return
	}
	pageSizeStr := req.URL.Query().Get("pageSize")
	offsetStr := req.URL.Query().Get("offset")
	if pageSizeStr == "" {
		pageSizeStr = config.GetEnvVar("PAGE_SIZE")
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		s.logger.Errorf("Failed to parse PageSize: %v", err)
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to parse the request payload. Please try again.", Status: http.StatusInternalServerError})
		return
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil && offsetStr != "" {
		s.logger.Errorf("Failed to parse PageSize: %v", err)
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to parse the request payload. Please try again.", Status: http.StatusInternalServerError})
		return
	}
	tigerRepo := repositories.NewSightingRepository(s.db, s.logger)
	tigerRepo.CacheMutex = sync.RWMutex{}
	tigerRepo.CacheExpiry = 15 * time.Minute
	response, err := tigerRepo.ListSightings(tigerID, pageSize, offset)
	if err != nil {
		s.logger.Errorf("Failed to fetch all sightings: %v", err)
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to fetch sightings. Please try again.", Status: http.StatusInternalServerError})
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)
}
