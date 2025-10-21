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
	if len(username) == 0 {
		return errors.New("username cannot be empty")
	}
	if len(username) > 100 {
		return errors.New("username is too long (max 100)")
	}
	if len(password) == 0 {
		return errors.New("password cannot be empty")
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

func GetUserIDFromUsername(username string) (int, error) {
	var userID int
	err := DATABASE.QueryRow("SELECT id FROM user WHERE username = ?", username).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func GetCustomerIDFromUserID(userID int) (int, error) {
	var customerID int
	err := DATABASE.QueryRow("SELECT id FROM customer WHERE user_id = ?", userID).Scan(&customerID)
	if err != nil {
		return 0, err
	}
	return customerID, nil
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
	}

	// Get user role
	var role string
	err = DATABASE.QueryRow("SELECT role FROM user WHERE username = ?", username).Scan(&role)
	if err != nil {
		log.Println(err)
		return false, "Something went wrong. Try again"
	}

	return true, role
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
	defer rows.Close()
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

// GetUserRole returns the role string (e.g., "ADMIN", "DELIVERY", "CUSTOMER") for a given username.
func GetUserRole(username string) (string, error) {
	rows, err := DATABASE.Query("SELECT role FROM user WHERE username = ?", username)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	if rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return "", err
		}
		return role, nil
	}
	return "", errors.New("no such user")
}

func GetCustomerDetails(username string, password string) (Customer, error) {
	exists, err := doesUserExist(username)
	if err != nil {
		return Customer{}, err
	}
	if !exists {
		return Customer{}, errors.New("customer not found")
	}

	userID, err := getUserIDFromUsername(username)
	if err != nil {
		return Customer{}, err
	}

	rows, err := DATABASE.Query(
		"SELECT name, gender, birth_date, address, postal_code FROM customer where user_id = ?",
		userID,
	)
	if err != nil {
		return Customer{}, err
	}

	defer rows.Close()

	if rows.Next() {
		var customer Customer
		var birthDate sql.NullString
		err := rows.Scan(&customer.Name, &customer.Gender, &birthDate, &customer.Address, &customer.PostCode)
		if err != nil {
			return Customer{}, err
		}

		if birthDate.Valid {
			customer.BirthDate = birthDate.String
		} else {
			customer.BirthDate = ""
		}

		customer.Username = username
		customer.Password = password
		return customer, nil

	} else {
		return Customer{}, errors.New("customer not found")
	}

}

// Admin functions for user management

func GetAllUsers() ([]map[string]interface{}, error) {
	query := `
		SELECT u.id, u.username, u.role, 
			COALESCE(c.name, dp.name, 'N/A') as name,
			COALESCE(c.gender, 'N/A') as gender,
			COALESCE(c.birth_date, '') as birth_date,
			COALESCE(c.address, 'N/A') as address,
			COALESCE(c.postal_code, 'N/A') as postal_code
		FROM user u
		LEFT JOIN customer c ON u.id = c.user_id
		LEFT JOIN delivery_person dp ON u.id = dp.user_id
		WHERE u.role != 'ADMIN'
		ORDER BY u.id DESC
	`
	rows, err := DATABASE.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id int
		var username, role, name, gender, address, postalCode string
		var birthDate sql.NullString
		err := rows.Scan(&id, &username, &role, &name, &gender, &birthDate, &address, &postalCode)
		if err != nil {
			return nil, err
		}

		birthDateStr := ""
		if birthDate.Valid {
			birthDateStr = birthDate.String
		}

		user := map[string]interface{}{
			"id":          id,
			"username":    username,
			"role":        role,
			"name":        name,
			"gender":      gender,
			"birth_date":  birthDateStr,
			"address":     address,
			"postal_code": postalCode,
		}
		users = append(users, user)
	}

	// Return empty array instead of nil if no users found
	if users == nil {
		users = []map[string]interface{}{}
	}

	return users, nil
}

func DeleteUser(userID int) error {
	// Start transaction
	tx, err := DATABASE.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get user role to determine what to delete
	var role string
	err = tx.QueryRow("SELECT role FROM user WHERE id = ?", userID).Scan(&role)
	if err != nil {
		return err
	}

	// Don't allow deleting admin users
	if role == "ADMIN" {
		return errors.New("cannot delete admin user")
	}

	// Delete customer or delivery_person records
	if role == "CUSTOMER" {
		// Get customer ID
		var customerID int
		err = tx.QueryRow("SELECT id FROM customer WHERE user_id = ?", userID).Scan(&customerID)
		if err == nil {
			// Delete orders and order items for this customer
			_, err = tx.Exec("DELETE op FROM order_pizza op INNER JOIN `orders` o ON op.order_id = o.id WHERE o.customer_id = ?", customerID)
			if err != nil {
				return err
			}
			_, err = tx.Exec("DELETE FROM `orders` WHERE customer_id = ?", customerID)
			if err != nil {
				return err
			}
		}
		_, err = tx.Exec("DELETE FROM customer WHERE user_id = ?", userID)
		if err != nil {
			return err
		}
	} else if role == "DELIVERY" {
		_, err = tx.Exec("DELETE FROM delivery_person WHERE user_id = ?", userID)
		if err != nil {
			return err
		}
	}

	// Delete user
	_, err = tx.Exec("DELETE FROM user WHERE id = ?", userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// CheckCustomerBirthday checks if today is the customer's birthday
func CheckCustomerBirthday(userID int64) (bool, error) {
	var birthDate sql.NullTime
	query := `
		SELECT c.birth_date 
		FROM customer c 
		WHERE c.user_id = ?
	`
	err := DATABASE.QueryRow(query, userID).Scan(&birthDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	if !birthDate.Valid {
		return false, nil
	}

	// Check if today's month and day match the birth date
	now := time.Now()
	return now.Month() == birthDate.Time.Month() && now.Day() == birthDate.Time.Day(), nil
}
