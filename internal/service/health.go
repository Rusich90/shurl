package service

import (
	"context"
	"time"

	"github.com/Rusich90/shurl.git/internal/domain/url"
	"github.com/Rusich90/shurl.git/internal/repository/file"
	"github.com/Rusich90/shurl.git/internal/repository/postgres"
	"github.com/Rusich90/shurl.git/internal/transport/http/dto"
)

type HealthService struct {
	storage domain.URLRepository
}

func NewHealthService(storage domain.URLRepository) *HealthService {
	return &HealthService{
		storage: storage,
	}
}

func (s *HealthService) Ping() dto.PingResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.storage.Ping(ctx); err != nil {
		return dto.PingResponse{
			Status:  "error",
			Message: "Storage connection failed",
			Error:   err.Error(),
		}
	}

	switch s.storage.(type) {
	case *postgres.DBURLRepository:
		return dto.PingResponse{
			Status:  "ok",
			Message: "Database connection successful",
		}
	case *file.FileURLRepository:
		return dto.PingResponse{
			Status:  "ok",
			Message: "Service is running, file storage configured",
		}
	default:
		return dto.PingResponse{
			Status:  "ok",
			Message: "Service is running",
		}
	}
}
