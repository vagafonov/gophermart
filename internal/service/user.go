package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"gophermart/internal/contract"
	"gophermart/internal/entity"
	"gophermart/internal/errs"
	"time"
)

const AuthTokenExp = time.Hour

type User struct {
	userRepo contract.UserRepository
}

type AuthClaims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID
}

func NewUser(userRepo contract.UserRepository) contract.UserService {
	return &User{userRepo}
}

func (s *User) GetUserIDFromJWT(tokenString string) (uuid.UUID, error) {
	claims := &AuthClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(viper.GetString("crypto_key")), nil
		})
	if err != nil {
		return uuid.Nil, err
	}
	if !token.Valid {
		return uuid.Nil, errs.ErrJWTTokenNotValid
	}

	return claims.UserID, nil
}

func (s *User) Create(ctx context.Context, l string, p string) (*entity.User, error) {
	u, err := s.userRepo.GetByLogin(ctx, l)
	if err != nil && !errors.Is(err, errs.ErrNotFound) {
		return nil, fmt.Errorf("can't get user by login from repo: %w", err)
	}
	if u != nil {
		return nil, errs.ErrAlreadyExists
	}

	bcrypted, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("could not hash password: %w", err)
	}

	u, err = s.userRepo.Add(ctx, l, string(bcrypted))
	if err != nil {
		return nil, fmt.Errorf("could not add user: %w", err)
	}

	return u, nil
}

func (s *User) GetAuthToken(userID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, AuthClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AuthTokenExp)),
		},
		UserID: userID,
	})
	tokenString, err := token.SignedString([]byte(viper.GetString("crypto_key")))
	if err != nil {
		return "", fmt.Errorf("failed sign token:%w", err)
	}

	return tokenString, nil
}

func (s *User) GetByLoginAndPassword(ctx context.Context, p string, l string) (*entity.User, error) {
	u, err := s.userRepo.GetByLogin(ctx, l)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return nil, nil //nolint:nilnil
		}

		return nil, fmt.Errorf("can't get user by login from repo: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(p)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, nil //nolint:nilnil
		}

		return nil, fmt.Errorf("could not hash password: %w", err)
	}

	return u, nil
}
