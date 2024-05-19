package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	pb "currency-get-service/internal/generated"
	"currency-get-service/internal/loger"
	"currency-get-service/internal/metrics"
	"currency-get-service/internal/models"
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

func (r *Repository) GetData(ctx context.Context, formattedDate, code string) ([]models.DBItem, error) {
	var query string
	var params []interface{}

	if code == "" {
		query = "SELECT ID, TITLE, CODE, VALUE, A_DATE FROM R_CURRENCY WHERE A_DATE = ?"
		params = []interface{}{formattedDate}
	} else {
		query = "SELECT ID, TITLE, CODE, VALUE, A_DATE FROM R_CURRENCY WHERE A_DATE = ? AND CODE = ?"
		params = []interface{}{formattedDate, code}
	}

	startTime := time.Now()

	rows, err := r.Db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	duration := time.Since(startTime).Seconds()
	if code == "" {
		go r.Metrics.ObserveSelectDuration("select", "success", duration)
		go r.Metrics.IncSelectCount("select", "success")
	}

	var results []models.DBItem
	for rows.Next() {
		var item models.DBItem
		if err := rows.Scan(&item.ID, &item.Title, &item.Code, &item.Value, &item.Date); err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	if len(results) == 0 {
		r.Logerr.Error("No data found with these parameters")
	}

	return results, nil
}

func (r *Repository) GetCurrency(ctx context.Context, req *pb.GetCurrencyRequest) (*pb.GetCurrencyResponse, error) {
	query := "SELECT A_DATE, CODE, VALUE FROM R_CURRENCY WHERE A_DATE = ? AND CODE = ?"
	row := r.Db.QueryRowContext(ctx, query, req.Date, req.Code)

	var date string
	var code string
	var rate float64

	err := row.Scan(&date, &code, &rate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("currency data not found")
		}
		return nil, err
	}

	response := &pb.GetCurrencyResponse{
		Date: date,
		Code: code,
		Rate: rate,
	}

	return response, nil
}
