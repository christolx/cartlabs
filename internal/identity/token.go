package identity

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type tokenManager struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
	now      func() time.Time
}

type accessClaims struct {
	Role Role `json:"role"`
	jwt.RegisteredClaims
}

func (m tokenManager) issue(user User) (string, int, error) {
	now := m.now().UTC()
	expires := now.Add(m.ttl)
	claims := accessClaims{
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   user.ID,
			Audience:  jwt.ClaimStrings{m.audience},
			ExpiresAt: jwt.NewNumericDate(expires),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	return signed, int(m.ttl.Seconds()), err
}

func (m tokenManager) parse(raw string) (Principal, error) {
	claims := &accessClaims{}
	parsed, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer), jwt.WithAudience(m.audience), jwt.WithExpirationRequired(), jwt.WithLeeway(5*time.Second), jwt.WithTimeFunc(m.now))
	if err != nil || !parsed.Valid || claims.Subject == "" {
		return Principal{}, ErrInvalidCredentials
	}
	if claims.Role != RoleBuyer && claims.Role != RoleSeller && claims.Role != RoleAdmin {
		return Principal{}, ErrInvalidCredentials
	}
	return Principal{UserID: claims.Subject, Role: claims.Role}, nil
}

func newOpaqueToken() (string, []byte, error) {
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return "", nil, fmt.Errorf("generate refresh token: %w", err)
	}
	raw := base64.RawURLEncoding.EncodeToString(random)
	hash := sha256.Sum256([]byte(raw))
	return raw, hash[:], nil
}

func hashOpaqueToken(raw string) []byte {
	hash := sha256.Sum256([]byte(raw))
	return hash[:]
}
