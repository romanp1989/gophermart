package accrual

import (
	"database/sql"
	"github.com/romanp1989/gophermart/internal/domain"
)

type DBStorage struct {
	db *sql.DB
}

// Лучше использовать const для запросов, но использовал var чтобы было удобнее скрывать блок с запросами при редактировании кода
var (
	getNewOrdersToSendQuery = `UPDATE orders o SET status = $1
								WHERE o.id IN (
										SELECT id FROM orders
										WHERE status = $2 ORDER BY id LIMIT 10
									) AND status = $2
								returning o.id, o.number, o.status, o.user_id, o.created_at;`

	addBalanceQuery = `INSERT INTO balance (order_number, sum, type, user_id, created_at) VALUES ($1, $2, $3, $4, $5);`

	updateOrderQuery = `UPDATE orders SET status = $1 WHERE user_id = $2 AND number = $3`

	updateQuery = `UPDATE orders 
		SET status = $1 
		WHERE id = $2;`
)

func NewDBStorage(db *sql.DB) *DBStorage {
	return &DBStorage{
		db: db,
	}
}

func (d *DBStorage) GetNewOrdersToSend() ([]domain.Order, error) {
	orders := make([]domain.Order, 0)

	rows, err := d.db.Query(getNewOrdersToSendQuery, domain.OrderStatusProcessing, domain.OrderStatusNew)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var order domain.Order

		err = rows.Scan(&order.ID, &order.Number, &order.Status, &order.UserID, &order.CreatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (d *DBStorage) AddBalance(o *domain.Balance) error {
	_, err := d.db.Exec(addBalanceQuery, o.OrderNumber, o.Sum, o.Type, o.UserID, o.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (d *DBStorage) UpdateOrder(order *domain.AccrualResponse, userID domain.UserID) error {
	_, err := d.db.Exec(updateOrderQuery, order.Status, userID, order.Order)
	if err != nil {
		return err
	}
	return nil
}

func (d *DBStorage) Update(o domain.Order) error {
	_, err := d.db.Exec(updateQuery, o.Status, o.ID)

	if err != nil {
		return err
	}

	return nil
}
