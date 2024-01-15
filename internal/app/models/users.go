package models

import (
	"encoding/json"
	"io"

	"github.com/golang-jwt/jwt/v4"
)

// swagger:model
// User represents a user in the system.
type User struct {
	ID uint `json:"user_id,omitempty" swaggerignore:"true"`
	// Name of user
	// swagger:meta
	// in: body
	// required: true
	Username string `json:"username" validate:"required"`
	// Password of user
	// swagger:meta
	// in: body
	// required: true
	Password string `json:"password" validate:"required"`
	// Email of user
	// swagger:meta
	// in: body
	// required: true
	Email     string     `json:"email" validate:"required"`
	Sightings []Sighting `json:"sightings,omitempty" swaggerignore:"true"` // Relationship with sightings
}

// swagger:parameters CreateUserRequest
type CreateUserRequest struct {
	// User information to be created
	// in: body
	// required: true
	Body User
}

// swagger:response
type UserResponse struct {
	// in: body
	Body User
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func (u *User) FormJson(reader io.Reader) error {
	e := json.NewDecoder(reader)
	return e.Decode(u)
}

func (u *User) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(u)
}

// LoginResponse represents a login response for user login.
type LoginResponse struct {
	Message string `json:"message,omitempty"`
	Token   string `json:"token,omitempty"`
}

// GeneralResponse represents a login response for user login.
type GeneralResponse struct {
	Message string `json:"message,omitempty"`
}
