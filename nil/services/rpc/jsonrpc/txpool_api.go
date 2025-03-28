package jsonrpc

import (
	"context"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/NilFoundation/nil/nil/services/rpc/rawapi"
	"github.com/rs/zerolog"
)

type TxPoolAPI interface {
	GetTxpoolStatus(ctx context.Context, shardId types.ShardId) (uint64, error)
	GetTxpoolContent(ctx context.Context) ([]*types.TxnWithHash, error)
}

type TxPoolAPIImpl struct {
	logger zerolog.Logger
	rawApi rawapi.NodeApi
}

var _ TxPoolAPI = &TxPoolAPIImpl{}

func NewTxPoolAPI(rawApi rawapi.NodeApi, logger zerolog.Logger) *TxPoolAPIImpl {
	return &TxPoolAPIImpl{
		logger: logger,
		rawApi: rawApi,
	}
}

func (api *TxPoolAPIImpl) GetTxpoolStatus(ctx context.Context, shardId types.ShardId) (uint64, error) {
	return api.rawApi.GetTxpoolStatus(ctx, shardId)
}

func (api *TxPoolAPIImpl) GetTxpoolContent(cxt context.Context) ([]*types.TxnWithHash, error) {
	//return api.txnPool[0].Peek(api.txnPool[0].GetQueue().Len())
	return []*types.TxnWithHash{}, nil
}
