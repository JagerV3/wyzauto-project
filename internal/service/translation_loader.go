package service

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/raymond/wyzauto-project/internal/domain"
)

type TranslationRepository interface {
	LoadTranslations(ctx context.Context, entityType domain.EntityType, entityIDs []string, locales []string) (domain.TranslationMap, error)
	LoadUpdatedSince(ctx context.Context, cursor time.Time, locales []string) ([]domain.Translation, error)
}

type TranslationLoader interface {
	Load(ctx context.Context, entityType domain.EntityType, entityIDs []string, locales []string) (domain.TranslationMap, error)
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

func (l *Loader) Load(ctx context.Context, entityType domain.EntityType, entityIDs []string, locales []string) (domain.TranslationMap, error) {
	entityIDs = uniqueStrings(entityIDs)
	locales = uniqueStrings(locales)
	if len(entityIDs) == 0 || len(locales) == 0 {
		return domain.TranslationMap{}, nil
	}

	result := domain.TranslationMap{}
	missingIDs := make([]string, 0, len(entityIDs))

	for _, entityID := range entityIDs {
		cached, ok := l.cache.Get(entityType, entityID, locales)
		if ok {
			mergeTranslations(result, cached)
			continue
		}
		missingIDs = append(missingIDs, entityID)
	}

	if len(missingIDs) == 0 {
		return result, nil
	}

	loaded, err := l.repo.LoadTranslations(ctx, entityType, missingIDs, locales)
	if err != nil {
		return nil, fmt.Errorf("load translations for %s: %w", entityType, err)
	}

	byEntity := groupByEntity(entityType, missingIDs, loaded)
	for entityID, translations := range byEntity {
		l.cache.Set(entityType, entityID, locales, translations)
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

func groupByEntity(entityType domain.EntityType, entityIDs []string, translations domain.TranslationMap) map[string]domain.TranslationMap {
	grouped := make(map[string]domain.TranslationMap, len(entityIDs))
	for _, entityID := range entityIDs {
		grouped[entityID] = domain.TranslationMap{}
	}

	for key, translation := range translations {
		if key.EntityType != entityType {
			continue
		}
		if _, ok := grouped[key.EntityID]; !ok {
			grouped[key.EntityID] = domain.TranslationMap{}
		}
		grouped[key.EntityID][key] = translation
	}

	return grouped
}
