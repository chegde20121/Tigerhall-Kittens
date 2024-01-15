package models

import (
	"encoding/json"
	"io"
)

// Sighting represents a sighting of a tiger.
type Sighting struct {
	ID              int      `json:"sighting_id"`
	TigerID         int      `json:"tiger_id"`
	Timestamp       UnixTime `json:" last_seen_timestamp"`
	LastCoordinates struct {
		Latitude  float64 `json:"last_seen_coordinates_lat"`
		Longitude float64 `json:"last_seen_coordinates_lon"`
	} `json:"last_coordinates"`
	ImageBlob []byte `json:"image,omitempty"`
	User      *User  `json:"user,omitempty"` // Relationship with the user who reported the sighting
	TigerName string `json:"tigername,omitempty"`
	Image     string `json:"encoded_image,omitempty"`
}

func (s *Sighting) FormJson(reader io.Reader) error {
	e := json.NewDecoder(reader)
	return e.Decode(s)
}

// SightingRequest represents a request for creating a new sighting.
// swagger:parameters createSighting
type SightingRequest struct {
	// Tiger ID associated with the sighting
	//
	// required: true
	// example: 1
	TigerID int `json:"tiger_id"`

	// Timestamp of the sighting in string format (e.g., "18/02/1998")
	//
	// required: true
	// example: "18/02/1998"
	Timestamp string `json:"last_seen_timestamp"`

	// Last coordinates where the tiger was seen
	//
	// required: true
	// example: {"last_seen_coordinates_lat": 37.7749, "last_seen_coordinates_lon": -122.4194}
	LastCoordinates struct {
		// Latitude of the last seen coordinates
		//
		// required: true
		// example: 37.7749
		Latitude float64 `json:"last_seen_coordinates_lat"`

		// Longitude of the last seen coordinates
		//
		// required: true
		// example: -122.4194
		Longitude float64 `json:"last_seen_coordinates_lon"`
	} `json:"last_coordinates"`

	// Image blob associated with the sighting
	//
	// example: "base64-encoded-image-data"
	ImageBlob []byte `json:"image,omitempty" swaggerignore:"true"`

	// User who reported the sighting
	//
	// example: {"user_id": 1, "username": "john_doe"}
	User *User `json:"user,omitempty" swaggerignore:"true"`

	// Name of the tiger associated with the sighting
	//
	// example: "Rajahuli"
	TigerName string `json:"tigername,omitempty"`

	// Encoded image data associated with the sighting
	//
	// example: "base64-encoded-image-data"
	Image string `json:"encoded_image,omitempty"`
}

// SightingsResponse represents a response containing a list of sightings.
// swagger:model
type SightingsResponse struct {
	// List of sightings
	//
	// required: true
	Sightings []Sighting `json:"sightings"`

	// Offset for paginating through the list of sightings
	//
	// required: true
	// example: 0
	Offset int `json:"offset"`
}
