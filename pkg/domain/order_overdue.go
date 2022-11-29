package domain

import (
	"context"
	"time"

	"go.uber.org/zap"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/ent"
)

type OrderOverdueCheckup interface {
	Checkup(ctx context.Context, cln *ent.Client, logger *zap.Logger)
	PeriodicalCheckup(ctx context.Context, overdueTimeCheckDuration time.Duration, cln *ent.Client, logger *zap.Logger)
}
