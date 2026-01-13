package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Bucheli05/stock-app-backend/internal/models"
	_ "github.com/lib/pq"
)

type StockRepository struct {
	db *sql.DB
}

func NewStockRepository(db *sql.DB) *StockRepository {
	return &StockRepository{db: db}
}

// InitDB initializes the database connection
func InitDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// CreateTable creates the stocks table if it doesn't exist
func (r *StockRepository) CreateTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS stocks (
			id SERIAL PRIMARY KEY,
			ticker VARCHAR(20) NOT NULL,
			target_from VARCHAR(50),
			target_to VARCHAR(50),
			company VARCHAR(255) NOT NULL,
			action VARCHAR(100),
			brokerage VARCHAR(255),
			rating_from VARCHAR(50),
			rating_to VARCHAR(50),
			time TIMESTAMP NOT NULL,
			fetched_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT NOW(),
			INDEX idx_ticker (ticker),
			INDEX idx_fetched_at (fetched_at),
			INDEX idx_time (time)
		)
	`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

// SaveStocks inserts multiple stocks into the database
func (r *StockRepository) SaveStocks(ctx context.Context, stocks []models.StockItem, fetchedAt time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO stocks (ticker, target_from, target_to, company, action, brokerage, rating_from, rating_to, time, fetched_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, stock := range stocks {
		_, err = stmt.ExecContext(ctx,
			stock.Ticker,
			stock.TargetFrom,
			stock.TargetTo,
			stock.Company,
			stock.Action,
			stock.Brokerage,
			stock.RatingFrom,
			stock.RatingTo,
			stock.Time,
			fetchedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetLatestStocks retrieves the most recent stocks based on fetched_at
func (r *StockRepository) GetLatestStocks(ctx context.Context) ([]models.StockItem, error) {
	query := `
		SELECT id, ticker, target_from, target_to, company, action, brokerage,
		       rating_from, rating_to, time, fetched_at, created_at
		FROM stocks
		WHERE fetched_at = (SELECT MAX(fetched_at) FROM stocks)
		ORDER BY time DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []models.StockItem
	for rows.Next() {
		var stock models.StockItem
		err := rows.Scan(
			&stock.ID,
			&stock.Ticker,
			&stock.TargetFrom,
			&stock.TargetTo,
			&stock.Company,
			&stock.Action,
			&stock.Brokerage,
			&stock.RatingFrom,
			&stock.RatingTo,
			&stock.Time,
			&stock.FetchedAt,
			&stock.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		stocks = append(stocks, stock)
	}

	return stocks, rows.Err()
}

// GetStocksByDate retrieves stocks for a specific date
func (r *StockRepository) GetStocksByDate(ctx context.Context, date time.Time) ([]models.StockItem, error) {
	query := `
		SELECT id, ticker, target_from, target_to, company, action, brokerage,
		       rating_from, rating_to, time, fetched_at, created_at
		FROM stocks
		WHERE DATE(fetched_at) = DATE($1)
		ORDER BY time DESC
	`

	rows, err := r.db.QueryContext(ctx, query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []models.StockItem
	for rows.Next() {
		var stock models.StockItem
		err := rows.Scan(
			&stock.ID,
			&stock.Ticker,
			&stock.TargetFrom,
			&stock.TargetTo,
			&stock.Company,
			&stock.Action,
			&stock.Brokerage,
			&stock.RatingFrom,
			&stock.RatingTo,
			&stock.Time,
			&stock.FetchedAt,
			&stock.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		stocks = append(stocks, stock)
	}

	return stocks, rows.Err()
}

// DeleteOldRecords deletes records older than the specified number of days
func (r *StockRepository) DeleteOldRecords(ctx context.Context, daysToKeep int) error {
	query := `
		DELETE FROM stocks
		WHERE fetched_at < NOW() - INTERVAL '1 day' * $1
	`
	_, err := r.db.ExecContext(ctx, query, daysToKeep)
	return err
}
