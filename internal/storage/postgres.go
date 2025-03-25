package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // Импорт драйвера для работы с PostgreSQL
)

var (
	ErrUserNotFound = errors.New("пользователь не найден")
	ErrBalanceZero  = errors.New("не хватает баланса, чтобы списать средства")
)

type Order struct {
	Number   string    `json:"number"`
	Status   string    `json:"status"`
	Accrual  float64   `json:"accrual,omitempty"`
	Uploaded time.Time `json:"uploaded_at"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Withdraw struct {
	OrderNum  string    `json:"order"`
	Sum       float64   `json:"sum"`
	Processed time.Time `json:"processed_at"`
}

func CreatePostgreStorage(dsn string) (*PostgresStorage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	storage := &PostgresStorage{db: db}

	err = storage.createTables()
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *PostgresStorage) createTables() error {
	query := `
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    login VARCHAR(255) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    balance DECIMAL(10,2) NOT NULL DEFAULT 0.00 CHECK (balance >= 0),
    withdrawn DECIMAL(10,2) NOT NULL DEFAULT 0 CHECK (withdrawn >= 0)
);	
	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		author_id INT NOT NULL,
		order_num TEXT UNIQUE NOT NULL,
        uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL,
		FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
	);

CREATE TABLE IF NOT EXISTS withdrawals (
    id SERIAL PRIMARY KEY,
    author_id INT NOT NULL,
    order_num TEXT UNIQUE NOT NULL,
    quantity DECIMAL(10,2) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE   )
;`
	_, err := s.db.Exec(query)
	return err
}

func (s *PostgresStorage) CreateUser(ctx context.Context, login string, password string) error {
	rows, err := s.db.QueryContext(ctx, `
	INSERT INTO users (login, password)
	VALUES ($1, $2)
	`, login, password)
	if err != nil {
		return err
	}
	if rows.Err() != nil {
		return rows.Err()
	}
	return nil
}

func (s *PostgresStorage) IsUserExist(ctx context.Context, login string) (bool, error) {
	var id int
	err := s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE login = $1", login).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *PostgresStorage) GetPasswordFromLogin(ctx context.Context, login string) (string, error) {
	var password string
	err := s.db.QueryRowContext(ctx, "SELECT password FROM users WHERE login = $1", login).Scan(&password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}
	return password, nil

}

func (s *PostgresStorage) GetID(ctx context.Context, login string) (int, error) {
	var id int
	err := s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE login = $1", login).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *PostgresStorage) IsOrderExists(ctx context.Context, orderID string) (bool, error) {
	var id int
	err := s.db.QueryRowContext(ctx, "SELECT id FROM orders WHERE order_num = $1", orderID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil

}

func (s *PostgresStorage) GetAuthorOrder(ctx context.Context, orderID string) (int, error) {
	var authorID int
	err := s.db.QueryRowContext(ctx, "SELECT author_id FROM orders WHERE order_num = $1", orderID).Scan(&authorID)
	if err != nil {
		return 0, err
	}
	return authorID, nil
}

func (s *PostgresStorage) CreateOrder(ctx context.Context, authorID int, orderID string, uploaded time.Time) error {
	_, err := s.db.ExecContext(ctx, "INSERT INTO orders (author_id, order_num, uploaded_at) VALUES ($1, $2, $3)", authorID, orderID, uploaded)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) GetOrdersFromUser(ctx context.Context, userID int) ([]Order, error) {
	var orders []Order
	rows, err := s.db.QueryContext(ctx, "SELECT order_num, uploaded_at FROM orders WHERE author_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			orderNum string
			uploaded time.Time
		)
		if err := rows.Scan(&orderNum, &uploaded); err != nil {
			return nil, err
		}
		orders = append(orders, Order{Number: orderNum, Uploaded: uploaded})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (s *PostgresStorage) GetBalance(ctx context.Context, userID int) (*Balance, error) {
	var (
		balance   float64
		withdrawn float64
	)
	err := s.db.QueryRowContext(ctx, "SELECT  balance, withdrawn  FROM users WHERE id = $1", userID).Scan(&balance, &withdrawn)
	if err != nil {
		return nil, err
	}
	return &Balance{Current: balance, Withdrawn: withdrawn}, nil

}

func (s *PostgresStorage) AddBalance(ctx context.Context, userID int, accrual float64) error {
	query := `UPDATE users SET balance = balance + $1 WHERE id = $2`

	_, err := s.db.ExecContext(ctx, query, accrual, userID)
	if err != nil {
		return fmt.Errorf("не удалось обновить баланс пользователя %d: %w", userID, err)
	}

	return nil
}

func (s *PostgresStorage) WithdrawBalance(ctx context.Context, userID int, sum float64) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE users SET balance = balance - $1, withdrawn = withdrawn + $1 WHERE id = $2",
		sum, userID)

	if err != nil {
		if errors.Is(err, sql.ErrTxDone) {
			return ErrBalanceZero
		}

		return err
	}
	return nil
}

func (s *PostgresStorage) CreateWithdraw(ctx context.Context, userID int, orderID string, sum float64, processed time.Time) error {
	err := s.db.QueryRowContext(ctx, "INSERT INTO withdrawals(author_id, order_num, quantity, processed_at) VALUES ($1, $2, $3, $4)", userID, orderID, sum, processed)
	if err != nil {
		return err.Err()
	}
	return nil
}

func (s *PostgresStorage) GetWithdraws(ctx context.Context, userID int) ([]Withdraw, error) {
	var orders []Withdraw

	rows, err := s.db.QueryContext(ctx, "SELECT order_num, quantity, processed_at FROM withdrawals WHERE author_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order Withdraw
		err := rows.Scan(&order.OrderNum, &order.Sum, &order.Processed)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
