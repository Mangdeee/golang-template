package utils

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/felixlambertv/go-cleanplate/internal/controller/response"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type TokenClaims struct {
	jwt.StandardClaims
	Authorized bool                   `json:"authorized"`
	User       *response.UserResponse `json:"user"`
	Expire     int64                  `json:"expire"`
}

type Token struct {
	Token   string
	Expires time.Time
}

func GenerateToken(user *response.UserResponse, lifespan int, duration string, secret string) (*Token, error) {
	token := &Token{}
	expTime := time.Time{}

	switch duration {
	case "minute":
		expTime = time.Now().Add(time.Minute * time.Duration(lifespan))
	case "second":
		expTime = time.Now().Add(time.Second * time.Duration(lifespan))
	default:
		expTime = time.Now().Add(time.Hour * time.Duration(lifespan))
	}

	claims := TokenClaims{}
	claims.Authorized = true
	claims.User = user
	claims.Expire = expTime.Unix()
	unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, errSign := unsignedToken.SignedString([]byte(secret))

	token.Token = signedToken
	token.Expires = expTime.UTC()

	return token, errSign
}

func ParseToken(tokenString string, secret string) (*TokenClaims, error) {
	claims := &TokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*TokenClaims)
	if ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token invalid")
}

func EncryptPassword(password string) (string, error) {
	//turn password into hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), err
}

func GenerateRandomStringToken(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)

	if _, err := rand.Read(b); err != nil {
		return ""
	}

	return hex.EncodeToString(b)
}

func GenerateRandomToken() int {
	min := 1000
	max := 9999
	// set seed
	rand.Seed(time.Now().UnixNano())
	// generate random number between 1000 - 9999
	return rand.Intn(max-min+1) + min
}
