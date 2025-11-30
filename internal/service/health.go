package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/Rusich90/shurl.git/internal/model"
)

type HealthService struct {
	db *sql.DB
}

func NewHealthService(db *sql.DB) *HealthService {
	return &HealthService{
		db: db,
	}
}

func (s *HealthService) Ping() model.PingResponse {
	if s.db == nil {
		return model.PingResponse{
			Status:  "ok",
			Message: "Service is running, database not configured",
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		return model.PingResponse{
			Status:  "error",
			Message: "Database connection failed",
			Error:   err.Error(),
		}
	}

	return model.PingResponse{
		Status:  "ok",
		Message: "Database connection successful",
	}
}
