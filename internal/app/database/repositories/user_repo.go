package repositories

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/chegde20121/Tigerhall-Kittens/internal/app/database"
	"github.com/chegde20121/Tigerhall-Kittens/internal/app/models"
	"github.com/sirupsen/logrus"
)

type UserRepository struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewUserRepository(db *sql.DB, logger *logrus.Logger) *UserRepository {
	return &UserRepository{db: db, logger: logger}
}

func (ur *UserRepository) CreateUser(user *models.User) error {
	tx, err := ur.db.Begin()
	if err != nil {
		log.Println("Error beginning transaction:", err)
		return err
	}

	defer database.RollBack(tx, err)

	_, err = tx.Exec("INSERT INTO tigerhall.users (username, password_hash, email) VALUES ($1, $2, $3)", user.Username, user.Password, user.Email)
	if err != nil {
		log.Println("Error inserting user:", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Println("Error committing transaction:", err)
		return err
	}
	return err
}

func (ur *UserRepository) GetUserByUserName(username string) (*models.User, error) {
	query := "SELECT user_id, username, password_hash, email FROM tigerhall.users WHERE username=$1"
	row := ur.db.QueryRow(query, username)

	user := &models.User{}
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found for username %s", username)
		}
		return nil, fmt.Errorf("error scanning user row: %v", err)
	}

	return user, nil
}

func (ur *UserRepository) GetUsersByTigerId(tigerID int, userID int) ([]models.User, error) {
	query := `
		SELECT DISTINCT u.user_id, u.username, u.email
		FROM tigerhall.users u
		JOIN tigerhall.sightings s ON u.user_id = s.user_id
		WHERE s.tiger_id = $1 and u.user_id != $2
	`
	rows, err := ur.db.Query(query, tigerID, userID)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	// Slice to store users
	var users []models.User
	// Iterate over the result set and populate the users slice
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email); err != nil {
			log.Fatal(err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return users, nil
}
