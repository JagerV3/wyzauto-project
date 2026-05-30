package service

import (
	"context"
	"testing"
	"time"

	"github.com/raymond/wyzauto-project/internal/domain"
)

type fakeProductRepository struct{}

func (fakeProductRepository) FindProduct(_ context.Context, productID string) (domain.Product, error) {
	return domain.Product{ID: productID, SKU: "BP-OIL-5W30-1L", PartNumber: "5W30-1L", Brand: "bosch", CategoryID: "category-1"}, nil
}

func (fakeProductRepository) FindSpecificationsByProduct(_ context.Context, productID string) ([]domain.ProductSpecification, error) {
	return []domain.ProductSpecification{{ID: "spec-1", ProductID: productID, AttributeID: "attr-1", Value: "5w30"}}, nil
}

func (fakeProductRepository) FindAttributesByIDs(_ context.Context, _ []string) ([]domain.Attribute, error) {
	return []domain.Attribute{{ID: "attr-1", Code: "oil_grade"}}, nil
}

type fakeTranslationLoader struct{}

func (fakeTranslationLoader) Load(_ context.Context, entityType domain.EntityType, entityIDs []string, locales []string) (domain.TranslationMap, error) {
	result := domain.TranslationMap{}
	add := func(entityType domain.EntityType, entityID, locale, fieldName, value string) {
		result[domain.TranslationKey{EntityType: entityType, EntityID: entityID, Locale: locale, FieldName: fieldName}] = domain.Translation{
			EntityType: entityType,
			EntityID:   entityID,
			Locale:     locale,
			FieldName:  fieldName,
			FieldValue: value,
		}
	}

	for _, entityID := range entityIDs {
		for _, locale := range locales {
			switch {
			case entityType == domain.EntityTypeProduct && entityID == "product-1" && locale == "en":
				add(entityType, entityID, locale, domain.FieldProductName, "5W-30 Engine Oil 1L")
			case entityType == domain.EntityTypeProduct && entityID == "bosch" && locale == "en":
				add(entityType, entityID, locale, domain.FieldLabel, "Bosch")
			case entityType == domain.EntityTypeProduct && entityID == "bosch" && locale == "th":
				add(entityType, entityID, locale, domain.FieldLabel, "บอช")
			case entityType == domain.EntityTypeProductSpecification && entityID == "spec-1" && locale == "en":
				add(entityType, entityID, locale, domain.FieldValueLabel, "5W-30")
			case entityType == domain.EntityTypeProductSpecification && entityID == "spec-1" && locale == "th":
				add(entityType, entityID, locale, domain.FieldValueLabel, "5W-30")
			}
		}
	}

	return result, nil
}

func (fakeTranslationLoader) Invalidate(_ domain.EntityType, _ string) {}
func (fakeTranslationLoader) LoadUpdatedSince(_ context.Context, _ time.Time, _ []string) ([]domain.Translation, error) {
	return nil, nil
}

func TestProductDocumentBuilderUsesFallbackForMissingTranslations(t *testing.T) {
	builder := NewProductDocumentBuilder(fakeProductRepository{}, fakeTranslationLoader{}, []string{"en", "th"})

	doc, err := builder.Build(context.Background(), "product-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if doc.ProductName[1].Data != "5W-30 Engine Oil 1L" {
		t.Fatalf("expected Thai product name to fallback to English, got %q", doc.ProductName[1].Data)
	}

	if doc.Brand.Label["th"] != "บอช" {
		t.Fatalf("expected Thai brand label, got %q", doc.Brand.Label["th"])
	}
}
