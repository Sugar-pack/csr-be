package repositories

import (
	"context"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent/token"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/middlewares"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

type tokenRepository struct {
}

func (t *tokenRepository) UpdateAccessToken(ctx context.Context, accessToken, refreshToken string) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	_, err = tx.Token.Update().Where(token.RefreshToken(refreshToken)).SetAccessToken(accessToken).Save(ctx)
	if err != nil {
		return err
	}
	return nil
}

func NewTokenRepository() domain.TokenRepository {
	return &tokenRepository{}
}

func (t *tokenRepository) DeleteTokensByRefreshToken(ctx context.Context, refreshToken string) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	q, err := tx.Token.Delete().Where(token.RefreshTokenEQ(refreshToken)).Exec(ctx)
	if err != nil {
		return err
	}
	if q == 0 {
		return &ent.NotFoundError{}
	}
	return nil
}

func (t *tokenRepository) CreateTokens(ctx context.Context, ownerID int, accessToken, refreshToken string) error {
	tx, err := middlewares.TxFromContext(ctx)
	if err != nil {
		return err
	}
	_, err = tx.Token.
		Create().
		SetOwnerID(ownerID).
		SetAccessToken(accessToken).
		SetRefreshToken(refreshToken).
		Save(ctx)
	return err
}
