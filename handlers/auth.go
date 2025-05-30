// handlers/auth.go
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"p9e.in/ugcl/config"
	"p9e.in/ugcl/middleware"
	"p9e.in/ugcl/models"
)

type loginPayload struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type registerReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	// hash pw
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "error hashing password", http.StatusInternalServerError)
		return
	}
	u := models.User{
		Name:         req.Name,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: string(hash),
		Role:         req.Role,
	}
	if err := config.DB.Create(&u).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			http.Error(w, "username already taken", http.StatusConflict)
		} else {
			http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusCreated)
}

type loginReq struct {
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type loginResp struct {
	Token string      `json:"token"`
	User  userPayload `json:"user"`
}
type userPayload struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Phone string    `json:"phone"`
	Role  string    `json:"role"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Login request received")

	// ✅ support form values instead of JSON
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	phone := r.FormValue("phone")
	password := r.FormValue("password")

	var u models.User
	if err := config.DB.Where("phone = ?", phone).First(&u).Error; err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := middleware.GenerateToken(u.ID.String(), u.Role, u.Name, u.Phone)
	if err != nil {
		http.Error(w, "couldn't create token", http.StatusInternalServerError)
		return
	}

	u.PasswordHash = ""
	out := loginResp{
		Token: token,
		User: userPayload{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
			Phone: u.Phone,
			Role:  u.Role,
		},
	}

	json.NewEncoder(w).Encode(out)
}

// func Login(w http.ResponseWriter, r *http.Request) {
// 	var req loginReq
// 	fmt.Println("Login request received")
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "invalid JSON", http.StatusBadRequest)
// 		return
// 	}
// 	var u models.User
// 	if err := config.DB.Where("phone = ?", req.Phone).First(&u).Error; err != nil {
// 		http.Error(w, "invalid credentials", http.StatusUnauthorized)
// 		return
// 	}
// 	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
// 		http.Error(w, "invalid credentials", http.StatusUnauthorized)
// 		return
// 	}
// 	token, err := middleware.GenerateToken(u.ID.String(), u.Role, u.Name, u.Phone)
// 	if err != nil {
// 		http.Error(w, "couldn't create token", http.StatusInternalServerError)
// 		return
// 	}
// 	u.PasswordHash = "" // don't leak password hash
// 	out := loginResp{
// 		Token: token,
// 		User: userPayload{
// 			ID:    u.ID,
// 			Name:  u.Name,
// 			Email: u.Email,
// 			Phone: u.Phone,
// 			Role:  u.Role,
// 		},
// 	}
// 	json.NewEncoder(w).Encode(out)
// }

func GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// 1) Extract token
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		http.Error(w, "Missing Bearer token", http.StatusUnauthorized)
		return
	}
	tokenString := strings.TrimPrefix(auth, "Bearer ")

	// 2) Parse & validate
	secret := []byte(os.Getenv("JWT_SECRET"))
	token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	claims := token.Claims.(*models.JWTClaims)

	// 3) Fetch user record
	var user models.User
	if err := config.DB.First(&user, "id = ?", claims.UserID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// 4) Return only the fields you need
	resp := map[string]interface{}{
		"id":    user.ID,
		"name":  user.Name,
		"phone": user.Phone,
		"email": user.Email,
		"roles": user.Role,
	}
	json.NewEncoder(w).Encode(resp)
}
