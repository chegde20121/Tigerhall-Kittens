package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/chegde20121/Tigerhall-Kittens/internal/app/database/repositories"
	"github.com/chegde20121/Tigerhall-Kittens/internal/app/models"
	"github.com/golang-jwt/jwt/v4"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type username string
type UserService struct {
	logger *logrus.Logger
	db     *sql.DB
}

var (
	tokenCache = cache.New(5*time.Minute, 10*time.Minute)
)

const (
	secretKey  = "3aU2*GdfLs#4Np9!y^q8gFp6vTm"
	cookieName = "jwt_token"
)

func NewUserService(logger *logrus.Logger, db *sql.DB) *UserService {
	return &UserService{logger: logger, db: db}
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with the input paylod
// @Tags User
// @Accept  json
// @Produce  json
// @Param user body models.User true "Create user"
// @Success 201 {object} models.User
// @Failure 400 {object} models.ErrorResponse "Invalid JSON format"
// @Failure 500 {object} models.ErrorResponse "Failed to create user. Please try again"
// @Router /api/v1/register [post]
func (u *UserService) CreateUser(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	// Bind request payload with our model
	u.logger.Info("Persisting User information")
	user := &models.User{}
	err := user.FormJson(req.Body)
	if err != nil {
		u.logger.Error(err)
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to create user. Invalid JSON format", Status: http.StatusBadRequest})
		return
	}
	user.Password, err = hashPassword(user.Password)
	if err != nil {
		u.logger.Println("Error hashing password:", err)
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to create user. Please try again.", Status: http.StatusInternalServerError})
		return
	}
	userRepo := repositories.NewUserRepository(u.db, u.logger)
	err = userRepo.CreateUser(user)
	if err != nil {
		models.HandleErrorResponse(rw, models.ErrorResponse{Message: "Failed to create user. Please try again", Status: http.StatusInternalServerError})
		return
	}
	u.logger.Info("User created successfully")
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusCreated)
	user.Password = ""
	json.NewEncoder(rw).Encode(user)
}

// hashPassword hashes the given password using bcrypt.
func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// Login godoc
// @Summary Log in a user
// @Description Log in a user with the provided credentials
// @Tags User
// @Accept json
// @Produce json
// @Param credentials body models.Credentials true "User credentials"
// @Success 200 {object} models.LoginResponse
// @Failure 400 {object} models.ErrorResponse "Invalid JSON format"
// @Failure 401 {object} models.ErrorResponse "Invalid user credentials"
// @Failure 500 {object} models.ErrorResponse "Failed to log in. Please try again"
// @Router /api/v1/login [post]
func (u *UserService) Login(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var cred models.Credentials

	err := json.NewDecoder(r.Body).Decode(&cred)
	if err != nil {
		u.logger.Error("Errror occurred while decoding cred request body", err)
		models.HandleErrorResponse(w, models.ErrorResponse{Message: "Failed to Login.Please try again", Status: http.StatusInternalServerError})
		return
	}

	// Check if the token is already in the cache
	if token, found := tokenCache.Get(cred.Username); found {
		setTokenCookie(w, token.(string))
		return
	}
	userRepo := repositories.NewUserRepository(u.db, u.logger)
	user, err := userRepo.GetUserByUserName(cred.Username)
	if err != nil {
		u.logger.Error("Failed to fetch user from repository: ", err)
		models.HandleErrorResponse(w, models.ErrorResponse{Message: "Invalid User or failed to fetch user details.Please try again", Status: http.StatusInternalServerError})
		return
	}
	err = verifyPassword(cred.Password, user.Password)
	if err != nil {
		models.HandleErrorResponse(w, models.ErrorResponse{Message: "Invalid password", Status: http.StatusUnauthorized})
		return
	}

	// Generate JWT token
	token, err := generateToken(cred.Username)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Set the token in the cache
	tokenCache.Set(cred.Username, token, 15*time.Minute)

	setTokenCookie(w, token)
}

// verifyPassword compares a provided password with a stored bcrypt hash
func verifyPassword(providedPassword, hashedPassword string) error {
	// Compare the provided password with the stored hash
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(providedPassword))
}

// generateToken generates a JWT token for the given username
func generateToken(username string) (string, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &models.Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// setTokenCookie sets the JWT token in an HTTP-only secure cookie
func setTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	loginResponse := models.LoginResponse{Message: "Login successful", Token: token}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(loginResponse)

}

// Logout godoc
// @Summary Logout the authenticated user
// @Description Logout the authenticated user and invalidate the session
// @Tags User
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.GeneralResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /api/v1/logout [get]
func (u *UserService) Logout(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	// Check if the user is authenticated
	username, err := getUsernameFromToken(r)
	if err != nil {
		// If the token is not present or invalid, consider the user as not authenticated
		models.HandleErrorResponse(w, models.ErrorResponse{Message: "User not authenticated", Status: http.StatusUnauthorized})
		return
	}

	// Remove the token from the cache
	tokenCache.Delete(username)

	// Delete the token cookie on the client side
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Expires:  time.Now().Add(-time.Hour), // Set an expired time in the past to delete the cookie
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	logoutResponse := models.GeneralResponse{Message: "Logout successful"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(logoutResponse)
}

// getUsernameFromToken retrieves the username from the JWT token in the request
func getUsernameFromToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", err
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil || !token.Valid {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", errors.New("username not found in token claims")
	}

	return username, nil
}

// protectedHandler is a sample protected handler that requires authentication

// AuthMiddleware interceptor to authenticate users
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var username username = "username"

		// Check for Bearer Token in Authorization Header
		bearerToken := extractBearerToken(r)
		if bearerToken != "" {
			// Validate and extract claims from the JWT
			claims, err := validateJWT(bearerToken)
			if err != nil {
				models.HandleErrorResponse(w, models.ErrorResponse{Message: "Invalid or expired token", Status: http.StatusUnauthorized})
				return
			}

			// Attach the username to the request context
			ctx := context.WithValue(r.Context(), username, claims.Username)
			next(w, r.WithContext(ctx))
			return
		}

		// Check for Cookie
		cookie, err := r.Cookie(cookieName)
		if err != nil || cookie.Value == "" {
			models.HandleErrorResponse(w, models.ErrorResponse{Message: "User not authenticated", Status: http.StatusUnauthorized})
			return
		}

		// Validate and extract claims from the JWT in the cookie
		claims, err := validateJWT(cookie.Value)
		if err != nil {
			// Handle invalid or expired token
			models.HandleErrorResponse(w, models.ErrorResponse{Message: "Invalid or expired cookie", Status: http.StatusUnauthorized})
			return
		}

		// Attach the username to the request context
		ctx := context.WithValue(r.Context(), username, claims.Username)
		next(w, r.WithContext(ctx))
	}
}

// extractBearerToken extracts the JWT from the Authorization header.
func extractBearerToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Check if the Authorization header has a Bearer token
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) == 2 && authHeaderParts[0] == "Bearer" {
			return authHeaderParts[1]
		}
	}
	return ""
}

// validateJWT validates the JWT and returns the claims.
func validateJWT(token string) (*models.Claims, error) {
	// Validate and parse the JWT
	parsedToken, err := jwt.ParseWithClaims(token, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	if err != nil || !parsedToken.Valid {
		return nil, errors.New("invalid or expired token")
	}

	claims, ok := parsedToken.Claims.(*models.Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
