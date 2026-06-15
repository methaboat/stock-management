package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"stock-management/backend/app"
)

type Service interface {
	Authenticate(ctx context.Context, req LoginRequest) (*LoginResponse, error)
}

type authService struct {
	userCol *mongo.Collection
}

func NewService(userCol *mongo.Collection) Service {
	return &authService{userCol: userCol}
}

func (s *authService) Authenticate(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	var user app.User
	err := s.userCol.FindOne(ctx, bson.M{"username": req.Username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Generate Token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &app.Claims{
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(app.JwtKey)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token:    tokenStr,
		Username: user.Username,
		Role:     user.Role,
	}, nil
}
