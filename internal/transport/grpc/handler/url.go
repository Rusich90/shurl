// Package handler предоставляет gRPC-обработчики (хендлеры) для API-эндпоинтов.
//
// Содержит обработчики для создания, получения и списка коротких URL.
package handler

import (
	"context"
	"net/url"
	"strings"

	"github.com/Rusich90/shurl.git/internal/config"
	"github.com/Rusich90/shurl.git/internal/service"
	"github.com/Rusich90/shurl.git/internal/transport/grpc/authcontext"
	proto "github.com/Rusich90/shurl.git/pkg"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Handler обрабатывает gRPC-запросы для работы с короткими URL.
type Handler struct {
	proto.UnimplementedShortenerServiceServer
	urlService *service.URLService
	cfg        *config.Config
	logger     *zap.Logger
}

// NewHandler создает новый Handler с указанными зависимостями.
func NewHandler(urlService *service.URLService, cfg *config.Config, logger *zap.Logger) *Handler {
	return &Handler{
		urlService: urlService,
		cfg:        cfg,
		logger:     logger,
	}
}

// ShortenURL создает короткую ссылку для указанного URL.
func (h *Handler) ShortenURL(ctx context.Context, req *proto.URLShortenRequest) (*proto.URLShortenResponse, error) {
	if req.GetUrl() == "" {
		return nil, status.Error(codes.InvalidArgument, "URL is required")
	}

	if !strings.HasPrefix(req.GetUrl(), "http://") && !strings.HasPrefix(req.GetUrl(), "https://") {
		return nil, status.Error(codes.InvalidArgument, "Invalid URL format")
	}

	userID, err := authcontext.GetUserID(ctx)
	if err != nil {
		h.logger.Error("Failed to get user ID from context", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get user ID")
	}

	result, err := h.urlService.CreateShortURL(ctx, req.GetUrl(), userID)
	if err != nil {
		h.logger.Error("Failed to create short URL", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to create short URL")
	}

	return proto.URLShortenResponse_builder{
		Result: result.URL,
	}.Build(), nil
}

// ExpandURL возвращает исходный URL по короткому идентификатору.
func (h *Handler) ExpandURL(ctx context.Context, req *proto.URLExpandRequest) (*proto.URLExpandResponse, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	domainURL, ok := h.urlService.GetOriginalURL(ctx, req.GetId())
	if !ok {
		return nil, status.Error(codes.NotFound, "URL not found")
	}

	if domainURL.IsDeleted {
		return nil, status.Error(codes.NotFound, "URL is deleted")
	}

	return proto.URLExpandResponse_builder{
		Result: domainURL.OriginalURL,
	}.Build(), nil
}

// ListUserURLs возвращает все URL пользователя.
func (h *Handler) ListUserURLs(ctx context.Context, req *emptypb.Empty) (*proto.UserURLsResponse, error) {
	userID, err := authcontext.GetUserID(ctx)
	if err != nil {
		h.logger.Error("Failed to get user ID from context", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get user ID")
	}

	if userID == nil {
		return nil, status.Error(codes.Unauthenticated, "user-id is required")
	}

	urls, err := h.urlService.GetUserOriginalURLs(ctx, userID)
	if err != nil {
		h.logger.Error("Failed to get user URLs", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed to get user URLs")
	}

	var urlData []*proto.URLData
	for _, domainURL := range urls {
		shortURL, err := url.JoinPath(h.cfg.BaseURL, domainURL.ShortURL)
		if err != nil {
			h.logger.Error("Failed to build short URL", zap.Error(err))
			continue
		}

		urlData = append(urlData, proto.URLData_builder{
			ShortUrl:    shortURL,
			OriginalUrl: domainURL.OriginalURL,
		}.Build())
	}

	return proto.UserURLsResponse_builder{
		Url: urlData,
	}.Build(), nil
}
