package item

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/aarondl/opt/omit"
	"github.com/aarondl/opt/omitnull"
	"github.com/spotdemo4/trevstack/server/internal/interceptors"
	"github.com/spotdemo4/trevstack/server/internal/models"
	itemv1 "github.com/spotdemo4/trevstack/server/internal/services/item/v1"
	"github.com/spotdemo4/trevstack/server/internal/services/item/v1/itemv1connect"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/sqlite"
	"github.com/stephenafamo/bob/dialect/sqlite/sm"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func itemToConnect(item *models.Item) *itemv1.Item {
	timestamp := timestamppb.New(item.Added)

	return &itemv1.Item{
		Id:          &item.ID,
		Name:        item.Name,
		Description: item.Description.GetOrZero(),
		Price:       item.Price.GetOrZero(),
		Quantity:    int32(item.Quantity.GetOrZero()),
		Added:       timestamp,
	}
}

type Handler struct {
	db  *bob.DB
	key []byte
}

func (h *Handler) GetItem(ctx context.Context, req *connect.Request[itemv1.GetItemRequest]) (*connect.Response[itemv1.GetItemResponse], error) {
	userid, ok := interceptors.GetUserContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
	}

	// Get item
	item, err := models.Items.Query(
		sqlite.WhereAnd(
			models.SelectWhere.Items.ID.EQ(req.Msg.Id),
			models.SelectWhere.Items.UserID.EQ(userid),
		),
	).One(ctx, h.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&itemv1.GetItemResponse{
		Item: itemToConnect(item),
	})
	return res, nil
}

func (h *Handler) GetItems(ctx context.Context, req *connect.Request[itemv1.GetItemsRequest]) (*connect.Response[itemv1.GetItemsResponse], error) {
	userid, ok := interceptors.GetUserContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
	}

	// Filters
	query := models.Items.Query(models.SelectWhere.Items.UserID.EQ(userid))
	countQuery := models.Items.Query(models.SelectWhere.Items.UserID.EQ(userid))

	// Counted filters
	if req.Msg.Start != nil {
		query.Apply(models.SelectWhere.Items.Added.GTE(req.Msg.Start.AsTime()))
		countQuery.Apply(models.SelectWhere.Items.Added.GTE(req.Msg.Start.AsTime()))
	}
	if req.Msg.End != nil {
		query.Apply(models.SelectWhere.Items.Added.LTE(req.Msg.End.AsTime()))
		countQuery.Apply(models.SelectWhere.Items.Added.LTE(req.Msg.End.AsTime()))
	}
	if req.Msg.Filter != nil && *req.Msg.Filter != "" {
		query.Apply(models.SelectWhere.Items.Name.Like("%" + *req.Msg.Filter + "%"))
		countQuery.Apply(models.SelectWhere.Items.Name.Like(*req.Msg.Filter))
	}

	// Uncounted filters
	if req.Msg.Limit != nil {
		query.Apply(sm.Limit(*req.Msg.Limit))
	}
	if req.Msg.Offset != nil {
		query.Apply(sm.Offset(*req.Msg.Offset))
	}

	// Get items & count
	items, err := query.All(ctx, h.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}

		return nil, connect.NewError(connect.CodeInternal, err)
	}

	count, err := query.Count(ctx, h.db)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Convert to connect v1 items
	resItems := []*itemv1.Item{}
	for _, item := range items {
		resItems = append(resItems, itemToConnect(item))
	}

	res := connect.NewResponse(&itemv1.GetItemsResponse{
		Items: resItems,
		Count: count,
	})
	return res, nil
}

func (h *Handler) CreateItem(ctx context.Context, req *connect.Request[itemv1.CreateItemRequest]) (*connect.Response[itemv1.CreateItemResponse], error) {
	userid, ok := interceptors.GetUserContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
	}

	item, err := models.Items.Insert(&models.ItemSetter{
		Name:        omit.From(req.Msg.Item.Name),
		Description: omitnull.From(req.Msg.Item.Description),
		Price:       omitnull.From(req.Msg.Item.Price),
		Quantity:    omitnull.From(int64(req.Msg.Item.Quantity)),
		Added:       omit.From(time.Now()),
		UserID:      omit.From(userid),
	}).One(ctx, h.db)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&itemv1.CreateItemResponse{
		Item: itemToConnect(item),
	})
	return res, nil
}

func (h *Handler) UpdateItem(ctx context.Context, req *connect.Request[itemv1.UpdateItemRequest]) (*connect.Response[itemv1.UpdateItemResponse], error) {
	userid, ok := interceptors.GetUserContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
	}

	// Validate
	if req.Msg.Item.Id == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("id is required"))
	}

	// Update item
	item, err := models.Items.Update(
		// Set col
		models.ItemSetter{
			Name:        omit.From(req.Msg.Item.Name),
			Description: omitnull.From(req.Msg.Item.Description),
			Price:       omitnull.From(req.Msg.Item.Price),
			Quantity:    omitnull.From(int64(req.Msg.Item.Quantity)),
		}.UpdateMod(),

		// Where
		sqlite.WhereAnd(
			models.UpdateWhere.Items.ID.EQ(*req.Msg.Item.Id),
			models.UpdateWhere.Items.UserID.EQ(userid),
		),
	).One(ctx, h.db)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&itemv1.UpdateItemResponse{
		Item: itemToConnect(item),
	})
	return res, nil
}

func (h *Handler) DeleteItem(ctx context.Context, req *connect.Request[itemv1.DeleteItemRequest]) (*connect.Response[itemv1.DeleteItemResponse], error) {
	userid, ok := interceptors.GetUserContext(ctx)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
	}

	// Delete item
	_, err := models.Items.Delete(
		sqlite.WhereAnd(
			models.DeleteWhere.Items.ID.EQ(req.Msg.Id),
			models.DeleteWhere.Items.UserID.EQ(userid),
		),
	).Exec(ctx, h.db)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&itemv1.DeleteItemResponse{})
	return res, nil
}

func NewHandler(db *bob.DB, key string) (string, http.Handler) {
	interceptors := connect.WithInterceptors(interceptors.NewAuthInterceptor(key))

	return itemv1connect.NewItemServiceHandler(
		&Handler{
			db:  db,
			key: []byte(key),
		},
		interceptors,
	)
}
