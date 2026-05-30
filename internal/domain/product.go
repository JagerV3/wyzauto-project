package domain

type Product struct {
	ID         string
	SKU        string
	PartNumber string
	Brand      string
	CategoryID string
}

type Attribute struct {
	ID         string
	Code       string
	MetricUnit *string
}

type ProductSpecification struct {
	ID          string
	ProductID   string
	AttributeID string
	Value       string
}

type BrandDocument struct {
	Code  string            `json:"code"`
	Label map[string]string `json:"label"`
}

type ProductNameDocument struct {
	Locale string `json:"locale"`
	Data   string `json:"data"`
}

type AttributeValueDocument struct {
	Code  string            `json:"code"`
	Label map[string]string `json:"label"`
}

type ProductDocument struct {
	UUID        string                            `json:"uuid"`
	SKU         string                            `json:"sku"`
	PartNumber  string                            `json:"part_number"`
	Brand       BrandDocument                     `json:"brand"`
	ProductName []ProductNameDocument             `json:"productname"`
	OilGrade    *AttributeValueDocument           `json:"oil_grade,omitempty"`
	Attributes  map[string]string                 `json:"attributes"`
	Dynamic     map[string]AttributeValueDocument `json:"-"`
}
