package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"calculator/internal/models"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var jwtKey = []byte("super_secret_key")

// UserClaims описывает содержимое JWT
// (можно расширить при необходимости)
type UserClaims struct {
	UserID string `json:"user_id"`
	Login  string `json:"login"`
	jwt.RegisteredClaims
}

// RegisterHandler — регистрация пользователя
func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Login    string `json:"login"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if req.Login == "" || req.Password == "" {
			http.Error(w, "login and password required", http.StatusBadRequest)
			return
		}
		id := uuid.New().String()
		hash := models.HashPassword(req.Password)
		_, err := db.Exec("INSERT INTO users (id, login, password) VALUES (?, ?, ?)", id, req.Login, hash)
		if err != nil {
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

// LoginHandler — вход пользователя и выдача JWT
func LoginHandler(db *sql.DB) http.HandlerFunc {
	defaultTokenTTL := time.Hour * 24
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Login    string `json:"login"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		var id, hash string
		err := db.QueryRow("SELECT id, password FROM users WHERE login = ?", req.Login).Scan(&id, &hash)
		if err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		if !models.CheckPassword(hash, req.Password) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		claims := UserClaims{
			UserID: id,
			Login:  req.Login,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(defaultTokenTTL)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString(jwtKey)
		if err != nil {
			http.Error(w, "could not sign token", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": signed})
	}
}
