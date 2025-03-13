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

type UserHandler struct {
	db  *gorm.DB
	key []byte
}

func (s *UserHandler) ChangePassword(ctx context.Context, req *connect.Request[userv1.ChangePasswordRequest]) (*connect.Response[userv1.ChangePasswordResponse], error) {
	userid, ok := interceptors.UserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
	}

	// Get user
	user := models.User{}
	if err := s.db.First(&user, "id = ?", userid).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Validate
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Msg.OldPassword)); err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("invalid password"))
	}
	if req.Msg.NewPassword != req.Msg.ConfirmPassword {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("passwords do not match"))
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Msg.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Update password
	if err := s.db.Model(&user).Update("password", string(hash)).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&userv1.ChangePasswordResponse{})
	return res, nil
}

func (s *UserHandler) APIKey(ctx context.Context, req *connect.Request[userv1.APIKeyRequest]) (*connect.Response[userv1.APIKeyResponse], error) {
	userid, ok := interceptors.UserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
	}

	// Get user
	user := models.User{}
	if err := s.db.First(&user, "id = ?", userid).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Validate
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Msg.Password)); err != nil {
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("invalid username or password"))
	}
	if req.Msg.Password != req.Msg.ConfirmPassword {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("passwords do not match"))
	}

	// Generate JWT
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:  "trevstack",
		Subject: strconv.FormatUint(uint64(user.ID), 10),
		IssuedAt: &jwt.NumericDate{
			Time: time.Now(),
		},
	})
	ss, err := t.SignedString(s.key)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&userv1.APIKeyResponse{
		Key: ss,
	})
	return res, nil
}

func NewUserHandler(db *gorm.DB, key string) (string, http.Handler) {
	interceptors := connect.WithInterceptors(interceptors.NewAuthInterceptor(key))

	return userv1connect.NewUserServiceHandler(
		&UserHandler{
			db:  db,
			key: []byte(key),
		},
		interceptors,
	)
}
