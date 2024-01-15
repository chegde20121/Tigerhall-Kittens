package models

// Credentials represents the user login credentials.
// swagger:model Credentials
type Credentials struct {
	// Password of the user.
	// required: true
	// example: MySecretPassword
	// min length: 6
	Password string `json:"password"`

	// Username of the user.
	// required: true
	// example: john_doe
	Username string `json:"username"`
}
