package httpapi

import (
	"fmt"

	"github.com/christolx/cartlabs/internal/contract"
	"github.com/christolx/cartlabs/internal/identity"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func toDomainRole(value contract.Role) (identity.Role, error) {
	if !value.Valid() {
		return "", fmt.Errorf("map invalid role %q", value)
	}
	return identity.Role(value), nil
}

func toContractRole(value identity.Role) (contract.Role, error) {
	result := contract.Role(value)
	if !result.Valid() {
		return "", fmt.Errorf("map invalid role %q", value)
	}
	return result, nil
}

func toContractUser(value identity.User) (contract.User, error) {
	id, err := contractUUID(value.ID, "user.id")
	if err != nil {
		return contract.User{}, err
	}
	role, err := toContractRole(value.Role)
	if err != nil {
		return contract.User{}, err
	}
	return contract.User{Id: id, Email: openapi_types.Email(value.Email), DisplayName: value.DisplayName, Role: role}, nil
}

func toContractSession(value identity.Session) (contract.Session, error) {
	user, err := toContractUser(value.User)
	if err != nil {
		return contract.Session{}, err
	}
	tokenType := contract.SessionTokenType(value.TokenType)
	if !tokenType.Valid() {
		return contract.Session{}, fmt.Errorf("map invalid token type %q", value.TokenType)
	}
	return contract.Session{AccessToken: value.AccessToken, TokenType: tokenType, ExpiresIn: value.ExpiresIn, User: user}, nil
}

func toContractAdminUser(value identity.AdminUser) (contract.AdminUser, error) {
	id, err := contractUUID(value.ID, "adminUser.id")
	if err != nil {
		return contract.AdminUser{}, err
	}
	role, err := toContractRole(value.Role)
	if err != nil {
		return contract.AdminUser{}, err
	}
	status := contract.UserStatus(value.Status)
	if !status.Valid() {
		return contract.AdminUser{}, fmt.Errorf("map invalid user status %q", value.Status)
	}
	return contract.AdminUser{Id: id, Email: openapi_types.Email(value.Email), DisplayName: value.DisplayName,
		Role: role, Status: status, CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt}, nil
}

func toContractAdminUsers(values []identity.AdminUser) ([]contract.AdminUser, error) {
	result := make([]contract.AdminUser, len(values))
	for i := range values {
		mapped, err := toContractAdminUser(values[i])
		if err != nil {
			return nil, err
		}
		result[i] = mapped
	}
	return result, nil
}
