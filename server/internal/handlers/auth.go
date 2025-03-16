package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spotdemo4/trevstack/server/internal/interceptors"
	"github.com/spotdemo4/trevstack/server/internal/models"
	userv1 "github.com/spotdemo4/trevstack/server/internal/services/user/v1"
	"github.com/spotdemo4/trevstack/server/internal/services/user/v1/userv1connect"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db  *gorm.DB
	key []byte
}

func (h *AuthHandler) Login(ctx context.Context, req *connect.Request[userv1.LoginRequest]) (*connect.Response[userv1.LoginResponse], error) {
	// Validate
	user := models.User{}
	if err := h.db.First(&user, "username = ?", req.Msg.Username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("invalid username or password"))
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Msg.Password)); err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("invalid username or password"))
	}

	// Generate JWT
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:  "trevstack",
		Subject: strconv.FormatUint(uint64(user.ID), 10),
		IssuedAt: &jwt.NumericDate{
			Time: time.Now(),
		},
		ExpiresAt: &jwt.NumericDate{
			Time: time.Now().Add(time.Hour * 24),
		},
	})
	ss, err := t.SignedString(h.key)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Create cookie
	cookie := http.Cookie{
		Name:     "token",
		Value:    ss,
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	res := connect.NewResponse(&userv1.LoginResponse{
		Token: ss,
	})
	res.Header().Set("Set-Cookie", cookie.String())
	return res, nil
}

func (h *AuthHandler) SignUp(ctx context.Context, req *connect.Request[userv1.SignUpRequest]) (*connect.Response[userv1.SignUpResponse], error) {
	// Validate
	if err := h.db.First(&models.User{}, "username = ?", req.Msg.Username).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	} else {
		return nil, connect.NewError(connect.CodeAlreadyExists, errors.New("username already exists"))
	}
	if req.Msg.Password != req.Msg.ConfirmPassword {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("passwords do not match"))
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Msg.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Create user
	user := models.User{
		Username: req.Msg.Username,
		Password: string(hash),
	}
	if err := h.db.Create(&user).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&userv1.SignUpResponse{})
	return res, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *connect.Request[userv1.LogoutRequest]) (*connect.Response[userv1.LogoutResponse], error) {
	// Clear cookie
	cookie := http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	res := connect.NewResponse(&userv1.LogoutResponse{})
	res.Header().Set("Set-Cookie", cookie.String())
	return res, nil
}

func NewAuthHandler(db *gorm.DB, key string) (string, http.Handler) {
	interceptors := connect.WithInterceptors(interceptors.NewRateLimitInterceptor(key))

	return userv1connect.NewAuthServiceHandler(
		&AuthHandler{
			db:  db,
			key: []byte(key),
		},
		interceptors,
	)
}
