package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/peresvet668/Golang_Pr1/internal/domain"
)

const databaseOperationTimeout = 10 * time.Second

type walletOperations interface {
	ApplyOperation(ctx context.Context, walletID string, operation domain.OperationType, amount int64) (int64, error)
	GetBalance(ctx context.Context, walletID string) (int64, error)
}

type handler struct {
	wallets walletOperations
	logger  *log.Logger
}

type operationRequest struct {
	WalletID      string               `json:"walletId"`
	PDFWalletID   string               `json:"valletId,omitempty"`
	OperationType domain.OperationType `json:"operationType"`
	Amount        int64                `json:"amount"`
}

type walletResponse struct {
	WalletID string `json:"walletId"`
	Balance  int64  `json:"balance"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewHandler(wallets walletOperations, logger *log.Logger) http.Handler {
	if logger == nil {
		logger = log.Default()
	}

	h := &handler{wallets: wallets, logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/wallet", h.applyOperation)
	mux.HandleFunc("GET /api/v1/wallets/{walletID}", h.getBalance)
	mux.HandleFunc("GET /healthz", h.health)

	return recoverPanics(logger, mux)
}

func (h *handler) applyOperation(w http.ResponseWriter, r *http.Request) {
	var request operationRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	walletID, err := request.walletID()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), databaseOperationTimeout)
	defer cancel()

	balance, err := h.wallets.ApplyOperation(ctx, walletID, request.OperationType, request.Amount)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, walletResponse{WalletID: walletID, Balance: balance})
}

func (h *handler) getBalance(w http.ResponseWriter, r *http.Request) {
	walletID := r.PathValue("walletID")
	ctx, cancel := context.WithTimeout(r.Context(), databaseOperationTimeout)
	defer cancel()

	balance, err := h.wallets.GetBalance(ctx, walletID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, walletResponse{WalletID: walletID, Balance: balance})
}

func (h *handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidWalletID), errors.Is(err, domain.ErrInvalidAmount), errors.Is(err, domain.ErrInvalidOperation):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrWalletNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
	case errors.Is(err, domain.ErrInsufficientFunds), errors.Is(err, domain.ErrBalanceOverflow):
		writeJSON(w, http.StatusConflict, errorResponse{Error: err.Error()})
	default:
		h.logger.Printf("wallet request failed: %v", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}

func (request operationRequest) walletID() (string, error) {
	if request.WalletID != "" && request.PDFWalletID != "" && request.WalletID != request.PDFWalletID {
		return "", errors.New("walletId and valletId must match when both are provided")
	}
	if request.WalletID != "" {
		return request.WalletID, nil
	}
	if request.PDFWalletID != "" {
		return request.PDFWalletID, nil
	}
	return "", errors.New("walletId is required")
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("request contains multiple JSON values")
		}
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func recoverPanics(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Printf("panic while handling %s %s: %v", r.Method, r.URL.Path, recovered)
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
