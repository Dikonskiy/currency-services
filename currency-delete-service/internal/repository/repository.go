package repository

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	pb "currency-delete-service/internal/generated"
	"currency-delete-service/internal/loger"
	"currency-delete-service/internal/metrics"
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

func (r *Repository) DeleteData(ctx context.Context, formattedDate, code string) (int64, error) {
	var query string
	var params []interface{}

	if code == "" {
		query = "DELETE FROM R_CURRENCY WHERE A_DATE = ?"
		params = []interface{}{formattedDate}
	} else {
		query = "DELETE FROM R_CURRENCY WHERE A_DATE = ? AND CODE = ?"
		params = []interface{}{formattedDate, code}
	}

	startTime := time.Now()

	result, err := r.Db.ExecContext(ctx, query, params...)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	duration := time.Since(startTime).Seconds()
	if code == "" {
		go r.Metrics.ObserveDeleteDuration("delete", "success", duration)
		go r.Metrics.IncDeleteCount("delete", "success")
	}

	if rowsAffected == 0 {
		r.Logerr.Error("No data deleted with these parameters")
	}

	return rowsAffected, nil
}

func (r *Repository) DeleteCurrency(ctx context.Context, req *pb.DeleteCurrencyRequest) (bool, error) {
	query := "DELETE FROM R_CURRENCY WHERE A_DATE = ? AND CODE = ?"
	result, err := r.Db.ExecContext(ctx, query, req.Date, req.Code)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	success := rowsAffected > 0

	return success, nil
}
