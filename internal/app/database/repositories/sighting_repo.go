package repositories

import (
	"database/sql"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/chegde20121/Tigerhall-Kittens/internal/app/database"
	"github.com/chegde20121/Tigerhall-Kittens/internal/app/models"
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
)

type SightingRepository struct {
	db          *sql.DB
	logger      *log.Logger
	CacheMutex  sync.RWMutex
	CacheExpiry time.Duration
}

type cachedSighitingResult struct {
	sightings  []models.Sighting
	nextOffset int
	expiryTime time.Time
}

var sightingCache *cache.Cache = cache.New(15*time.Minute, 10*time.Minute)

func NewSightingRepository(db *sql.DB, logger *log.Logger) *SightingRepository {
	return &SightingRepository{db: db, logger: logger}
}

func (sr *SightingRepository) CreateSight(sighting *models.Sighting) (int, error) {
	tx, err := sr.db.Begin()
	if err != nil {
		log.Println("Error beginning sransaction:", err)
		return 0, err
	}
	defer database.RollBack(tx, err)
	query := `
	INSERT INTO tigerhall.sightings (
		tiger_id,
		user_id,
		last_seen_timestamp,
		last_seen_coordinates_lat,
		last_seen_coordinates_lon,
		image
	) VALUES ($1, $2, $3, $4, $5, $6) RETURNING sighting_id
`
	var sightingID int
	err = sr.db.QueryRow(
		query,
		sighting.TigerID,
		sighting.User.ID,
		sighting.Timestamp.Time,
		sighting.LastCoordinates.Latitude,
		sighting.LastCoordinates.Longitude,
		sighting.ImageBlob,
	).Scan(&sightingID)
	if err != nil {
		return 0, err
	}
	err = tx.Commit()
	if err != nil {
		log.Println("Error committing sransaction:", err)
		return 0, err
	}
	return sightingID, nil
}

func (r *SightingRepository) GetPreviousSightingCoordinates(tigerID int) (struct{ Latitude, Longitude float64 }, error) {
	query := "SELECT last_seen_coordinates_lat, last_seen_coordinates_lon FROM tigerhall.sightings WHERE tiger_id = $1 ORDER BY last_seen_timestamp DESC LIMIT 1"
	row := r.db.QueryRow(query, tigerID)

	var coordinates struct{ Latitude, Longitude float64 }
	err := row.Scan(&coordinates.Latitude, &coordinates.Longitude)
	if err != nil {
		if err == sql.ErrNoRows {
			// No previous sighting found
			return coordinates, nil
		}
		log.Printf("Error resrieving previous sighting coordinates: %v", err)
		return coordinates, fmt.Errorf("error resrieving previous sighting coordinates")
	}

	return coordinates, nil
}

func (sr *SightingRepository) ListSightings(tigerId int, pageSize int, offset int) (*models.SightingsResponse, error) {
	cacheKey := sr.generateCacheKey(tigerId, pageSize, offset)
	sr.CacheMutex.RLock()
	result, found := sightingCache.Get(cacheKey)
	sr.CacheMutex.RUnlock()
	if found {
		result := result.(cachedSighitingResult)
		response := &models.SightingsResponse{
			Sightings: result.sightings,
			Offset:    result.nextOffset,
		}
		sr.logger.Infof("results fetched from cache: [%v] sightings resrieved", len(result.sightings))
		return response, nil
	}
	// Data not found in the cache, fetch from the database
	return sr.FetchSightingsAndStore(tigerId, pageSize, offset, cacheKey)
}

func (sr *SightingRepository) FetchSightingsAndStore(tigerId int, pageSize int, offset int, cacheKey string) (*models.SightingsResponse, error) {
	sightings, nextOffset, err := sr.FetchFromDatabase(tigerId, pageSize, offset)
	if err != nil {
		return nil, err
	}
	sr.CacheMutex.Lock()
	if len(sightings) > 0 {
		sightingCache.Set(cacheKey, cachedSighitingResult{
			sightings:  sightings,
			nextOffset: nextOffset,
			expiryTime: time.Now().Add(sr.CacheExpiry),
		}, cache.DefaultExpiration)
		sr.CacheMutex.Unlock()
	}
	sr.logger.Infof("Fetched %v sightings from database", len(sightings))
	return &models.SightingsResponse{Sightings: sightings, Offset: nextOffset}, nil
}

func (sr *SightingRepository) FetchFromDatabase(tigerId, pageSize int, offset int) (sightings []models.Sighting, nextOffset int, err error) {
	query := `
		SELECT sighting_id,tiger_id, last_seen_timestamp, last_seen_coordinates_lat, last_seen_coordinates_lon,image
		FROM tigerhall.sightings where tiger_id = $1 
		ORDER BY last_seen_timestamp DESC, sighting_id DESC
		LIMIT $2 OFFSET $3;
	`
	rows, err := sr.db.Query(query, tigerId, pageSize+1, offset)
	if err != nil {
		sr.logger.Error("Error querying sightings:", err)
		return sightings, 0, err
	}
	defer rows.Close()
	// Iterate through the rows and populate the sightings slice
	for rows.Next() {
		var sighting models.Sighting
		err := rows.Scan(
			&sighting.ID,
			&sighting.TigerID,
			&sighting.Timestamp,
			&sighting.LastCoordinates.Latitude,
			&sighting.LastCoordinates.Longitude,
			&sighting.ImageBlob,
		)
		if err != nil {
			sr.logger.Error("Error scanning row:", err)
			return sightings, 0, err
		}
		// Append the sightings to the result slice
		sightings = append(sightings, sighting)
	}
	// Determine the next continue token based on the fetched results
	if len(sightings) > pageSize {
		// If there are more results than the requested page size, set the next continue token to the last seen timestamp and ID of the last sightings
		nextOffset = offset + pageSize
		sightings = sightings[:pageSize] // srim the slice to the requested page size
	}
	return sightings, nextOffset, nil
}

func (sr *SightingRepository) generateCacheKey(tiger_id int, pageSize int, offset int) string {
	// Customize the cache key based on your specific requirements
	return "list_sightings:" + "tiger_id:" + strconv.Itoa(tiger_id) + ":" + strconv.Itoa(pageSize) + ":" + fmt.Sprint(offset)
}
