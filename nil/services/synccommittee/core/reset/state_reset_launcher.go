package reset

import (
	"context"
	"fmt"
	"time"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/rs/zerolog"
)

const (
	fetchResumeDelay        = 10 * time.Minute
	fetchResumeTimeout      = time.Minute
	gracefulShutdownTimeout = 5 * time.Minute
)

type BlockFetcher interface {
	Pause(ctx context.Context) error
	Resume(ctx context.Context) error
}

type Service interface {
	Stop() (stopped <-chan struct{})
}

type stateResetLauncher struct {
	blockFetcher BlockFetcher
	resetter     StateResetter
	service      Service
	logger       zerolog.Logger
}

func NewResetLauncher(
	blockFetcher BlockFetcher,
	resetter StateResetter,
	service Service,
	logger zerolog.Logger,
) *stateResetLauncher {
	return &stateResetLauncher{
		blockFetcher: blockFetcher,
		resetter:     resetter,
		service:      service,
		logger:       logger,
	}
}

func (l *stateResetLauncher) LaunchPartialResetWithSuspension(ctx context.Context, firstMainHashToPurge common.Hash) error {
	l.logger.Info().
		Stringer(logging.FieldBlockMainShardHash, firstMainHashToPurge).
		Msg("Launching state reset process")

	if err := l.blockFetcher.Pause(ctx); err != nil {
		return fmt.Errorf("failed to pause block fetching: %w", err)
	}

	if err := l.resetter.ResetProgressPartial(ctx, firstMainHashToPurge); err != nil {
		l.onResetError(ctx, err, firstMainHashToPurge)
		return nil
	}

	l.logger.Info().
		Stringer(logging.FieldBlockMainShardHash, firstMainHashToPurge).
		Msgf("State reset completed, block fetching will be resumed after %s", fetchResumeDelay)

	detachedCtx := context.WithoutCancel(ctx)
	time.AfterFunc(fetchResumeDelay, func() {
		l.resumeBlockFetching(detachedCtx)
	})
	return nil
}

func (l *stateResetLauncher) onResetError(
	ctx context.Context, resetErr error, failedMainBlockHash common.Hash,
) {
	l.logger.Error().Err(resetErr).Stringer(logging.FieldBlockMainShardHash, failedMainBlockHash).Send()
	l.resumeBlockFetching(ctx)
}

func (l *stateResetLauncher) resumeBlockFetching(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, fetchResumeTimeout)
	defer cancel()

	l.logger.Info().Msg("Resuming block fetching")
	err := l.blockFetcher.Resume(ctx)

	if err == nil {
		l.logger.Info().Msg("Block fetching successfully resumed")
		return
	}

	l.logger.Error().Err(err).Msg("Failed to resume block fetching, service will be terminated")

	stopped := l.service.Stop()

	select {
	case <-time.After(gracefulShutdownTimeout):
		l.logger.Fatal().Err(err).Msgf("Service did not stop after %s, force termination", gracefulShutdownTimeout)
	case <-stopped:
	}
}
