package httpapi

import (
	"fmt"

	"github.com/christolx/cartlabs/internal/contract"
	marketstore "github.com/christolx/cartlabs/internal/store"
)

func toDomainStoreInput(value contract.StoreInput) marketstore.Input {
	return marketstore.Input{Name: value.Name, Slug: value.Slug, Description: value.Description}
}

func toContractModerationStatus(value string) (contract.ModerationStatus, error) {
	result := contract.ModerationStatus(value)
	if !result.Valid() {
		return "", fmt.Errorf("map invalid moderation status %q", value)
	}
	return result, nil
}

func toContractStore(value marketstore.Store) (contract.Store, error) {
	id, err := contractUUID(value.ID, "store.id")
	if err != nil {
		return contract.Store{}, err
	}
	sellerID, err := contractUUID(value.SellerID, "store.sellerId")
	if err != nil {
		return contract.Store{}, err
	}
	status, err := toContractModerationStatus(value.Status)
	if err != nil {
		return contract.Store{}, err
	}
	return contract.Store{
		Id: id, SellerId: sellerID, Name: value.Name, Slug: value.Slug,
		Description: value.Description, Status: status, ModerationNote: value.ModerationNote,
		CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt,
	}, nil
}

func toContractStoreProfile(value marketstore.Profile) (contract.StoreProfile, error) {
	id, err := contractUUID(value.ID, "storeProfile.id")
	if err != nil {
		return contract.StoreProfile{}, err
	}
	return contract.StoreProfile{Id: id, Name: value.Name, Slug: value.Slug, Description: value.Description,
		SellerDisplayName: value.SellerDisplayName, CreatedAt: value.CreatedAt}, nil
}

func toContractStores(values []marketstore.Store) ([]contract.Store, error) {
	result := make([]contract.Store, len(values))
	for i := range values {
		mapped, err := toContractStore(values[i])
		if err != nil {
			return nil, err
		}
		result[i] = mapped
	}
	return result, nil
}
