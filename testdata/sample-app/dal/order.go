package dal

import "context"

type Order struct {
	ID        string
	UserID    int64
	ProductID int64
	Quantity  int32
	Address   string
	Status    string
}

type OrderDAL struct{}

func NewOrderDAL() *OrderDAL {
	return &OrderDAL{}
}

func (d *OrderDAL) Create(ctx context.Context, order *Order) error {
	// stub: insert into database
	return nil
}

func (d *OrderDAL) GetByID(ctx context.Context, id string) (*Order, error) {
	// stub: query from database
	return &Order{ID: id, Status: "created"}, nil
}
