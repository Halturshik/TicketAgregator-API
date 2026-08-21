package service

import (
	"context"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
	"github.com/redis/go-redis/v9"
)

func (s *Service) GetPage(ctx context.Context, searchID string, offset int, limit int, userID *int) (*search.SearchResult, error) {
	if limit <= 0 || limit > MaxPageSize {
		limit = DefaultPageSize
	}
	result, err := s.store.Get(ctx, searchID)
	if err == redis.Nil {
		return nil, apierror.ErrNotFound
	}
	if err != nil {
		logger.Error("Ошибка получения кеша поиска searchID=%s: %v", searchID, err)
		return nil, err
	}
	return s.page(result, offset, limit, userID), nil
}

func (s *Service) page(result *search.CachedResult, offset int, limit int, userID *int) *search.SearchResult {
	if offset < 0 {
		offset = 0
	}
	end := offset + limit
	if offset > len(result.Items) {
		offset = len(result.Items)
	}
	if end > len(result.Items) {
		end = len(result.Items)
	}
	items := append([]search.Offer(nil), result.Items[offset:end]...)
	for i := range items {
		items[i].BonusHint = ""
		if userID == nil {
			items[i].BonusHint = fmt.Sprintf("Зарегистрируйтесь, чтобы получить %d баллов за оплату", items[i].BonusEarn)
		}
	}
	return &search.SearchResult{SearchID: result.SearchID, Total: result.Total, Offset: offset, Limit: limit, Items: items}
}
