package repository

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"strconv"
	"time"

	pb "currency-save-service/internal/generated"
	"currency-save-service/internal/loger"
	"currency-save-service/internal/metrics"
	"currency-save-service/internal/models"
)

type Repository struct {
	Db      *sql.DB
	Logerr  *slog.Logger
	Metrics *metrics.Metrics
}

func NewRepository(MysqlConnectionString string, logerr *loger.Logerr, metrics *metrics.Metrics) *Repository {
	db, err := sql.Open("mysql", MysqlConnectionString)
	if err != nil {
		logerr.Logerr.Error("Failed initialize database connection")
		return nil
	}

	db.SetMaxOpenConns(39)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(3 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil
	}

	return &Repository{
		Db:      db,
		Logerr:  logerr.Logerr,
		Metrics: metrics,
	}
}

func (r *Repository) InsertData(rates models.Rates, formattedDate string) {
	savedItemCount := 0

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*30))
	defer cancel()

	for _, item := range rates.Items {
		value, err := strconv.ParseFloat(item.Value, 64)
		if err != nil {
			r.Logerr.Error("Failed to convert float: %s", err)
			continue
		}

		startTime := time.Now()

		rows, err := r.Db.QueryContext(ctx, "INSERT INTO R_CURRENCY (TITLE, CODE, VALUE, A_DATE) VALUES (?, ?, ?, ?)", item.Title, item.Code, value, formattedDate)
		if err != nil {
			r.Logerr.Error("Failed to insert in the database:", err)
		} else {
			savedItemCount++
			r.Logerr.Info("Item saved",
				"count", savedItemCount,
			)
		}
		defer rows.Close()

		duration := time.Since(startTime).Seconds()
		go r.Metrics.ObserveInsertDuration("insert", "success", duration)
	}
	r.Logerr.Info("Items saved:",
		"All", savedItemCount,
	)
}

func (r *Repository) SaveCurrency(ctx context.Context, currency *pb.SaveCurrencyRequest) (bool, error) {
	stmt, err := r.Db.PrepareContext(ctx, "INSERT INTO R_CURRENCY (TITLE, CODE, VALUE, A_DATE) VALUES (?, ?, ?, ?)")
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	// Execute the SQL statement with currency data
	result, err := stmt.ExecContext(ctx, currency.Title, currency.Code, currency.Value, currency.Date)
	if err != nil {
		return false, err
	}

	// Check the number of rows affected by the query
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	// Log the success message
	log.Printf("Inserted %d rows into R_CURRENCY table", rowsAffected)

	return true, nil
}
