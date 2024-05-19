package handlers

import (
	"context"
	"currency-save-service/internal/config"
	pb "currency-save-service/internal/generated"
	"currency-save-service/internal/loger"
	"currency-save-service/internal/repository"
	"currency-save-service/internal/service"
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
	if repo == nil {
		repo.Logerr.Error("Failed to initialize the repository")
	}

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

// @Summary Save currency data
// @Description Save currency data for a specific date.
// @Tags currency
// @Accept json
// @Param date path string true "Date in DD.MM.YYYY format"
// @Router /currency/save/{date} [post]
func (h *Handler) SaveCurrencyHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	vars := mux.Vars(r)
	date := vars["date"]

	formattedDate, err := DateFormat(date)
	if err != nil {
		h.RespondWithError(w, http.StatusBadRequest, "Failed to parse and format the date", err)
		return
	}

	var service = service.NewService(h.R.Logerr, h.R.Metrics)
	cfg := config.Config{}
	apiUrl := cfg.GetEnv("API_URL", "")

	go h.R.InsertData(*service.GetData(ctx, date, apiUrl), formattedDate)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true}`))
	h.R.Logerr.Info("Success: true")
}

func (s *CurrencyServiceServer) SaveCurrency(ctx context.Context, req *pb.SaveCurrencyRequest) (*pb.SaveCurrencyResponse, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}

	success, err := s.Repo.SaveCurrency(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.SaveCurrencyResponse{Success: success}, nil
}
