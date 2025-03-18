package l2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/services/relayer/internal/storage"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/jonboulle/clockwork"
)

const (
	// pendingEventsTable stores events that are finalized on L1 and ready to be forwarded to L2
	// Key: Hash of the Event
	pendingEventsTable = "pending_l2_events"
)

type EventStorageMetrics interface {
	// TODO(oclaw)
}

type EventStorage struct {
	*storage.BaseStorage
	metrics EventStorageMetrics
}

func NewEventStorage(
	ctx context.Context,
	database db.DB,
	clock clockwork.Clock,
	metrics EventStorageMetrics,
	logger logging.Logger,
) *EventStorage {
	es := &EventStorage{
		BaseStorage: storage.NewBaseStorage(ctx, database, clock, logger),
		metrics:     metrics,
	}
	return es
}

func (es *EventStorage) StoreEvents(ctx context.Context, evts []*Event) error {
	var emptyHash ethcommon.Hash
	for _, evt := range evts {
		if evt.Hash == emptyHash {
			return errors.New("cannot store event without hash")
		}
	}

	return es.RetryRunner.Do(ctx, func(ctx context.Context) error {
		writer := storage.NewJSONWriter[*Event](pendingEventsTable, es.BaseStorage, false)
		reqs := storage.MakeInsertRequests(
			evts,
			func(e *Event) []byte {
				return e.Hash[:]
			},
		)
		return writer.PutManyTx(ctx, reqs)

		// TODO (oclaw) metrics
	})
}

func (es *EventStorage) IterateEventsByBatch(
	ctx context.Context,
	batchSize int,
	callback func([]*Event) error,
) error {
	return es.RetryRunner.Do(ctx, func(ctx context.Context) error {
		tx, err := es.Database.CreateRoTx(ctx)
		if err != nil {
			return err
		}

		iter, err := tx.Range(pendingEventsTable, nil, nil)
		if err != nil {
			return err
		}

		batch := make([]*Event, batchSize)
		idx := 0
		for iter.HasNext() {
			_, val, err := iter.Next()
			if err != nil {
				return err
			}
			if err := json.Unmarshal(val, &batch[idx]); err != nil {
				return fmt.Errorf("%w: %w", storage.ErrSerializationFailed, err)
			}

			idx++
			if idx >= batchSize {
				if err := callback(batch); err != nil {
					return err
				}
				idx = 0
			}
		}
		if idx > 0 {
			return callback(batch[:idx])
		}

		return nil
	})
}

func (es *EventStorage) DeleteEvents(ctx context.Context, hashes []ethcommon.Hash) error {
	return es.RetryRunner.Do(ctx, func(ctx context.Context) error {
		tx, err := es.Database.CreateRwTx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		for _, hash := range hashes {
			if err := tx.Delete(pendingEventsTable, hash.Bytes()); err != nil && !errors.Is(err, db.ErrKeyNotFound) {
				return err
			}
		}

		return es.Commit(tx)
	})
}

func (*EventStorage) Commit(tx db.RwTx) error {
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
