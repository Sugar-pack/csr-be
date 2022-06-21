package repositories

import (
	"context"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent/token"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/ent"
)

type TokenRepository interface {
	CreateTokens(ctx context.Context, ownerID int, accessToken, refreshToken string) error
	DeleteTokensByRefreshToken(ctx context.Context, refreshToken string) error
	UpdateAccessToken(ctx context.Context, accessToken, refreshToken string) error
}

type tokenRepository struct {
	client *ent.Client
}

func (t *tokenRepository) UpdateAccessToken(ctx context.Context, accessToken, refreshToken string) error {
	_, err := t.client.Token.Update().Where(token.RefreshToken(refreshToken)).SetAccessToken(accessToken).Save(ctx)
	if err != nil {
		return err
	}
	return nil
}

func NewTokenRepository(client *ent.Client) TokenRepository {
	return &tokenRepository{client: client}
}

func (t *tokenRepository) DeleteTokensByRefreshToken(ctx context.Context, refreshToken string) error {
	_, err := t.client.Token.Delete().Where(token.RefreshTokenEQ(refreshToken)).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (t *tokenRepository) CreateTokens(ctx context.Context, ownerID int, accessToken, refreshToken string) error {
	_, err := t.client.Token.
		Create().
		SetOwnerID(ownerID).
		SetAccessToken(accessToken).
		SetRefreshToken(refreshToken).
		Save(ctx)
	return err
}
