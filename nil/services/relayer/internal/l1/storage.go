package l1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/internal/db"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/jonboulle/clockwork"
)

const (
	// pendingEventsTable stores events received from L1BridgeMessenger waiting to be finalized
	// Key: Hash of the Event
	pendingEventsTable = "pending_events"

	// monotonic counter managing ordering between received events
	pendingEventsSequencer = "pending_events_sequencer"

	// lastProcessedBlockTable stores number (and some meta-info)
	// for last block events from which were successfully stored to the local database (single value)
	// Key: lastProcessedBlockKey
	lastProcessedBlockTable = "last_processed_block"
	lastProcessedBlockKey   = "last_processed_block_key"
)

type EventStorageMetrics interface {
	// TODO(oclaw)
}

type EventStorage struct {
	database        db.DB
	retryRunner     common.RetryRunner
	clock           clockwork.Clock
	metrics         EventStorageMetrics
	eventsSequencer db.Sequence
}

func NewEventStorage(
	ctx context.Context,
	database db.DB,
	clock clockwork.Clock,
	metrics EventStorageMetrics,
	logger logging.Logger,
) (*EventStorage, error) {
	es := &EventStorage{
		database: database,
		retryRunner: common.NewRetryRunner(
			common.RetryConfig{
				ShouldRetry: common.ComposeRetryPolicies(
					common.LimitRetries(10),
					common.DoNotRetryIf(ErrKeyExists, ErrSerializationFailed),
				),
				NextDelay: common.DelayJitter(20*time.Millisecond, 100*time.Millisecond, logger),
			},
			logger,
		),
		clock:   clock,
		metrics: metrics,
	}
	var err error
	es.eventsSequencer, err = database.GetSequence(ctx, []byte(pendingEventsSequencer), 100)
	if err != nil {
		return nil, err
	}

	return es, nil
}

func (es *EventStorage) StoreEvent(ctx context.Context, evt *Event) error {
	var emptyHash ethcommon.Hash
	if evt.Hash == emptyHash {
		return errors.New("cannot store event without hash")
	}

	return es.retryRunner.Do(ctx, func(ctx context.Context) error {
		var err error
		evt.SequenceNumber, err = es.eventsSequencer.Next()
		if err != nil {
			return err
		}

		writer := jsonDbWriter[*Event]{
			table:   pendingEventsTable,
			storage: es,
			upsert:  false,
		}

		return writer.putTx(ctx, evt.Hash.Bytes(), evt)

		// TODO (oclaw) metrics
	})
}

func (es *EventStorage) IterateEventsByBatch(
	ctx context.Context,
	batchSize int,
	callback func([]*Event) error,
) error {
	return es.retryRunner.Do(ctx, func(ctx context.Context) error {
		tx, err := es.database.CreateRoTx(ctx)
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
				return fmt.Errorf("%w: %w", ErrSerializationFailed, err)
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

func (es *EventStorage) DeleteEvents(ctx context.Context, hashes []common.Hash) error {
	return es.retryRunner.Do(ctx, func(ctx context.Context) error {
		tx, err := es.database.CreateRwTx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback()

		for _, hash := range hashes {
			if err := tx.Delete(pendingEventsTable, hash.Bytes()); err != nil && !errors.Is(err, db.ErrKeyNotFound) {
				return err
			}
		}

		return es.commit(tx)
	})
}

func (es *EventStorage) GetLastProcessedBlock(ctx context.Context) (*ProcessedBlock, error) {
	var ret *ProcessedBlock
	err := es.retryRunner.Do(ctx, func(ctx context.Context) error {
		tx, err := es.database.CreateRoTx(ctx)
		if err != nil {
			return err
		}

		data, err := tx.Get(lastProcessedBlockTable, []byte(lastProcessedBlockKey))
		if errors.Is(err, db.ErrKeyNotFound) {
			return nil
		}
		if err != nil {
			return err
		}

		var blk ProcessedBlock
		if err := json.Unmarshal(data, &blk); err != nil {
			return fmt.Errorf("%w: %w", ErrSerializationFailed, err)
		}

		ret = &blk

		return nil
	})
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (es *EventStorage) SetLastProcessedBlock(ctx context.Context, blk *ProcessedBlock) error {
	var emptyHash ethcommon.Hash
	if blk.BlockHash == emptyHash {
		return errors.New("empty last processed block hash")
	}
	if blk.BlockNumber == 0 {
		return errors.New("empty last processed block number")
	}

	return es.retryRunner.Do(ctx, func(ctx context.Context) error {
		writer := jsonDbWriter[*ProcessedBlock]{
			table:   lastProcessedBlockTable,
			storage: es,
			upsert:  true,
		}
		return writer.putTx(ctx, []byte(lastProcessedBlockKey), blk)

		// TODO(oclaw) metrics
	})
}

func (*EventStorage) commit(tx db.RwTx) error {
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

type jsonDbWriter[T any] struct {
	table   db.TableName
	storage *EventStorage
	upsert  bool
}

func (jdwr *jsonDbWriter[T]) putTx(ctx context.Context, key []byte, value T) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSerializationFailed, err)
	}

	tx, err := jdwr.storage.database.CreateRwTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if !jdwr.upsert {
		exists, err := tx.Exists(jdwr.table, key)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("%w: table=%s key=%v", ErrKeyExists, jdwr.table, key)
		}
	}

	if err := tx.Put(jdwr.table, key, data); err != nil {
		return err
	}

	return jdwr.storage.commit(tx)
}
