package order

import (
	"context"
	"database/sql"
	"errors"
	"github.com/romanp1989/gophermart/internal/domain"
)

type DBStorage struct {
	db *sql.DB
}

var (
	insertCreateOrderQuery = `INSERT INTO orders (number, status, user_id, created_at) VALUES ($1, $2, $3, $4) RETURNING id;`

	selectLoadOrderQuery = `SELECT o.id, o.number, o.status, o.user_id, o.created_at 
							FROM orders o 
							WHERE o.number = $1;`

	selectLoadOrdersWithBalanceQuery = `SELECT o.id, o.number, o.status, o.user_id, o.created_at, b.sum 
										FROM orders o 
										LEFT JOIN balance b on o.number = b.order_number AND b.type = $1 
										WHERE o.user_id = $2 
										ORDER BY o.created_at DESC;`

	updateLoadOrdersToProcessQuery = `UPDATE orders o SET status = $1
									WHERE o.id IN (
											SELECT id FROM orders
											WHERE status = $2 ORDER BY id LIMIT 10
										) AND status = $2
									returning o.id, o.number, o.status, o.user_id, o.created_at;`
)

func NewDBStorage(db *sql.DB) *DBStorage {
	return &DBStorage{
		db: db,
	}
}

func (d *DBStorage) CreateOrder(ctx context.Context, order *domain.Order) (int64, error) {
	var id int64

	err := d.db.QueryRowContext(
		ctx,
		insertCreateOrderQuery,
		order.Number, order.Status, order.UserID, order.CreatedAt,
	).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (d *DBStorage) LoadOrder(ctx context.Context, orderNumber string) (*domain.Order, error) {
	orderLoaded := &domain.Order{}

	err := d.db.QueryRowContext(
		ctx,
		selectLoadOrderQuery,
		orderNumber).
		Scan(&orderLoaded.ID, &orderLoaded.Number, &orderLoaded.Status, &orderLoaded.UserID, &orderLoaded.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFoundOrder
		}

		return nil, err
	}

	return orderLoaded, nil
}

func (d *DBStorage) LoadOrdersWithBalance(ctx context.Context, userID domain.UserID) ([]domain.OrderWithBalance, error) {
	rows, err := d.db.QueryContext(
		ctx,
		selectLoadOrdersWithBalanceQuery,
		domain.BalanceTypeAdded,
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var orders []domain.OrderWithBalance
	for rows.Next() {
		var balance sql.NullFloat64

		orderWithBalance := domain.OrderWithBalance{
			Order: domain.Order{},
		}

		err = rows.Scan(&orderWithBalance.ID, &orderWithBalance.Number, &orderWithBalance.Status, &orderWithBalance.UserID, &orderWithBalance.CreatedAt, &balance)
		if err != nil {
			return nil, err
		}

		if balance.Valid {
			orderWithBalance.Balance = balance.Float64
		}

		orders = append(orders, orderWithBalance)
	}

	return orders, nil
}

func (d *DBStorage) LoadOrdersToProcess(ctx context.Context) ([]domain.Order, error) {
	rows, err := d.db.QueryContext(
		ctx,
		updateLoadOrdersToProcessQuery,
		domain.OrderStatusProcessing, domain.OrderStatusNew)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var orders []domain.Order
	for rows.Next() {
		o := domain.Order{}

		err = rows.Scan(&o.ID, &o.Number, &o.Status, &o.UserID, &o.CreatedAt)
		if err != nil {
			return nil, err
		}

		orders = append(orders, o)
	}

	return orders, nil
}
