package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Halturshik/TicketAgregator-API/internal/common/apierror"
	"github.com/Halturshik/TicketAgregator-API/internal/platform/logger"
	"github.com/Halturshik/TicketAgregator-API/internal/search"
)

func (s *Service) GetPage(ctx context.Context, searchID string, offset int, limit int, userID *int) (*search.SearchResult, error) {
	if limit <= 0 || limit > MaxPageSize {
		limit = DefaultPageSize
	}
	result, err := s.store.Get(ctx, searchID)
	if errors.Is(err, search.ErrCachedResultNotFound) {
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
	if offset > len(result.Items) {
		offset = len(result.Items)
	}
	end := offset + min(limit, len(result.Items)-offset)
	items := append([]search.TripOption(nil), result.Items[offset:end]...)
	for i := range items {
		items[i].BonusHint = ""
		if userID == nil {
			items[i].BonusHint = fmt.Sprintf("Зарегистрируйтесь, чтобы получить %d баллов за оплату", items[i].BonusEarn)
		}
	}
	return &search.SearchResult{SearchID: result.SearchID, Total: result.Total, Offset: offset, Limit: limit, Items: items}
}
