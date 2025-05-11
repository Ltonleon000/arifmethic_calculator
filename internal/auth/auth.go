package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Ключ для подписи JWT
var jwtKey = []byte("simple_calculator_key")

type contextKey string

const userIDKey contextKey = "userID"

// UserClaims описывает содержимое JWT
type UserClaims struct {
	UserID string `json:"user_id"`
	Login  string `json:"login"`
	jwt.RegisteredClaims
}

// HashPassword хеширует пароль для безопасного хранения
func HashPassword(password string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash)
}

// CheckPassword проверяет, соответствует ли пароль хешу
func CheckPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// RegisterHandler — регистрация пользователя
func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
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
		
		// Проверяем, что пользователь с таким логином не существует
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE login = ?", req.Login).Scan(&count)
		if err != nil || count > 0 {
			http.Error(w, "user already exists", http.StatusConflict)
			return
		}
		
		id := uuid.New().String()
		hash := HashPassword(req.Password)
		_, err = db.Exec("INSERT INTO users (id, login, password) VALUES (?, ?, ?)", id, req.Login, hash)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}

// LoginHandler — вход пользователя и выдача JWT
func LoginHandler(db *sql.DB) http.HandlerFunc {
	defaultTokenTTL := time.Hour * 24 // Токен действителен 24 часа
	
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
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
		
		if !CheckPassword(hash, req.Password) {
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
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"token": signed})
	}
}

// JWTMiddleware проверяет JWT и добавляет userID в context
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, "missing or invalid token", http.StatusUnauthorized)
			return
		}
		
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims := &UserClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		
		if err != nil || !token.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID извлекает userID из context
func GetUserID(r *http.Request) string {
	val := r.Context().Value(userIDKey)
	if id, ok := val.(string); ok {
		return id
	}
	return ""
}
