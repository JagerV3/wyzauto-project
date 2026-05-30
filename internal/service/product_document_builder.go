package service

import (
	"context"
	"fmt"

	"github.com/raymond/wyzauto-project/internal/domain"
)

type ProductRepository interface {
	FindProduct(ctx context.Context, productID string) (domain.Product, error)
	FindSpecificationsByProduct(ctx context.Context, productID string) ([]domain.ProductSpecification, error)
	FindAttributesByIDs(ctx context.Context, attributeIDs []string) ([]domain.Attribute, error)
}

type ProductDocumentBuilder struct {
	products           ProductRepository
	translations       TranslationLoader
	locales            []string
	defaultFallbackLoc string
}

func NewProductDocumentBuilder(products ProductRepository, translations TranslationLoader, locales []string) *ProductDocumentBuilder {
	return &ProductDocumentBuilder{
		products:           products,
		translations:       translations,
		locales:            locales,
		defaultFallbackLoc: "en",
	}
}

func (b *ProductDocumentBuilder) Build(ctx context.Context, productID string) (domain.ProductDocument, error) {
	product, err := b.products.FindProduct(ctx, productID)
	if err != nil {
		return domain.ProductDocument{}, fmt.Errorf("find product %s: %w", productID, err)
	}

	specs, err := b.products.FindSpecificationsByProduct(ctx, productID)
	if err != nil {
		return domain.ProductDocument{}, fmt.Errorf("find product specifications for %s: %w", productID, err)
	}

	attributeIDs := make([]string, 0, len(specs))
	for _, spec := range specs {
		attributeIDs = append(attributeIDs, spec.AttributeID)
	}

	attributes, err := b.products.FindAttributesByIDs(ctx, attributeIDs)
	if err != nil {
		return domain.ProductDocument{}, fmt.Errorf("find attributes for product %s: %w", productID, err)
	}

	attributeByID := make(map[string]domain.Attribute, len(attributes))
	for _, attribute := range attributes {
		attributeByID[attribute.ID] = attribute
	}

	productTranslations, err := b.translations.Load(ctx, domain.EntityTypeProduct, []string{product.ID, product.Brand}, b.locales)
	if err != nil {
		return domain.ProductDocument{}, err
	}

	attributeTranslations, err := b.translations.Load(ctx, domain.EntityTypeAttribute, attributeIDs, b.locales)
	if err != nil {
		return domain.ProductDocument{}, err
	}

	specIDs := make([]string, 0, len(specs))
	for _, spec := range specs {
		specIDs = append(specIDs, spec.ID)
	}
	specTranslations, err := b.translations.Load(ctx, domain.EntityTypeProductSpecification, specIDs, b.locales)
	if err != nil {
		return domain.ProductDocument{}, err
	}

	doc := domain.ProductDocument{
		UUID:       product.ID,
		SKU:        product.SKU,
		PartNumber: product.PartNumber,
		Brand: domain.BrandDocument{
			Code:  product.Brand,
			Label: labelsFor(productTranslations, domain.EntityTypeProduct, product.Brand, domain.FieldLabel, b.locales),
		},
		ProductName: productNamesFor(productTranslations, product.ID, b.locales),
		Attributes:  map[string]string{},
		Dynamic:     map[string]domain.AttributeValueDocument{},
	}

	for _, spec := range specs {
		attribute, ok := attributeByID[spec.AttributeID]
		if !ok {
			continue
		}

		doc.Attributes[attribute.Code] = spec.Value
		valueDoc := domain.AttributeValueDocument{
			Code:  spec.Value,
			Label: labelsFor(specTranslations, domain.EntityTypeProductSpecification, spec.ID, domain.FieldValueLabel, b.locales),
		}

		if isEmptyLabel(valueDoc.Label) {
			valueDoc.Label = labelsFor(attributeTranslations, domain.EntityTypeAttribute, attribute.ID, domain.FieldLabel, b.locales)
		}

		doc.Dynamic[attribute.Code] = valueDoc
		if attribute.Code == "oil_grade" {
			doc.OilGrade = &valueDoc
		}
	}

	return doc, nil
}

func productNamesFor(translations domain.TranslationMap, productID string, locales []string) []domain.ProductNameDocument {
	names := make([]domain.ProductNameDocument, 0, len(locales))
	for _, locale := range locales {
		name := translations.ValueWithFallback(domain.EntityTypeProduct, productID, locale, domain.FieldProductName)
		names = append(names, domain.ProductNameDocument{Locale: locale, Data: name})
	}
	return names
}

func labelsFor(translations domain.TranslationMap, entityType domain.EntityType, entityID string, fieldName string, locales []string) map[string]string {
	labels := make(map[string]string, len(locales))
	for _, locale := range locales {
		labels[locale] = translations.ValueWithFallback(entityType, entityID, locale, fieldName)
	}
	return labels
}

func isEmptyLabel(labels map[string]string) bool {
	for _, value := range labels {
		if value != "" {
			return false
		}
	}
	return true
}
