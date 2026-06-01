package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raymond/wyzauto-project/internal/domain"
	"github.com/raymond/wyzauto-project/internal/repository"
)

func TestIntegrationProductDocumentBuilder(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION") != "1" {
		t.Skip("set RUN_INTEGRATION=1 to run integration test")
	}

	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://wyzauto:wyzauto@localhost:5433/wyzauto?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}
	defer pool.Close()

	seedIntegrationData(t, ctx, pool)

	productRepo := repository.NewProductPostgresRepository(pool)
	translationRepo := repository.NewTranslationPostgresRepository(pool)
	loader := NewTranslationLoader(translationRepo, time.Minute)
	builder := NewProductDocumentBuilder(productRepo, loader, []string{"en", "th"})

	doc, err := builder.Build(ctx, "00000000-0000-0000-0000-000000000001")
	if err != nil {
		t.Fatalf("build document: %v", err)
	}

	if doc.SKU != "BP-OIL-5W30-1L" {
		t.Fatalf("unexpected sku: %s", doc.SKU)
	}

	if doc.Brand.Label["th"] != "บอช" {
		t.Fatalf("unexpected Thai brand label: %s", doc.Brand.Label["th"])
	}

	if doc.ProductName[1].Data != "น้ำมันเครื่อง 5W-30 1 ลิตร" {
		t.Fatalf("unexpected Thai product name: %s", doc.ProductName[1].Data)
	}

	oilGrade, ok := doc.Dynamic["oil_grade"]
	if !ok || oilGrade.Code != "5w30" {
		t.Fatalf("unexpected oil grade: %+v", doc.Dynamic)
	}

	if oilGrade.Label["en"] != "5W-30" {
		t.Fatalf("unexpected oil grade label: %+v", oilGrade.Label)
	}
}

func seedIntegrationData(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	statements := []string{
		`TRUNCATE translation, product_specification, attribute, product RESTART IDENTITY`,
		`INSERT INTO product (id, sku, part_number, brand, category_id) VALUES ('00000000-0000-0000-0000-000000000001', 'BP-OIL-5W30-1L', '5W30-1L', 'bosch', '00000000-0000-0000-0000-000000000010')`,
		`INSERT INTO attribute (id, code, metric_unit) VALUES ('00000000-0000-0000-0000-000000000002', 'oil_grade', NULL)`,
		`INSERT INTO product_specification (id, product_id, attribute_id, value) VALUES ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000002', '5w30')`,
		`INSERT INTO translation (entity_type, entity_id, locale, field_name, field_value, updated_at) VALUES
		 ('product', '00000000-0000-0000-0000-000000000001', 'en', 'productname', '5W-30 Engine Oil 1L', now()),
		 ('product', '00000000-0000-0000-0000-000000000001', 'th', 'productname', 'น้ำมันเครื่อง 5W-30 1 ลิตร', now()),
		 ('product', 'bosch', 'en', 'label', 'Bosch', now()),
		 ('product', 'bosch', 'th', 'label', 'บอช', now()),
		 ('product_specification', '00000000-0000-0000-0000-000000000003', 'en', 'value_label', '5W-30', now()),
		 ('product_specification', '00000000-0000-0000-0000-000000000003', 'th', 'value_label', '5W-30', now())`,
	}

	for _, statement := range statements {
		if _, err := pool.Exec(ctx, statement); err != nil {
			t.Fatalf("seed database: %v", err)
		}
	}
}

var _ = domain.ProductDocument{}
