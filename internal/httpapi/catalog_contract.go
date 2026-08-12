package httpapi

import (
	"fmt"

	"github.com/christolx/cartlabs/internal/catalog"
	"github.com/christolx/cartlabs/internal/contract"
)

func toDomainProductInput(value contract.ProductInput) catalog.ProductInput {
	return catalog.ProductInput{CategoryID: value.CategoryId.String(), Name: value.Name, Slug: value.Slug, Description: value.Description}
}

func toDomainVariantInput(value contract.VariantInput) catalog.VariantInput {
	return catalog.VariantInput{SKU: value.Sku, Name: value.Name, Attributes: value.Attributes, PriceMinor: value.PriceMinor, Currency: string(value.Currency), Stock: value.Stock}
}

func toDomainImageInput(value contract.ImageInput) catalog.ImageInput {
	return catalog.ImageInput{URL: value.Url, AltText: value.AltText, Position: value.Position}
}

func toContractCategory(value catalog.Category) (contract.Category, error) {
	id, err := contractUUID(value.ID, "category.id")
	if err != nil {
		return contract.Category{}, err
	}
	return contract.Category{Id: id, Name: value.Name, Slug: value.Slug}, nil
}

func toContractCategories(values []catalog.Category) ([]contract.Category, error) {
	result := make([]contract.Category, len(values))
	for i := range values {
		mapped, err := toContractCategory(values[i])
		if err != nil {
			return nil, err
		}
		result[i] = mapped
	}
	return result, nil
}

func toContractVariant(value catalog.Variant) (contract.Variant, error) {
	id, err := contractUUID(value.ID, "variant.id")
	if err != nil {
		return contract.Variant{}, err
	}
	currency := contract.VariantCurrency(value.Currency)
	if !currency.Valid() {
		return contract.Variant{}, fmt.Errorf("map invalid variant currency %q", value.Currency)
	}
	attributes := value.Attributes
	if attributes == nil {
		attributes = map[string]string{}
	}
	return contract.Variant{Id: id, Sku: value.SKU, Name: value.Name, Attributes: attributes, PriceMinor: value.PriceMinor, Currency: currency, Stock: value.Stock, Active: value.Active}, nil
}

func toContractVariants(values []catalog.Variant) ([]contract.Variant, error) {
	result := make([]contract.Variant, len(values))
	for i := range values {
		mapped, err := toContractVariant(values[i])
		if err != nil {
			return nil, err
		}
		result[i] = mapped
	}
	return result, nil
}

func toContractProductImage(value catalog.ProductImage) (contract.ProductImage, error) {
	id, err := contractUUID(value.ID, "productImage.id")
	if err != nil {
		return contract.ProductImage{}, err
	}
	return contract.ProductImage{Id: id, Url: value.URL, AltText: value.AltText, Position: value.Position}, nil
}

func toContractProductImages(values []catalog.ProductImage) ([]contract.ProductImage, error) {
	result := make([]contract.ProductImage, len(values))
	for i := range values {
		mapped, err := toContractProductImage(values[i])
		if err != nil {
			return nil, err
		}
		result[i] = mapped
	}
	return result, nil
}

func toContractProduct(value catalog.Product) (contract.ProductDetail, error) {
	id, err := contractUUID(value.ID, "product.id")
	if err != nil {
		return contract.ProductDetail{}, err
	}
	storeID, err := contractUUID(value.StoreID, "product.storeId")
	if err != nil {
		return contract.ProductDetail{}, err
	}
	category, err := toContractCategory(value.Category)
	if err != nil {
		return contract.ProductDetail{}, err
	}
	variants, err := toContractVariants(value.Variants)
	if err != nil {
		return contract.ProductDetail{}, err
	}
	images, err := toContractProductImages(value.Images)
	if err != nil {
		return contract.ProductDetail{}, err
	}
	status := contract.ProductDetailStatus(value.Status)
	if !status.Valid() {
		return contract.ProductDetail{}, fmt.Errorf("map invalid product status %q", value.Status)
	}
	moderationStatus, err := toContractModerationStatus(value.ModerationStatus)
	if err != nil {
		return contract.ProductDetail{}, err
	}
	result := contract.ProductDetail{
		Id: id, StoreId: storeID, Category: category, Name: value.Name, Slug: value.Slug,
		Description: value.Description, Status: status, ModerationStatus: moderationStatus,
		ModerationNote: value.ModerationNote, Variants: variants, Images: images,
		CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt,
	}
	if value.StoreName != "" {
		result.StoreName = &value.StoreName
	}
	return result, nil
}

func toContractProducts(values []catalog.Product) ([]contract.ProductDetail, error) {
	result := make([]contract.ProductDetail, len(values))
	for i := range values {
		mapped, err := toContractProduct(values[i])
		if err != nil {
			return nil, err
		}
		result[i] = mapped
	}
	return result, nil
}

func toContractProductSummary(value catalog.Summary) (contract.ProductSummary, error) {
	id, err := contractUUID(value.ID, "productSummary.id")
	if err != nil {
		return contract.ProductSummary{}, err
	}
	category, err := toContractCategory(value.Category)
	if err != nil {
		return contract.ProductSummary{}, err
	}
	return contract.ProductSummary{Id: id, Name: value.Name, Slug: value.Slug, StoreName: value.StoreName, Category: category, MinPriceMinor: value.MinPriceMinor, Currency: value.Currency, InStock: value.InStock, ImageUrl: value.ImageURL}, nil
}

func toContractProductPage(value catalog.Page) (contract.ProductPage, error) {
	items := make([]contract.ProductSummary, len(value.Items))
	for i := range value.Items {
		mapped, err := toContractProductSummary(value.Items[i])
		if err != nil {
			return contract.ProductPage{}, err
		}
		items[i] = mapped
	}
	return contract.ProductPage{Items: items, Page: value.Page, PageSize: value.PageSize, Total: value.Total}, nil
}
