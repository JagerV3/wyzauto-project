package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raymond/wyzauto-project/internal/domain"
)

type ProductPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewProductPostgresRepository(pool *pgxpool.Pool) *ProductPostgresRepository {
	return &ProductPostgresRepository{pool: pool}
}

func (r *ProductPostgresRepository) FindProduct(ctx context.Context, productID string) (domain.Product, error) {
	const query = `
SELECT id::text, sku, part_number, brand, category_id::text
FROM product
WHERE id = $1`

	var product domain.Product
	if err := r.pool.QueryRow(ctx, query, productID).Scan(
		&product.ID,
		&product.SKU,
		&product.PartNumber,
		&product.Brand,
		&product.CategoryID,
	); err != nil {
		if err == pgx.ErrNoRows {
			return domain.Product{}, fmt.Errorf("product %s not found: %w", productID, err)
		}
		return domain.Product{}, fmt.Errorf("query product %s: %w", productID, err)
	}

	return product, nil
}

func (r *ProductPostgresRepository) FindSpecificationsByProduct(ctx context.Context, productID string) ([]domain.ProductSpecification, error) {
	const query = `
SELECT id::text, product_id::text, attribute_id::text, value
FROM product_specification
WHERE product_id = $1
ORDER BY id`

	rows, err := r.pool.Query(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("query product specifications for %s: %w", productID, err)
	}
	defer rows.Close()

	specs := make([]domain.ProductSpecification, 0)
	for rows.Next() {
		var spec domain.ProductSpecification
		if err := rows.Scan(&spec.ID, &spec.ProductID, &spec.AttributeID, &spec.Value); err != nil {
			return nil, fmt.Errorf("scan product specification: %w", err)
		}
		specs = append(specs, spec)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate product specifications: %w", err)
	}

	return specs, nil
}

func (r *ProductPostgresRepository) FindAttributesByIDs(ctx context.Context, attributeIDs []string) ([]domain.Attribute, error) {
	if len(attributeIDs) == 0 {
		return []domain.Attribute{}, nil
	}

	const query = `
SELECT id::text, code, metric_unit
FROM attribute
WHERE id = ANY($1)
ORDER BY code`

	rows, err := r.pool.Query(ctx, query, attributeIDs)
	if err != nil {
		return nil, fmt.Errorf("query attributes: %w", err)
	}
	defer rows.Close()

	attributes := make([]domain.Attribute, 0)
	for rows.Next() {
		var attribute domain.Attribute
		if err := rows.Scan(&attribute.ID, &attribute.Code, &attribute.MetricUnit); err != nil {
			return nil, fmt.Errorf("scan attribute: %w", err)
		}
		attributes = append(attributes, attribute)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate attributes: %w", err)
	}

	return attributes, nil
}
