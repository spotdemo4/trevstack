package v1

import (
	"context"
	_ "embed"

	numberv1 "trev.zip/trev/stack/server/connect/number/v1"
	"trev.zip/trev/stack/server/database"
)

//go:embed add.sql
var addSQL string

//go:embed sum.sql
var sumSQL string

func (h *Handler) Add(
	ctx context.Context,
	req *numberv1.AddRequest,
) (*numberv1.AddResponse, error) {
	db := database.FromContext(ctx)

	// Add the number to the database.
	_, err := db.ExecContext(ctx, addSQL, req.GetName(), req.GetNumber())
	if err != nil {
		return nil, err
	}

	// Calculate the sum of all numbers in the database.
	var sum uint64
	err = db.QueryRowContext(ctx, sumSQL).Scan(&sum)
	if err != nil {
		return nil, err
	}

	response := &numberv1.AddResponse{}
	response.SetSum(sum)

	return response, nil
}
