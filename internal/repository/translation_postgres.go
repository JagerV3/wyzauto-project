package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raymond/wyzauto-project/internal/domain"
)

type TranslationPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewTranslationPostgresRepository(pool *pgxpool.Pool) *TranslationPostgresRepository {
	return &TranslationPostgresRepository{pool: pool}
}

func (r *TranslationPostgresRepository) LoadTranslations(ctx context.Context, entityType domain.EntityType, entityIDs []string, locales []string) (domain.TranslationMap, error) {
	const query = `
		SELECT entity_type, entity_id, locale, field_name, field_value, updated_at
		FROM translation
		WHERE entity_type = $1
		AND entity_id = ANY($2)
		AND locale = ANY($3)
		ORDER BY entity_id, locale, field_name`

	rows, err := r.pool.Query(ctx, query, string(entityType), entityIDs, locales)
	if err != nil {
		return nil, fmt.Errorf("query translation rows: %w", err)
	}
	defer rows.Close()

	translations, err := scanTranslations(rows)
	if err != nil {
		return nil, err
	}

	return translations, nil
}

func (r *TranslationPostgresRepository) LoadUpdatedSince(ctx context.Context, cursor time.Time, locales []string) ([]domain.Translation, error) {
	const query = `
		SELECT entity_type, entity_id, locale, field_name, field_value, updated_at
		FROM translation
		WHERE updated_at > $1
		AND locale = ANY($2)
		ORDER BY updated_at ASC, entity_type, entity_id`

	rows, err := r.pool.Query(ctx, query, cursor, locales)
	if err != nil {
		return nil, fmt.Errorf("query updated translation rows: %w", err)
	}
	defer rows.Close()

	result := make([]domain.Translation, 0)
	for rows.Next() {
		translation, err := scanTranslation(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, translation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate updated translation rows: %w", err)
	}

	return result, nil
}

func scanTranslations(rows pgx.Rows) (domain.TranslationMap, error) {
	translations := domain.TranslationMap{}
	for rows.Next() {
		translation, err := scanTranslation(rows)
		if err != nil {
			return nil, err
		}

		translations[domain.TranslationKey{
			EntityType: translation.EntityType,
			EntityID:   translation.EntityID,
			Locale:     translation.Locale,
			FieldName:  translation.FieldName,
		}] = translation
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate translation rows: %w", err)
	}

	return translations, nil
}

func scanTranslation(rows pgx.Rows) (domain.Translation, error) {
	var translation domain.Translation
	var entityType string
	if err := rows.Scan(
		&entityType,
		&translation.EntityID,
		&translation.Locale,
		&translation.FieldName,
		&translation.FieldValue,
		&translation.UpdatedAt,
	); err != nil {
		return domain.Translation{}, fmt.Errorf("scan translation row: %w", err)
	}

	translation.EntityType = domain.EntityType(entityType)
	return translation, nil
}
