package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	"gofermart/internal/gophermart/core/model"
	"gofermart/internal/gophermart/core/repositories"
)

func (a *Application) UserRegister(ctx context.Context, request model.User) (string, error) {
	newPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("can't hash password: %w", err)
	}

	if err := a.repo.CreateUser(ctx, request.Login, string(newPassword)); err != nil {
		if errors.Is(err, repositories.ErrDuplicate) {
			return "", fmt.Errorf("user already exists: %w", ErrUserExists)
		}

		return "", fmt.Errorf("can't create user: %w", err)
	}

	token, err := a.generateJwtToken(request.Login)
	if err != nil {
		return "", fmt.Errorf("can't generate token: %w", err)
	}

	return token, nil
}

func (a *Application) UserLogin(ctx context.Context, request model.User) (string, error) {
	password, err := a.repo.GetUserPassword(ctx, request.Login)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return "", fmt.Errorf("user not found: %w", ErrUserNotFound)
		}

		return "", fmt.Errorf("can't get user password: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(request.Password)); err != nil {
		return "", fmt.Errorf("invalid password: %w, %w", err, ErrIncorrectPass)
	}

	token, err := a.generateJwtToken(request.Login)
	if err != nil {
		return "", fmt.Errorf("can't generate token: %w", err)
	}

	return token, nil
}

func (a *Application) generateJwtToken(login string) (string, error) {
	claims := &jwt.RegisteredClaims{
		Issuer:    login,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := []byte(a.secret)

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("can't sign token: %w", err)
	}

	return tokenString, nil
}

func (a *Application) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.secret), nil
	})
	if err != nil {
		return "", fmt.Errorf("can't parse token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token claims")
	}

	if claims["exp"] == nil {
		return "", errors.New("invalid token claims")
	}

	exp, ok := claims["exp"].(float64)
	if !ok || time.Now().Unix() > int64(exp) {
		return "", errors.New("token expired")
	}

	login, ok := claims["iss"].(string)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	return login, nil
}
