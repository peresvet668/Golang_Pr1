package http

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/peresvet668/Golang_Pr1/internal/domain"
)

type walletServiceStub struct {
	applyCalled bool
	walletID    string
	operation   domain.OperationType
	amount      int64
}

func (s *walletServiceStub) ApplyOperation(_ context.Context, walletID string, operation domain.OperationType, amount int64) (int64, error) {
	s.applyCalled = true
	s.walletID = walletID
	s.operation = operation
	s.amount = amount
	return amount, nil
}

func (s *walletServiceStub) GetBalance(context.Context, string) (int64, error) {
	return 10, nil
}

func TestApplyAcceptsWalletIDSpellingFromAssignment(t *testing.T) {
	service := &walletServiceStub{}
	handler := NewHandler(service, log.New(io.Discard, "", 0))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", strings.NewReader(`{"valletId":"11111111-1111-4111-8111-111111111111","operationType":"DEPOSIT","amount":25}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if !service.applyCalled || service.walletID != "11111111-1111-4111-8111-111111111111" || service.operation != domain.Deposit || service.amount != 25 {
		t.Fatalf("unexpected service call: %+v", service)
	}
}

func TestApplyRejectsUnknownJSONFields(t *testing.T) {
	handler := NewHandler(&walletServiceStub{}, log.New(io.Discard, "", 0))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/wallet", strings.NewReader(`{"walletId":"11111111-1111-4111-8111-111111111111","operationType":"DEPOSIT","amount":25,"extra":true}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}
