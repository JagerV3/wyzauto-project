package domain

import "time"

type EntityType string

const (
	EntityTypeProduct              EntityType = "product"
	EntityTypeAttribute            EntityType = "attribute"
	EntityTypeProductSpecification EntityType = "product_specification"
)

const (
	FieldProductName = "productname"
	FieldDescription = "description"
	FieldRemark      = "remark"
	FieldLabel       = "label"
	FieldValueLabel  = "value_label"
)

type TranslationLoadRequest struct {
	EntityType EntityType
	EntityID   string
}

type Translation struct {
	EntityType EntityType
	EntityID   string
	Locale     string
	FieldName  string
	FieldValue string
	UpdatedAt  time.Time
}

type TranslationKey struct {
	EntityType EntityType
	EntityID   string
	Locale     string
	FieldName  string
}

type TranslationMap map[TranslationKey]Translation

func (m TranslationMap) Value(entityType EntityType, entityID, locale, fieldName string) (string, bool) {
	translation, ok := m[TranslationKey{
		EntityType: entityType,
		EntityID:   entityID,
		Locale:     locale,
		FieldName:  fieldName,
	}]
	if !ok {
		return "", false
	}

	return translation.FieldValue, true
}

func (m TranslationMap) ValueWithFallback(entityType EntityType, entityID, locale, fieldName string) string {
	if value, ok := m.Value(entityType, entityID, locale, fieldName); ok {
		return value
	}

	if locale != "en" {
		if value, ok := m.Value(entityType, entityID, "en", fieldName); ok {
			return value
		}
	}

	return ""
}
