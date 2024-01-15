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

type TigerRepository struct {
	db     *sql.DB
	logger *log.Logger

	CacheMutex  sync.RWMutex
	CacheExpiry time.Duration
}

var tigersCache *cache.Cache = cache.New(15*time.Minute, 10*time.Minute)

type cachedResult struct {
	tigers     []models.Tiger
	offset     int
	expiryTime time.Time
}

func NewTigerRepository(db *sql.DB, logger *log.Logger) *TigerRepository {
	return &TigerRepository{db: db, logger: logger}
}

func (tr *TigerRepository) CreateTiger(tiger *models.Tiger) (int, error) {
	tx, err := tr.db.Begin()
	if err != nil {
		log.Println("Error beginning transaction:", err)
		return 0, err
	}
	defer database.RollBack(tx, err)
	var tigerID int
	err = tx.QueryRow(
		"INSERT INTO tigerhall.tigers (name, date_of_birth,last_seen_timestamp, last_seen_coordinates_lat, last_seen_coordinates_lon) "+
			"VALUES ($1, $2, $3, $4, $5) RETURNING tiger_id",
		tiger.Name, tiger.DateOfBirth, tiger.LastSeenAt, tiger.LastCoordinates.Latitude, tiger.LastCoordinates.Longitude,
	).Scan(&tigerID)
	if err != nil {
		return 0, err
	}
	err = tx.Commit()
	if err != nil {
		log.Println("Error committing transaction:", err)
		return 0, err
	}
	return tigerID, nil
}

func (tr *TigerRepository) GetTigerIDByName(name string) (int, error) {
	query := "SELECT tiger_id FROM tigerhall.tigers WHERE name = $1"
	var tigerID int
	err := tr.db.QueryRow(query, name).Scan(&tigerID)
	if err != nil {
		tr.logger.Error("Error getting tiger ID by name:", err)
		return 0, err
	}
	return tigerID, nil
}

func (tr *TigerRepository) ListTigers(pageSize int, offset int) (*models.TigersResponse, error) {
	cacheKey := tr.generateCacheKey(pageSize, offset)
	tr.CacheMutex.RLock()
	result, found := tigersCache.Get(cacheKey)
	tr.CacheMutex.RUnlock()
	if found {
		result := result.(cachedResult)
		response := &models.TigersResponse{
			Tigers: result.tigers,
			Offset: result.offset,
		}
		tr.logger.Infof("results fetched from cache: [%v] tigers retrieved", len(result.tigers))
		return response, nil
	}
	// Data not found in the cache, fetch from the database
	return tr.FetchTigersAndStore(pageSize, offset, cacheKey)
}

func (tr *TigerRepository) FetchTigersAndStore(pageSize int, offset int, cacheKey string) (*models.TigersResponse, error) {
	tigers, nextOffset, err := tr.FetchFromDatabase(pageSize, offset)
	if err != nil {
		return nil, err
	}
	tr.CacheMutex.Lock()
	tigersCache.Set(cacheKey, cachedResult{
		tigers:     tigers,
		offset:     nextOffset,
		expiryTime: time.Now().Add(tr.CacheExpiry),
	}, cache.DefaultExpiration)
	tr.CacheMutex.Unlock()
	tr.logger.Infof("Fetched %v tigers from database", len(tigers))
	return &models.TigersResponse{Tigers: tigers, Offset: nextOffset}, nil
}

func (tr *TigerRepository) FetchFromDatabase(pageSize int, offset int) (tigers []models.Tiger, nextOffset int, err error) {
	// Placeholder for the SQL query to fetch tigers
	// query_with_token := `
	// 	SELECT tiger_id, name, date_of_birth, last_seen_timestamp, last_seen_coordinates_lat, last_seen_coordinates_lon
	// 	FROM tigerhall.tigers
	// 	WHERE (tiger_id,last_seen_timestamp) < ($1, $2)
	// 	ORDER BY last_seen_timestamp DESC, tiger_id DESC
	// 	LIMIT $3;
	// `

	query := `
		SELECT tiger_id, name, date_of_birth, last_seen_timestamp, last_seen_coordinates_lat, last_seen_coordinates_lon
		FROM tigerhall.tigers
		ORDER BY last_seen_timestamp DESC, tiger_id DESC
		LIMIT $1 OFFSET $2;
	`
	// if len(continueToken) == 0 {
	// 	query = query_without_token
	// 	rows, err = tr.db.Query(query, pageSize+1)
	// 	if err != nil {
	// 		tr.logger.Error("Error querying tigers:", err)
	// 		return tigers, "", err
	// 	}
	// } else {
	rows, err := tr.db.Query(query, pageSize+1, offset)
	if err != nil {
		tr.logger.Error("Error querying tigers:", err)
		return tigers, 0, err
	}
	// }
	defer rows.Close()
	// Iterate through the rows and populate the tigers slice
	for rows.Next() {
		var tiger models.Tiger
		err := rows.Scan(
			&tiger.ID,
			&tiger.Name,
			&tiger.DateOfBirth,
			&tiger.LastSeenAt,
			&tiger.LastCoordinates.Latitude,
			&tiger.LastCoordinates.Longitude,
		)
		if err != nil {
			tr.logger.Error("Error scanning row:", err)
			return tigers, 0, err
		}

		// Append the tiger to the result slice
		tigers = append(tigers, tiger)
	}
	// Determine the next continue token based on the fetched results
	if len(tigers) > pageSize {
		// If there are more results than the requested page size, set the next continue token to the last seen timestamp and ID of the last tiger
		nextOffset = offset + pageSize
		if err != nil {
			tr.logger.Error("Error encoding token:", err)
			return tigers, 0, err
		}
		tigers = tigers[:pageSize] // Trim the slice to the requested page size
	}
	return tigers, nextOffset, nil
}

func (tr *TigerRepository) generateCacheKey(pageSize int, offset int) string {
	// Customize the cache key based on your specific requirements
	return "list_tigers:" + strconv.Itoa(pageSize) + ":" + fmt.Sprint(offset)
}
