package service

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/raymond/wyzauto-project/internal/domain"
)

type TranslationRepository interface {
	LoadTranslations(ctx context.Context, requests []domain.TranslationLoadRequest, locales []string) (domain.TranslationMap, error)
	LoadUpdatedSince(ctx context.Context, cursor time.Time, locales []string) ([]domain.Translation, error)
}

type TranslationLoader interface {
	Load(ctx context.Context, requests []domain.TranslationLoadRequest, locales []string) (domain.TranslationMap, error)
	Invalidate(entityType domain.EntityType, entityID string)
	LoadUpdatedSince(ctx context.Context, cursor time.Time, locales []string) ([]domain.Translation, error)
}

type Loader struct {
	repo  TranslationRepository
	cache *TranslationCache
}

func NewTranslationLoader(repo TranslationRepository, ttl time.Duration) *Loader {
	return &Loader{
		repo:  repo,
		cache: NewTranslationCache(ttl),
	}
}

func (l *Loader) Load(
	ctx context.Context,
	requests []domain.TranslationLoadRequest,
	locales []string,
) (domain.TranslationMap, error) {
	requests = uniqueRequests(requests)
	locales = uniqueStrings(locales)

	if len(requests) == 0 || len(locales) == 0 {
		return domain.TranslationMap{}, nil
	}

	result := domain.TranslationMap{}
	missingRequests := make([]domain.TranslationLoadRequest, 0, len(requests))

	for _, request := range requests {
		cached, ok := l.cache.Get(request.EntityType, request.EntityID, locales)
		if ok {
			mergeTranslations(result, cached)
			continue
		}

		missingRequests = append(missingRequests, request)
	}

	if len(missingRequests) == 0 {
		return result, nil
	}

	loaded, err := l.repo.LoadTranslations(ctx, missingRequests, locales)
	if err != nil {
		return nil, fmt.Errorf("load translations: %w", err)
	}

	byRequest := groupByRequest(missingRequests, loaded)
	for request, translations := range byRequest {
		l.cache.Set(request.EntityType, request.EntityID, locales, translations)
	}

	mergeTranslations(result, loaded)

	return result, nil
}

func (l *Loader) Invalidate(entityType domain.EntityType, entityID string) {
	l.cache.Invalidate(entityType, entityID)
}

func (l *Loader) LoadUpdatedSince(ctx context.Context, cursor time.Time, locales []string) ([]domain.Translation, error) {
	translations, err := l.repo.LoadUpdatedSince(ctx, cursor, uniqueStrings(locales))
	if err != nil {
		return nil, fmt.Errorf("load translations updated since %s: %w", cursor.Format(time.RFC3339Nano), err)
	}

	for _, translation := range translations {
		l.Invalidate(translation.EntityType, translation.EntityID)
	}

	return translations, nil
}

func uniqueRequests(requests []domain.TranslationLoadRequest) []domain.TranslationLoadRequest {
	seen := make(map[domain.TranslationLoadRequest]struct{}, len(requests))
	unique := make([]domain.TranslationLoadRequest, 0, len(requests))

	for _, request := range requests {
		if request.EntityID == "" {
			continue
		}

		if _, ok := seen[request]; ok {
			continue
		}

		seen[request] = struct{}{}
		unique = append(unique, request)
	}

	return unique
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	unique := make([]string, 0, len(values))

	for _, value := range values {
		if value == "" {
			continue
		}

		if _, ok := seen[value]; ok {
			continue
		}

		seen[value] = struct{}{}
		unique = append(unique, value)
	}

	slices.Sort(unique)

	return unique
}

func mergeTranslations(dst, src domain.TranslationMap) {
	for key, value := range src {
		dst[key] = value
	}
}

func groupByRequest(
	requests []domain.TranslationLoadRequest,
	translations domain.TranslationMap,
) map[domain.TranslationLoadRequest]domain.TranslationMap {
	grouped := make(map[domain.TranslationLoadRequest]domain.TranslationMap, len(requests))

	for _, request := range requests {
		grouped[request] = domain.TranslationMap{}
	}

	for key, translation := range translations {
		request := domain.TranslationLoadRequest{
			EntityType: key.EntityType,
			EntityID:   key.EntityID,
		}

		if _, ok := grouped[request]; !ok {
			continue
		}

		grouped[request][key] = translation
	}

	return grouped
}
