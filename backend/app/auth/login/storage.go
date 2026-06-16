package login

import (
	"context"
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"stock-management/backend/app"
	"stock-management/backend/config"
	"time"
)

type Storage struct {
	col *mongo.Collection
}

func NewStorage(col *mongo.Collection) *Storage {
	return &Storage{col: col}
}

func (s *Storage) Authenticate(ctx context.Context, req Request) (*Response, error) {
	var user app.User
	err := s.col.FindOne(ctx, bson.M{"username": req.Username}).Decode(&user)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	expiry := time.Now().Add(config.C.JWTExpiry)
	claims := &app.Claims{
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(app.JwtKey)
	if err != nil {
		return nil, err
	}

	return &Response{Token: tokenStr, Username: user.Username, Role: user.Role}, nil
}
