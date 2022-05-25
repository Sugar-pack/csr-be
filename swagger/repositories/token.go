package repositories

import (
	"context"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
)

type TokenRepository interface {
	CreateTokens(ctx context.Context, ownerID int, accessToken, refreshToken string) error
}

type tokenRepository struct {
	client *ent.Client
}

func NewTokenRepository(client *ent.Client) TokenRepository {
	return &tokenRepository{client: client}
}
func (t tokenRepository) CreateTokens(ctx context.Context, ownerID int, accessToken, refreshToken string) error {
	_, err := t.client.Token.
		Create().
		SetOwnerID(ownerID).
		SetAccessToken(accessToken).
		SetRefreshToken(refreshToken).
		Save(ctx)
	return err
}
