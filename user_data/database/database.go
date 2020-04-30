package database

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"general"
	"io"
)

// This key will be exported to a file
const key string = "WlAye5L1uzZq61p41A6PyhpBxsnklABk6FPAOeOXUwqWuouUEvTG8Apqkqo1uloZ"

// Database will be used to extract dependencies on db
type Database interface {
	SignUp(username, password string) (int, error)
	Login(username, password string) (general.Credentials, error)
	FindUser(username string) (general.Credentials, error)
}

// UserDB is a sql database
type UserDB struct {
	database *sql.DB
}

// NewUserDB returns a UserDB
func NewUserDB(db *sql.DB) *UserDB {
	return &UserDB{database: db}
}

// SignUp adds a new user to the database and returns the newly added id
func (db *UserDB) SignUp(username, password string) (int, error) {
	salt := generateSalt()
	hash := hashPass(password, string(salt))
	result, err := db.database.Exec("INSERT INTO users(username, password,salt, role) VALUES ( ?, ?,?, ?)", username, hash, salt, "user")
	if err != nil {
		return 0, general.MySQLErrorToDBError(err)
	}
	userID, errorID := result.LastInsertId()
	if errorID != nil {
		return 0, general.ErrorToUnknownDBError(errorID)
	}
	return int(userID), nil
}

func generateSalt() []byte {
	randomBytes := make([]byte, 64)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil
	}
	return randomBytes
}

func hashPass(password string, salt string) string {
	hash := hmac.New(sha512.New, []byte(key))
	io.WriteString(hash, password+salt)
	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}

// Login compares the given password with the password from the database.
func (db *UserDB) Login(username, password string) (general.Credentials, error) {
	var userdata general.Credentials
	var passwordDB, saltDB string
	err := db.database.QueryRow("SELECT id, username,role, password,salt FROM users WHERE username=?", username).Scan(&userdata.ID, &userdata.Username, &userdata.Role, &passwordDB, &saltDB)
	if err != nil {
		if err != sql.ErrNoRows {
			return general.Credentials{}, general.ErrorToUnknownDBError(err)
		}
		return general.Credentials{}, general.GetDBError("The credentials do not match", general.InvalidInput)
	}
	hash := hashPass(password, string(saltDB))
	if hash != passwordDB {
		return general.Credentials{}, general.GetDBError("The credentials do not match", general.InvalidInput)
	}
	return userdata, nil
}

// FindUser searches the database for an user with the given username
func (db *UserDB) FindUser(username string) (general.Credentials, error) {
	var userdata general.Credentials
	err := db.database.QueryRow("SELECT id, username, role FROM users WHERE username=?", username).Scan(&userdata.ID, &userdata.Username, &userdata.Role)
	if err != nil {
		if err != sql.ErrNoRows {
			return general.Credentials{}, general.ErrorToUnknownDBError(err)
		}
		return general.Credentials{}, general.GetDBError("Can't find user", general.NotFoundError)
	}
	return userdata, nil
}
