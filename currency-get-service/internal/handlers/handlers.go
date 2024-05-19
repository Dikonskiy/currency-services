package handlers

import (
	"context"
	pb "currency-get-service/internal/generated"
	"currency-get-service/internal/loger"
	"currency-get-service/internal/repository"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Handler struct {
	R     *repository.Repository
	loger *loger.Logerr
}

type CurrencyServiceServer struct {
	Repo *repository.Repository
	pb.UnimplementedCurrencyServiceServer
}

func NewHandler(repo *repository.Repository, loger *loger.Logerr) *Handler {
	return &Handler{
		R:     repo,
		loger: loger,
	}
}

func DateFormat(date string) (string, error) {
	parsedDate, err := time.Parse("02.01.2006", date)
	if err != nil {
		return "", err
	}
	formattedDate := parsedDate.Format("2006-01-02")
	return formattedDate, nil
}

func (h *Handler) RespondWithError(w http.ResponseWriter, status int, errorMsg string, err error) {
	http.Error(w, errorMsg, status)
	h.R.Logerr.Error(errorMsg+": ", err)
}

// @Summary Get currency data by date
// @Description Get currency data for a specific date.
// @Tags currency
// @Accept json
// @Param date path string true "Date in DD.MM.YYYY format"
// @Router /currency/{date} [get]
// @Summary Get currency data by date and code
// @Description Get currency data for a specific date and currency code.
// @Tags currency
// @Accept json
// @Param code path string true "Currency code (e.g., USD)"
// @Router /currency/{date}/{code} [get]
func (h *Handler) GetCurrencyHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	vars := mux.Vars(r)
	date := vars["date"]
	code := vars["code"]

	formattedDate, err := DateFormat(date)
	if err != nil {
		h.RespondWithError(w, http.StatusBadRequest, "Failed to parse the date", err)
		return
	}

	data, err := h.R.GetData(ctx, formattedDate, code)
	if err != nil {
		h.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve data", err)
		return
	}
	h.R.Logerr.Info("Data was showed")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *CurrencyServiceServer) GetCurrency(ctx context.Context, req *pb.GetCurrencyRequest) (*pb.GetCurrencyResponse, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}

	currencyData, err := s.Repo.GetCurrency(ctx, req)
	if err != nil {
		return nil, err
	}

	response := &pb.GetCurrencyResponse{
		Date: currencyData.Date,
		Code: currencyData.Code,
		Rate: currencyData.Rate,
	}

	return response, nil
}
