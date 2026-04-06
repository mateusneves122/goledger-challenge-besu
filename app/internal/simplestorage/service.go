package simplestorage

import (
	"context"
	"fmt"
	"math/big"
)

type ContractReader interface {
	GetValue(ctx context.Context) (*big.Int, error)
}

type ContractWriter interface {
	SetValue(ctx context.Context, value *big.Int) (txHash string, err error)
}

type Repository interface {
	Save(ctx context.Context, contractAddress, value string) error
	GetLatest(ctx context.Context, contractAddress string) (string, error)
}

type Service struct {
	reader          ContractReader
	writer          ContractWriter
	repo            Repository
	contractAddress string
}

func NewService(reader ContractReader, writer ContractWriter, repo Repository, contractAddress string) *Service {
	return &Service{reader: reader, writer: writer, repo: repo, contractAddress: contractAddress}
}

func (s *Service) Set(ctx context.Context, value *big.Int) (txHash string, err error) {
	txHash, err = s.writer.SetValue(ctx, value)
	if err != nil {
		return "", fmt.Errorf("blockchain set: %w", err)
	}

	return txHash, nil
}

func (s *Service) Get(ctx context.Context) (string, error) {
	v, err := s.reader.GetValue(ctx)
	if err != nil {
		return "", fmt.Errorf("blockchain get: %w", err)
	}
	return v.String(), nil
}

func (s *Service) Sync(ctx context.Context) (string, error) {
	v, err := s.reader.GetValue(ctx)
	if err != nil {
		return "", fmt.Errorf("blockchain get: %w", err)
	}
	value := v.String()

	if err = s.repo.Save(ctx, s.contractAddress, value); err != nil {
		return "", fmt.Errorf("db save: %w", err)
	}

	return value, nil
}

func (s *Service) Check(ctx context.Context) (bool, error) {
	v, err := s.reader.GetValue(ctx)
	if err != nil {
		return false, fmt.Errorf("blockchain get: %w", err)
	}

	dbValue, err := s.repo.GetLatest(ctx, s.contractAddress)
	if err != nil {
		return false, fmt.Errorf("db get latest: %w", err)
	}

	return v.String() == dbValue, nil
}
