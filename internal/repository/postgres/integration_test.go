package postgres

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/peresvet668/Golang_Pr1/internal/domain"
	"github.com/peresvet668/Golang_Pr1/internal/service"
	httpapi "github.com/peresvet668/Golang_Pr1/internal/transport/http"
)

func TestConcurrentHTTPRequestsKeepWalletBalanceConsistent(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run the PostgreSQL integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	repository, err := Open(ctx, databaseURL, 64)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()
	if err := repository.Migrate(ctx); err != nil {
		t.Fatal(err)
	}

	walletID := uuid.NewString()
	if _, err := repository.db.ExecContext(ctx, `INSERT INTO wallets (id, balance) VALUES ($1::uuid, 0)`, walletID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = repository.db.ExecContext(context.Background(), `DELETE FROM wallets WHERE id = $1::uuid`, walletID)
	})

	server := httptest.NewServer(httpapi.NewHandler(service.New(repository), log.New(io.Discard, "", 0)))
	defer server.Close()

	transport := &http.Transport{MaxIdleConns: 256, MaxIdleConnsPerHost: 256, MaxConnsPerHost: 256}
	client := &http.Client{Transport: transport, Timeout: 30 * time.Second}
	defer transport.CloseIdleConnections()

	const requestCount = 1000
	assertConcurrentStatus(t, client, server.URL, walletID, domain.Deposit, requestCount, http.StatusOK)

	balance, err := repository.GetBalance(ctx, walletID)
	if err != nil {
		t.Fatal(err)
	}
	if balance != requestCount {
		t.Fatalf("balance after concurrent deposits = %d, want %d", balance, requestCount)
	}

	assertConcurrentStatus(t, client, server.URL, walletID, domain.Withdraw, requestCount, http.StatusOK)

	balance, err = repository.GetBalance(ctx, walletID)
	if err != nil {
		t.Fatal(err)
	}
	if balance != 0 {
		t.Fatalf("balance after concurrent withdrawals = %d, want 0", balance)
	}

	response, err := client.Get(server.URL + "/api/v1/wallets/" + walletID)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("GET balance status = %d, want %d", response.StatusCode, http.StatusOK)
	}
}

func assertConcurrentStatus(t *testing.T, client *http.Client, baseURL, walletID string, operation domain.OperationType, count, wantStatus int) {
	t.Helper()

	var wait sync.WaitGroup
	errorsFound := make(chan error, count)
	for i := 0; i < count; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			body := fmt.Sprintf(`{"walletId":%q,"operationType":%q,"amount":1}`, walletID, operation)
			request, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/wallet", strings.NewReader(body))
			if err != nil {
				errorsFound <- err
				return
			}
			request.Header.Set("Content-Type", "application/json")
			response, err := client.Do(request)
			if err != nil {
				errorsFound <- err
				return
			}
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			if response.StatusCode != wantStatus {
				errorsFound <- fmt.Errorf("%s returned HTTP %d, want %d", operation, response.StatusCode, wantStatus)
			}
		}()
	}
	wait.Wait()
	close(errorsFound)
	for err := range errorsFound {
		t.Error(err)
	}
}

func BenchmarkConcurrentWalletWrites(b *testing.B) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		b.Skip("set TEST_DATABASE_URL to run the PostgreSQL benchmark")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	repository, err := Open(ctx, databaseURL, 64)
	if err != nil {
		b.Fatal(err)
	}
	defer repository.Close()
	if err := repository.Migrate(ctx); err != nil {
		b.Fatal(err)
	}

	walletID := uuid.NewString()
	if _, err := repository.db.ExecContext(ctx, `INSERT INTO wallets (id, balance) VALUES ($1::uuid, 0)`, walletID); err != nil {
		b.Fatal(err)
	}
	defer func() {
		_, _ = repository.db.ExecContext(context.Background(), `DELETE FROM wallets WHERE id = $1::uuid`, walletID)
	}()

	server := httptest.NewServer(httpapi.NewHandler(service.New(repository), log.New(io.Discard, "", 0)))
	defer server.Close()
	transport := &http.Transport{MaxIdleConns: 256, MaxIdleConnsPerHost: 256, MaxConnsPerHost: 256}
	client := &http.Client{Transport: transport, Timeout: 30 * time.Second}
	defer transport.CloseIdleConnections()
	requestBody := fmt.Sprintf(`{"walletId":%q,"operationType":"DEPOSIT","amount":1}`, walletID)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			request, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/wallet", strings.NewReader(requestBody))
			if err != nil {
				b.Error(err)
				return
			}
			request.Header.Set("Content-Type", "application/json")
			response, err := client.Do(request)
			if err != nil {
				b.Error(err)
				return
			}
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			if response.StatusCode != http.StatusOK {
				b.Errorf("POST returned HTTP %d, want %d", response.StatusCode, http.StatusOK)
				return
			}
		}
	})
	b.StopTimer()
	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "requests/s")
}
