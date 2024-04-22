package server

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"math-calc/internal/application"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

const jwtSecretKey = "~I?EMyb77IHo~GPLIdGtEM}}5HXyp~\xab\xcd\xef"

type UserCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func hashPassword(password string, salt string) string {
	sum := sha256.Sum256([]byte(password + salt))
	return fmt.Sprintf("%x", sum)
}

func issueJWT(userId int) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": strconv.Itoa(userId),
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"jti": rand.Int(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(jwtSecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}

func checkJWT(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(jwtSecretKey), nil
	})
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		sub, err := claims.GetSubject()
		if err != nil {
			return 0, fmt.Errorf("failed to get subject: %w", err)
		}
		return strconv.Atoi(sub)
	}
	return 0, fmt.Errorf("failed to parse claims")
}

func userRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to read request body: %s", err)
		return
	}

	creds := UserCredentials{}
	err = json.Unmarshal(body, &creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "failed to unparse json: %s", err)
		return
	}

	if creds.Login == "" || creds.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "login and password are required")
		return
	}

	app := r.Context().Value("app").(*application.Application)

	// Check if user already exists
	_, err = app.Database.GetUserByUsername(creds.Login)
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintln(w, "user already exists")
		return
	}

	// Generate salt and hash password
	salt := fmt.Sprintf("%d", rand.Int())
	hash := hashPassword(creds.Password, salt)

	// Create user
	_, err = app.Database.CreateUser(creds.Login, salt, hash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to create user: %s", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"status": "ok"}`)
}

func userLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to read request body: %s", err)
		return
	}

	creds := UserCredentials{}
	err = json.Unmarshal(body, &creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "failed to unparse json: %s", err)
		return
	}

	app := r.Context().Value("app").(*application.Application)
	user, err := app.Database.GetUserByUsername(creds.Login)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "user not found")
		return
	}

	hash := hashPassword(creds.Password, user.PasswordSalt)
	if hash != user.PasswordHash {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintln(w, "invalid password")
		return
	}

	token, err := issueJWT(user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to issue token: %s", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"token": "%s"}`, token)
}
