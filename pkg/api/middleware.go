package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type AuthRequest struct {
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type Token struct {
	Password string `json:"password"`
	Exp      int64  `json:"exp"`
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendResponse(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "unsupported request method"})
		return
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendResponse(w, http.StatusBadRequest, ErrorResponse{Error: "password is required"})
		return
	}

	if req.Password != appConfig.Password {
		sendResponse(w, http.StatusUnauthorized, ErrorResponse{Error: "wrong password"})
		return
	}

	token := generateJWT(req.Password)

	sendResponse(w, http.StatusOK, AuthResponse{Token: token})
}

func generateJWT(password string) string {
	expirationTime := time.Now().Add(8 * time.Hour).Unix()
	tokenString := fmt.Sprintf("%s:%d", password, expirationTime)
	token := base64.StdEncoding.EncodeToString([]byte(tokenString))
	return token
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(appConfig.Password) > 0 {
			session, err := r.Cookie("token")
			if err != nil || !isValidSession(session.Value, appConfig.Password) {
				sendResponse(w, http.StatusUnauthorized, ErrorResponse{Error: "authentification required"})
				return
			}
		}
		next(w, r)
	}
}

func isValidSession(token, expectedPassword string) bool {
	decodedToken, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		log.Printf("token decode error: %v", err)
		return false
	}

	parts := strings.Split(string(decodedToken), ":")
	if len(parts) != 2 {
		log.Printf("token parsing error: %v", err)
		return false
	}

	password := parts[0]
	expirationTimeStr := parts[1]

	var expirationTime int64
	_, err = fmt.Sscanf(expirationTimeStr, "%d", &expirationTime)
	if err != nil {
		log.Printf("texpiration time conversion error: %v", err)
		return false
	}

	currentTime := time.Now().Unix()
	if currentTime > expirationTime {
		log.Println("the token expired")
		return false
	}

	if password != expectedPassword {
		log.Println("wrong password")
		return false
	}

	return true
}
