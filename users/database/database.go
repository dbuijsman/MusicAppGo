package database

import (
	"MusicAppGo/common"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"io"
)

// This key will be exported to a file
const key string = "WlAye5L1uzZq61p41A6PyhpBxsnklABk6FPAOeOXUwqWuouUEvTG8Apqkqo1uloZ"

// Database will be used to extract dependencies on db
type Database interface {
	InsertUser(username, password string) error
	Login(username, password string) (bool, error)
	FindUser(username string) (RowUserDB, error)
}

// UserDB is a sql database
type UserDB struct {
	database *sql.DB
}

// NewUserDB returns a UserDB
func NewUserDB(db *sql.DB) *UserDB {
	return &UserDB{database: db}
}

// RowUserDB represents a row of user credentials in the database.
type RowUserDB struct {
	Username string `json: "username"`
	password string `json: "-"`
	salt     []byte `json: "-"`
	Role     string `json: "role"`
}

// NewRowUserDB returns a row with the given username and password
func NewRowUserDB(username, password string) *RowUserDB {
	return &RowUserDB{Username: username, password: password}
}

// InsertUser adds a new user to the database
func (db *UserDB) InsertUser(username, password string) error {
	salt := generateSalt()
	hash := hashPass(password, string(salt))
	_, err := db.database.Exec("INSERT INTO users(username, password,salt) VALUES ( ?, ?,?)", username, hash, salt)
	if err != nil {
		if err.Error()[6:10] == "1062" {
			return common.GetDBError(err.Error(), common.DuplicateEntry)
		}
		return common.GetDBError(err.Error(), common.UnknownError)
	}
	return nil
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
func (db *UserDB) Login(username, password string) (bool, error) {
	var userdata RowUserDB
	err := db.database.QueryRow("SELECT username, password,salt FROM users WHERE username=?", username).Scan(&userdata.Username, &userdata.password, &userdata.salt)
	if err != nil {
		return false, common.GetDBError(err.Error(), common.UnknownError)
	}
	if userdata.Username == "" {
		return false, nil
	}
	hash := hashPass(password, string(userdata.salt))
	return hash == userdata.password, nil
}

// FindUser searches the database for an user with the given username
func (db *UserDB) FindUser(username string) (RowUserDB, error) {
	var userdata RowUserDB
	err := db.database.QueryRow("SELECT username, role FROM users WHERE username=?", username).Scan(&userdata.Username, &userdata.Role)
	if err != nil {
		return RowUserDB{}, common.GetDBError(err.Error(), common.UnknownError)
	}
	return userdata, nil
}
