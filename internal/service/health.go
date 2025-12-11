package service

import (
	"context"
	"time"

	"github.com/Rusich90/shurl.git/internal/model"
	"github.com/Rusich90/shurl.git/internal/repository"
)

type HealthService struct {
	storage repository.URLRepository
}

func NewHealthService(storage repository.URLRepository) *HealthService {
	return &HealthService{
		storage: storage,
	}
}

func (s *HealthService) Ping() model.PingResponse {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.storage.Ping(ctx); err != nil {
		return model.PingResponse{
			Status:  "error",
			Message: "Storage connection failed",
			Error:   err.Error(),
		}
	}

	switch s.storage.(type) {
	case *repository.DBURLRepository:
		return model.PingResponse{
			Status:  "ok",
			Message: "Database connection successful",
		}
	case *repository.FileURLRepository:
		return model.PingResponse{
			Status:  "ok",
			Message: "Service is running, file storage configured",
		}
	default:
		return model.PingResponse{
			Status:  "ok",
			Message: "Service is running",
		}
	}
}
