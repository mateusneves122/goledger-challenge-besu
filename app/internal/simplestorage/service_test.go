package simplestorage

import (
	"context"
	"errors"
	"math/big"
	"testing"
)

// --- mocks ---

type mockReader struct {
	getValueFn func(ctx context.Context) (*big.Int, error)
}

func (m *mockReader) GetValue(ctx context.Context) (*big.Int, error) {
	return m.getValueFn(ctx)
}

type mockWriter struct {
	setValueFn func(ctx context.Context, value *big.Int) (string, error)
}

func (m *mockWriter) SetValue(ctx context.Context, value *big.Int) (string, error) {
	return m.setValueFn(ctx, value)
}

type mockRepo struct {
	saveFn      func(ctx context.Context, contractAddress, value string) error
	getLatestFn func(ctx context.Context, contractAddress string) (string, error)
}

func (m *mockRepo) Save(ctx context.Context, contractAddress, value string) error {
	return m.saveFn(ctx, contractAddress, value)
}

func (m *mockRepo) GetLatest(ctx context.Context, contractAddress string) (string, error) {
	return m.getLatestFn(ctx, contractAddress)
}

func newService(r ContractReader, w ContractWriter, repo Repository) *Service {
	return NewService(r, w, repo, "0xcontract")
}

// --- Set ---

func TestSet_Success(t *testing.T) {
	writer := &mockWriter{
		setValueFn: func(_ context.Context, v *big.Int) (string, error) {
			if v.String() != "42" {
				t.Errorf("expected value 42, got %s", v.String())
			}
			return "0xtxhash", nil
		},
	}
	svc := newService(nil, writer, nil)

	txHash, err := svc.Set(context.Background(), big.NewInt(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if txHash != "0xtxhash" {
		t.Errorf("expected 0xtxhash, got %s", txHash)
	}
}

func TestSet_BlockchainError(t *testing.T) {
	writer := &mockWriter{
		setValueFn: func(_ context.Context, _ *big.Int) (string, error) {
			return "", errors.New("connection refused")
		},
	}
	svc := newService(nil, writer, nil)

	_, err := svc.Set(context.Background(), big.NewInt(1))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- Get ---

func TestGet_Success(t *testing.T) {
	reader := &mockReader{
		getValueFn: func(_ context.Context) (*big.Int, error) {
			return big.NewInt(99), nil
		},
	}
	svc := newService(reader, nil, nil)

	val, err := svc.Get(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "99" {
		t.Errorf("expected 99, got %s", val)
	}
}

func TestGet_BlockchainError(t *testing.T) {
	reader := &mockReader{
		getValueFn: func(_ context.Context) (*big.Int, error) {
			return nil, errors.New("node unreachable")
		},
	}
	svc := newService(reader, nil, nil)

	_, err := svc.Get(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- Sync ---

func TestSync_Success(t *testing.T) {
	reader := &mockReader{
		getValueFn: func(_ context.Context) (*big.Int, error) {
			return big.NewInt(7), nil
		},
	}
	repo := &mockRepo{
		saveFn: func(_ context.Context, addr, value string) error {
			if value != "7" {
				t.Errorf("expected value 7, got %s", value)
			}
			return nil
		},
	}
	svc := newService(reader, nil, repo)

	val, err := svc.Sync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "7" {
		t.Errorf("expected 7, got %s", val)
	}
}

func TestSync_BlockchainError(t *testing.T) {
	reader := &mockReader{
		getValueFn: func(_ context.Context) (*big.Int, error) {
			return nil, errors.New("rpc error")
		},
	}
	svc := newService(reader, nil, nil)

	_, err := svc.Sync(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSync_SaveError(t *testing.T) {
	reader := &mockReader{
		getValueFn: func(_ context.Context) (*big.Int, error) {
			return big.NewInt(5), nil
		},
	}
	repo := &mockRepo{
		saveFn: func(_ context.Context, _, _ string) error {
			return errors.New("db error")
		},
	}
	svc := newService(reader, nil, repo)

	_, err := svc.Sync(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- Check ---

func TestCheck_Match(t *testing.T) {
	reader := &mockReader{
		getValueFn: func(_ context.Context) (*big.Int, error) {
			return big.NewInt(42), nil
		},
	}
	repo := &mockRepo{
		getLatestFn: func(_ context.Context, _ string) (string, error) {
			return "42", nil
		},
	}
	svc := newService(reader, nil, repo)

	match, err := svc.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !match {
		t.Error("expected match=true")
	}
}

func TestCheck_NoMatch(t *testing.T) {
	reader := &mockReader{
		getValueFn: func(_ context.Context) (*big.Int, error) {
			return big.NewInt(42), nil
		},
	}
	repo := &mockRepo{
		getLatestFn: func(_ context.Context, _ string) (string, error) {
			return "0", nil
		},
	}
	svc := newService(reader, nil, repo)

	match, err := svc.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if match {
		t.Error("expected match=false")
	}
}

func TestCheck_BlockchainError(t *testing.T) {
	reader := &mockReader{
		getValueFn: func(_ context.Context) (*big.Int, error) {
			return nil, errors.New("rpc error")
		},
	}
	svc := newService(reader, nil, nil)

	_, err := svc.Check(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCheck_DBError(t *testing.T) {
	reader := &mockReader{
		getValueFn: func(_ context.Context) (*big.Int, error) {
			return big.NewInt(1), nil
		},
	}
	repo := &mockRepo{
		getLatestFn: func(_ context.Context, _ string) (string, error) {
			return "", errors.New("db error")
		},
	}
	svc := newService(reader, nil, repo)

	_, err := svc.Check(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
