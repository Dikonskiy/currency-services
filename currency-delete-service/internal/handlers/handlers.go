package handlers

import (
	"context"
	pb "currency-delete-service/internal/generated"
	"currency-delete-service/internal/loger"
	"currency-delete-service/internal/repository"
	"encoding/json"
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

// @Summary Delete currency data by date and code
// @Description Delete currency data for a specific date and currency code.
// @Tags currency
// @Accept json
// @Produce json
// @Param date path string true "Date in DD.MM.YYYY format"
// @Param code path string true "Currency code (e.g., USD)"
// @Success 200 {object} map[string]interface{} "Deleted data"
// @Failure 400 {object} map[string]string "Failed to parse the date"
// @Failure 500 {object} map[string]string "Failed to retrieve data"
// @Router /currency/{date}/{code} [delete]
func (h *Handler) DeleteCurrencyHandler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	vars := mux.Vars(r)
	date := vars["date"]
	code := vars["code"]

	formattedDate, err := DateFormat(date)
	if err != nil {
		h.RespondWithError(w, http.StatusBadRequest, "Failed to parse the date", err)
		return
	}

	data, err := h.R.DeleteData(ctx, formattedDate, code)
	if err != nil {
		h.RespondWithError(w, http.StatusInternalServerError, "Failed to retrieve data", err)
		return
	}
	h.R.Logerr.Info("Data was showed")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *CurrencyServiceServer) DeleteCurrency(ctx context.Context, req *pb.DeleteCurrencyRequest) (*pb.DeleteCurrencyResponse, error) {
	// Call the repository method to delete currency data
	success, err := s.Repo.DeleteCurrency(ctx, req)
	if err != nil {
		return nil, err
	}

	// Construct the response message
	response := &pb.DeleteCurrencyResponse{
		Success: success,
	}

	return response, nil
}
