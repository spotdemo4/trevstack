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

func (h *UserHandler) GetUser(ctx context.Context, req *connect.Request[userv1.GetUserRequest]) (*connect.Response[userv1.GetUserResponse], error) {
	userid, ok := interceptors.GetUserContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
	}

	// Get user
	user := models.User{}
	if err := h.db.Preload("ProfilePicture").First(&user, "id = ?", userid).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&userv1.GetUserResponse{
		User: user.ToConnectV1(),
	})
	return res, nil
}

func (h *UserHandler) UpdatePassword(ctx context.Context, req *connect.Request[userv1.UpdatePasswordRequest]) (*connect.Response[userv1.UpdatePasswordResponse], error) {
	userid, ok := interceptors.GetUserContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
	}

	// Get user
	user := models.User{}
	if err := h.db.First(&user, "id = ?", userid).Error; err != nil {
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
	if err := h.db.Model(&user).Update("password", string(hash)).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&userv1.UpdatePasswordResponse{})
	return res, nil
}

func (h *UserHandler) GetAPIKey(ctx context.Context, req *connect.Request[userv1.GetAPIKeyRequest]) (*connect.Response[userv1.GetAPIKeyResponse], error) {
	userid, ok := interceptors.GetUserContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
	}

	// Get user
	user := models.User{}
	if err := h.db.First(&user, "id = ?", userid).Error; err != nil {
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
	ss, err := t.SignedString(h.key)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&userv1.GetAPIKeyResponse{
		Key: ss,
	})
	return res, nil
}

func (h *UserHandler) UpdateProfilePicture(ctx context.Context, req *connect.Request[userv1.UpdateProfilePictureRequest]) (*connect.Response[userv1.UpdateProfilePictureResponse], error) {
	userid, ok := interceptors.GetUserContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
	}

	// Validate file
	fileType := http.DetectContentType(req.Msg.Data)
	if fileType != "image/jpeg" && fileType != "image/png" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid file type"))
	}

	// Save bytes into file
	file := models.File{
		Name:   req.Msg.FileName,
		Data:   req.Msg.Data,
		UserID: uint(userid),
	}
	if err := h.db.Create(&file).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Get user info
	user := models.User{}
	if err := h.db.First(&user, "id = ?", userid).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Get old profile picture ID
	var ppid *uint32
	if user.ProfilePicture != nil {
		ppid = &user.ProfilePicture.ID
	}

	// Update user profile picture
	fid := uint(file.ID)
	user.ProfilePictureID = &fid
	if err := h.db.Save(&user).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Delete old profile picture if exists
	if ppid != nil {
		if err := h.db.Delete(models.File{}, "id = ?", *ppid).Error; err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	res := connect.NewResponse(&userv1.UpdateProfilePictureResponse{
		User: user.ToConnectV1(),
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
