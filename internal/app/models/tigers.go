package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// Tiger represents a tiger in the wild.
// swagger:model
type Tiger struct {
	ID              int        `json:"tiger_id" swaggerignore:"true"`
	Name            string     `json:"name" validate:"required"`
	DateOfBirth     CustomTime `json:"date_of_birth" validate:"required"`       // Example: "18/02/1998"
	LastSeenAt      UnixTime   `json:"last_seen_timestamp" validate:"required"` // Example: "1705147765"
	LastCoordinates struct {
		Latitude  float64 `json:"last_seen_coordinates_lat"`
		Longitude float64 `json:"last_seen_coordinates_lon"`
	} `json:"last_coordinates" validate:"required"`
	Sightings []Sighting `json:"sightings,omitempty" swaggerignore:"true"` // Relationship with sightings
}

// TigersResponse represents the response containing a list of tigers and an offset.
// swagger:model
type TigersResponse struct {
	// Tigers is the list of tigers.
	Tigers []Tiger `json:"tigers"`

	// Offset is the offset for paginating the list.
	Offset int `json:"offset"`
}

type CustomTime struct {
	time.Time
}

type UnixTime struct {
	time.Time
}

const customTimeLayout = "\"02/01/2006\"" // Adjust the layout based on your date format
const customDateLayout = "02/01/2006"

func (c *CustomTime) UnmarshalJSON(b []byte) error {
	strDate := string(b)
	parsedTime, err := time.Parse(customTimeLayout, strDate)
	if err != nil {
		if parsedTime, err = time.Parse(customDateLayout, strDate); err != nil {
			return err
		}
	}
	c.Time = parsedTime
	return nil
}

func (u *CustomTime) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		u.Time = v
		return nil
	default:
		return fmt.Errorf("unsupported type for CustomTime: %T", value)
	}
}

func (u CustomTime) Value() (driver.Value, error) {
	return u.Time, nil
}

func (t *Tiger) FormJson(reader io.Reader) error {
	e := json.NewDecoder(reader)
	return e.Decode(t)
}

func (u *UnixTime) UnmarshalJSON(b []byte) error {
	var timestamp int64
	if err := json.Unmarshal(b, &timestamp); err != nil {
		return err
	}
	u.Time = time.Unix(timestamp, 0)
	return nil
}

func (u *UnixTime) Scan(value interface{}) error {
	switch v := value.(type) {
	case time.Time:
		u.Time = v
		return nil
	default:
		return fmt.Errorf("unsupported type for UnixTime: %T", value)
	}
}

func (u UnixTime) Value() (driver.Value, error) {
	return u.Time, nil
}

// CreateTiger godoc
// swagger:parameter createTiger
type CreateTigerRequest struct {
	// Name of the tiger
	//
	// Required: true
	// Example: RajahuliBangalore
	Name string `json:"name"`

	// Date of Birth of the tiger
	//
	// Required: true
	// Example: "18/02/1998"
	DateOfBirth string `json:"date_of_birth"`

	// Last seen timestamp of the tiger in Unix Epoch Time UTC format
	//
	// Required: true
	// Example: 1705147765
	LastSeenTimestamp int64 `json:"last_seen_timestamp"`

	// Last coordinates where the tiger was seen
	//
	// Required: true
	LastCoordinates struct {
		// Latitude of the last seen coordinates
		//
		// Required: true
		// Example: 37.7749
		Latitude float64 `json:"last_seen_coordinates_lat"`

		// Longitude of the last seen coordinates
		//
		// Required: true
		// Example: -122.4194
		Longitude float64 `json:"last_seen_coordinates_lon"`
	} `json:"last_coordinates"`
}
