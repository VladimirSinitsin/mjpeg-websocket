package service

import (
	"context"
	"errors"
	v1 "stream-server/api/v1"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/emptypb"
)

type HealthService struct {
	db *pgxpool.Pool
}

func NewHealthService(db *pgxpool.Pool) *HealthService {
	return &HealthService{db: db}
}

func (s *HealthService) Live(ctx context.Context, _ *emptypb.Empty) (*v1.HealthReply, error) {
	return &v1.HealthReply{Status: "ok"}, nil
}

func (s *HealthService) Ready(ctx context.Context, _ *emptypb.Empty) (*v1.HealthReply, error) {
	if s.db == nil {
		return nil, errors.New("database is not ready")
	}
	if err := s.db.Ping(ctx); err != nil {
		return nil, err
	}
	return &v1.HealthReply{Status: "ok"}, nil
}
