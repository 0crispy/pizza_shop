package database

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserRole int

const (
	AdminRole UserRole = iota
	DeliveryRole
	CustomerRole
)

func (r UserRole) String() string {
	switch r {
	case AdminRole:
		return "ADMIN"
	case DeliveryRole:
		return "DELIVERY"
	case CustomerRole:
		return "CUSTOMER"
	}
	panic(fmt.Sprintf("unhandled UserRole: %d", r))
}

type Customer struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	Name        string `json:"name"`
	Gender      string `json:"gender"`
	BirthDate   string `json:"birthDate"`
	NoBirthDate bool   `json:"noBirthDate"`
	Address     string `json:"address"`
	PostCode    string `json:"postcode"`
}

func AddUser(username string, password string, role UserRole) error {
	if len(username) > 100 {
		return errors.New("username is too long (max 100)")
	}

	salt, err := generateSalt(128)
	if err != nil {
		return err
	}

	passwordBytes := append([]byte(password), []byte(salt)...)
	passwordBytes = append(passwordBytes, PASSWORD_HASH_PEPPER...)

	// We have to pre hash. Because bcrypt cant take more than 72 bytes.
	preHash := sha256.Sum256(passwordBytes)

	passwordHash, err := bcrypt.GenerateFromPassword(preHash[:], bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = DATABASE.Exec("INSERT INTO user (username, password_hash, salt, role) VALUES (?, ?, ?, ?)", username, passwordHash, salt, role.String())
	return err
}

func TryAddCustomer(customer Customer) (bool, string) {
	if len(customer.Username) == 0 {
		return false, "Username cannot be empty!"
	}
	if len(customer.Username) > 100 {
		return false, "Username is too long! (max 100 characters)"
	}
	if len(customer.Password) == 0 {
		return false, "Password cannot be empty!"
	}
	if len(customer.Name) == 0 {
		return false, "Name cannot be empty!"
	}
	if len(customer.Name) > 100 {
		return false, "Name is too long! (max 100 characters)"
	}
	if len(customer.Gender) == 0 {
		return false, "Gender cannot be empty!"
	}
	if len(customer.Gender) > 50 {
		return false, "Gender string is too long (max 50 characters)"
	}

	var birthDate sql.NullString
	if !customer.NoBirthDate {
		if _, err := time.Parse("2006-01-02", customer.BirthDate); err != nil {
			return false, "Invalid date!"
		}
		birthDate = sql.NullString{String: customer.BirthDate, Valid: true}
	} else {
		birthDate = sql.NullString{String: "", Valid: false}
	}

	if len(customer.Address) == 0 {
		return false, "Address cannot be empty!"
	}
	if len(customer.Address) > 256 {
		return false, "Address is too long! (max 256 characters)"
	}
	if len(customer.PostCode) == 0 {
		return false, "Post code cannot be empty!"
	}
	if len(customer.PostCode) > 10 {
		return false, "Post code is too long! (max 10 characters)"
	}

	if exists, _ := doesUserExist(customer.Username); exists {
		return false, "Username is already taken!"
	}

	err := AddUser(customer.Username, customer.Password, CustomerRole)
	if err != nil {
		log.Println(err)
		return false, "Something went wrong. Try again."
	}

	userID, err := getUserIDFromUsername(customer.Username)
	if err != nil {
		log.Println(err)
		return false, "Something went wrong. Try again."
	}

	_, err = DATABASE.Exec(
		"INSERT INTO customer (user_id, name, gender, birth_date, address, postal_code) VALUES (?, ?, ?, ?, ?, ?)",
		userID,
		customer.Name,
		customer.Gender,
		birthDate,
		customer.Address,
		customer.PostCode,
	)
	if err != nil {
		log.Println(err)
		return false, "Something went wrong. Try again."
	}
	return true, ""
}

func TryLogin(username string, password string) (bool, string) {
	if len(username) == 0 {
		return false, "Username cannot be empty!"
	}
	user_exists, err := doesUserExist(username)
	if err != nil {
		log.Println(err)
		return false, "Something went wrong. Try again"
	}
	if !user_exists {
		return false, "No user with such username!"
	}

	passwordDB, salt, err := getUserPasswordAndSalt(username)
	if err != nil {
		log.Println(err)
		return false, "Something went wrong. Try again"
	}
	passwordBytes := append([]byte(password), []byte(salt)...)
	passwordBytes = append(passwordBytes, PASSWORD_HASH_PEPPER...)

	preHash := sha256.Sum256(passwordBytes)

	err = bcrypt.CompareHashAndPassword([]byte(passwordDB), preHash[:])
	if err != nil {
		return false, "Incorrect password"
	} else {
		return true, ""
	}
}

func generateSalt(size int) (string, error) {
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func doesUserExist(username string) (bool, error) {
	rows, err := DATABASE.Query("SELECT 1 FROM user WHERE username = ?", username)
	if err != nil {
		return false, err
	}
	if rows.Next() {
		return true, nil
	} else {
		return false, nil
	}
}

func getUserIDFromUsername(username string) (int64, error) {
	rows, err := DATABASE.Query("SELECT id FROM user WHERE username = ?", username)
	if err != nil {
		return 0, err
	}
	if rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			return 0, err
		}
		return id, nil
	} else {
		return 0, errors.New("no such user")
	}
}

func getUserPasswordAndSalt(username string) (string, string, error) {
	rows, err := DATABASE.Query("SELECT password_hash, salt FROM user WHERE username = ?", username)
	if err != nil {
		return "", "", err
	}
	if rows.Next() {
		var salt string
		var password string
		err := rows.Scan(&password, &salt)
		if err != nil {
			return "", "", err
		}

		return password, salt, nil
	} else {
		return "", "", errors.New("no such user")
	}
}
