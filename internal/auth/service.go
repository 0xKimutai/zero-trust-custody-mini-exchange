package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string
	Status       string
	CreatedAt    time.Time
}

type Service struct {
	db        *sql.DB
	jwtSecret []byte
}

func NewService(db *sql.DB, secret string) *Service {
	return &Service{
		db:        db,
		jwtSecret: []byte(secret),
	}
}

// Register creates a new user with hashed password
func (s *Service) Register(email, password string) (*User, error) {
	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         "user",
		Status:       "active",
	}

	query := `
		INSERT INTO users (email, password_hash, role, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	err = s.db.QueryRow(query, user.Email, user.PasswordHash, user.Role, user.Status).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login verifies credentials and returns a JWT
func (s *Service) Login(email, password string) (string, error) {
	var user User
	query := `SELECT id, password_hash, status FROM users WHERE email = $1`
	err := s.db.QueryRow(query, email).Scan(&user.ID, &user.PasswordHash, &user.Status)
	if err == sql.ErrNoRows {
		return "", errors.New("invalid credentials")
	}
	if err != nil {
		return "", fmt.Errorf("db error: %w", err)
	}

	if user.Status == "frozen" {
		return "", errors.New("account is frozen")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	return s.generateJWT(user.ID)
}

func (s *Service) generateJWT(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateToken parses and validates the JWT, returning the userID
func (s *Service) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if sub, ok := claims["sub"].(string); ok {
			return sub, nil
		}
	}

	return "", errors.New("invalid token claims")
}
