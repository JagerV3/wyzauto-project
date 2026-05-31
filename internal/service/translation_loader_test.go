package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/raymond/wyzauto-project/internal/domain"
)

type fakeTranslationRepository struct {
	calls        int
	lastRequests []domain.TranslationLoadRequest
	translations domain.TranslationMap
	err          error
}

func (f *fakeTranslationRepository) LoadTranslations(
	_ context.Context,
	requests []domain.TranslationLoadRequest,
	locales []string,
) (domain.TranslationMap, error) {
	f.calls++
	f.lastRequests = append([]domain.TranslationLoadRequest(nil), requests...)

	if f.err != nil {
		return nil, f.err
	}

	result := domain.TranslationMap{}

	for key, value := range f.translations {
		if containsRequest(requests, key.EntityType, key.EntityID) && contains(locales, key.Locale) {
			result[key] = value
		}
	}

	return result, nil
}

func (f *fakeTranslationRepository) LoadUpdatedSince(_ context.Context, _ time.Time, _ []string) ([]domain.Translation, error) {
	return nil, nil
}

func TestTranslationLoaderLoadsTranslationsInBulk(t *testing.T) {
	repo := &fakeTranslationRepository{translations: sampleTranslations()}
	loader := NewTranslationLoader(repo, time.Minute)

	result, err := loader.Load(context.Background(), []domain.TranslationLoadRequest{
		{EntityType: domain.EntityTypeProduct, EntityID: "p1"},
		{EntityType: domain.EntityTypeProduct, EntityID: "p2"},
	}, []string{"en", "th"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repo.calls != 1 {
		t.Fatalf("expected one repository call, got %d", repo.calls)
	}

	if len(repo.lastRequests) != 2 {
		t.Fatalf("expected two requests to be loaded in bulk, got %v", repo.lastRequests)
	}

	got := result.ValueWithFallback(domain.EntityTypeProduct, "p1", "th", domain.FieldProductName)
	if got != "น้ำมันเครื่อง" {
		t.Fatalf("expected Thai product name, got %q", got)
	}
}

func TestTranslationLoaderUsesCache(t *testing.T) {
	repo := &fakeTranslationRepository{translations: sampleTranslations()}
	loader := NewTranslationLoader(repo, time.Minute)

	_, err := loader.Load(context.Background(), []domain.TranslationLoadRequest{
		{EntityType: domain.EntityTypeProduct, EntityID: "p1"},
	}, []string{"en"})
	if err != nil {
		t.Fatalf("expected first load to succeed: %v", err)
	}

	_, err = loader.Load(context.Background(), []domain.TranslationLoadRequest{
		{EntityType: domain.EntityTypeProduct, EntityID: "p1"},
	}, []string{"en"})
	if err != nil {
		t.Fatalf("expected second load to succeed: %v", err)
	}

	if repo.calls != 1 {
		t.Fatalf("expected cached second load, got %d repository calls", repo.calls)
	}
}

func TestTranslationLoaderInvalidatesEntity(t *testing.T) {
	repo := &fakeTranslationRepository{translations: sampleTranslations()}
	loader := NewTranslationLoader(repo, time.Minute)

	_, _ = loader.Load(context.Background(), []domain.TranslationLoadRequest{
		{EntityType: domain.EntityTypeProduct, EntityID: "p1"},
	}, []string{"en"})

	loader.Invalidate(domain.EntityTypeProduct, "p1")

	_, _ = loader.Load(context.Background(), []domain.TranslationLoadRequest{
		{EntityType: domain.EntityTypeProduct, EntityID: "p1"},
	}, []string{"en"})

	if repo.calls != 2 {
		t.Fatalf("expected cache invalidation to force reload, got %d repository calls", repo.calls)
	}
}

func TestTranslationLoaderWrapsRepositoryErrors(t *testing.T) {
	repo := &fakeTranslationRepository{err: errors.New("database unavailable")}
	loader := NewTranslationLoader(repo, time.Minute)

	_, err := loader.Load(context.Background(), []domain.TranslationLoadRequest{
		{EntityType: domain.EntityTypeProduct, EntityID: "p1"},
	}, []string{"en"})
	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, repo.err) {
		t.Fatalf("expected wrapped repository error, got %v", err)
	}
}

func sampleTranslations() domain.TranslationMap {
	translations := []domain.Translation{
		{EntityType: domain.EntityTypeProduct, EntityID: "p1", Locale: "en", FieldName: domain.FieldProductName, FieldValue: "Engine Oil"},
		{EntityType: domain.EntityTypeProduct, EntityID: "p1", Locale: "th", FieldName: domain.FieldProductName, FieldValue: "น้ำมันเครื่อง"},
		{EntityType: domain.EntityTypeProduct, EntityID: "p2", Locale: "en", FieldName: domain.FieldProductName, FieldValue: "Brake Pad"},
	}

	result := domain.TranslationMap{}
	for _, translation := range translations {
		result[domain.TranslationKey{
			EntityType: translation.EntityType,
			EntityID:   translation.EntityID,
			Locale:     translation.Locale,
			FieldName:  translation.FieldName,
		}] = translation
	}

	return result
}

func containsRequest(requests []domain.TranslationLoadRequest, entityType domain.EntityType, entityID string) bool {
	for _, request := range requests {
		if request.EntityType == entityType && request.EntityID == entityID {
			return true
		}
	}

	return false
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}

	return false
}
