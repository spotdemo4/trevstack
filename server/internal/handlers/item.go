package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/spotdemo4/trevstack/server/internal/interceptors"
	"github.com/spotdemo4/trevstack/server/internal/models"
	itemv1 "github.com/spotdemo4/trevstack/server/internal/services/item/v1"
	"github.com/spotdemo4/trevstack/server/internal/services/item/v1/itemv1connect"
	"gorm.io/gorm"
)

type ItemHandler struct {
	db  *gorm.DB
	key []byte
}

func (h *ItemHandler) GetItem(ctx context.Context, req *connect.Request[itemv1.GetItemRequest]) (*connect.Response[itemv1.GetItemResponse], error) {
	userid, ok := interceptors.UserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
	}

	// Get item
	item := models.Item{}
	if err := h.db.First(&item, "id = ? AND user_id = ?", req.Msg.Id, userid).Error; err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}

	res := connect.NewResponse(&itemv1.GetItemResponse{
		Item: item.ToConnectV1(),
	})
	return res, nil
}

func (h *ItemHandler) GetItems(ctx context.Context, req *connect.Request[itemv1.GetItemsRequest]) (*connect.Response[itemv1.GetItemsResponse], error) {
	userid, ok := interceptors.UserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
	}

	// Filters
	sql := h.db.Where("user_id = ?", userid)
	if req.Msg.Start != nil {
		sql = sql.Where("added >= ?", req.Msg.Start.AsTime())
	}
	if req.Msg.End != nil {
		sql = sql.Where("added <= ?", req.Msg.End.AsTime())
	}
	if req.Msg.Filter != nil {
		sql = sql.Where("name LIKE ?", fmt.Sprintf("%%%s%%", *req.Msg.Filter))
	}

	// Uncounted filters
	sqlu := sql.Session(&gorm.Session{})
	if req.Msg.Limit != nil {
		sqlu = sqlu.Limit(int(*req.Msg.Limit))
	}
	if req.Msg.Offset != nil {
		sqlu = sqlu.Offset(int(*req.Msg.Offset))
	}

	// Get items & count
	items := []models.Item{}
	var count int64
	if err := sqlu.Order("added desc").Find(&items).Error; err != nil {
		return nil, connect.NewError(connect.CodeNotFound, err)
	}
	if err := sql.Model(&items).Count(&count).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert to connect v1 items
	resItems := []*itemv1.Item{}
	for _, item := range items {
		resItems = append(resItems, item.ToConnectV1())
	}

	res := connect.NewResponse(&itemv1.GetItemsResponse{
		Items: resItems,
		Count: uint64(count),
	})
	return res, nil
}

func (h *ItemHandler) CreateItem(ctx context.Context, req *connect.Request[itemv1.CreateItemRequest]) (*connect.Response[itemv1.CreateItemResponse], error) {
	userid, ok := interceptors.UserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
	}

	// Create item
	item := models.Item{
		Name:        req.Msg.Item.Name,
		Description: req.Msg.Item.Description,
		Price:       req.Msg.Item.Price,
		Quantity:    int(req.Msg.Item.Quantity),
		Added:       time.Now(),
		UserID:      uint(userid),
	}
	if err := h.db.Create(&item).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&itemv1.CreateItemResponse{
		Item: item.ToConnectV1(),
	})
	return res, nil
}

func (h *ItemHandler) UpdateItem(ctx context.Context, req *connect.Request[itemv1.UpdateItemRequest]) (*connect.Response[itemv1.UpdateItemResponse], error) {
	userid, ok := interceptors.UserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
	}

	// Validate
	if req.Msg.Item.Id == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("id is required"))
	}

	// Update item
	item := models.Item{
		ID:          *req.Msg.Item.Id,
		Name:        req.Msg.Item.Name,
		Description: req.Msg.Item.Description,
		Price:       req.Msg.Item.Price,
		Quantity:    int(req.Msg.Item.Quantity),
		UserID:      uint(userid),
	}
	if err := h.db.Where("id = ? AND user_id = ?", req.Msg.Item.Id, userid).Updates(&item).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&itemv1.UpdateItemResponse{
		Item: item.ToConnectV1(),
	})
	return res, nil
}

func (h *ItemHandler) DeleteItem(ctx context.Context, req *connect.Request[itemv1.DeleteItemRequest]) (*connect.Response[itemv1.DeleteItemResponse], error) {
	userid, ok := interceptors.UserFromContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
	}

	// Delete item
	if err := h.db.Delete(&models.Item{}, "id = ? AND user_id = ?", req.Msg.Id, userid).Error; err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&itemv1.DeleteItemResponse{})
	return res, nil
}

func NewItemHandler(db *gorm.DB, key string) (string, http.Handler) {
	interceptors := connect.WithInterceptors(interceptors.NewAuthInterceptor(key))

	return itemv1connect.NewItemServiceHandler(
		&ItemHandler{
			db:  db,
			key: []byte(key),
		},
		interceptors,
	)
}
