package jsonrpc

import (
	"context"
	"github.com/NilFoundation/nil/nil/internal/db"
	"github.com/NilFoundation/nil/nil/services/rpc/rawapi"
	"github.com/NilFoundation/nil/nil/services/txnpool"
	"testing"

	"github.com/NilFoundation/nil/nil/common"
	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/NilFoundation/nil/nil/internal/config"
	"github.com/NilFoundation/nil/nil/internal/execution"
	"github.com/NilFoundation/nil/nil/internal/types"
	"github.com/stretchr/testify/suite"
)

type SuiteTxnPoolApi struct {
	SuiteAccountsBase
	txnpoolApi *TxPoolAPIImpl
	pools      map[types.ShardId]txnpool.Pool
}

const defaultMaxFee = 500

var defaultAddress = types.ShardAndHexToAddress(types.BaseShardId, "11")
var defaultBaseFee = types.NewValueFromUint64(100)

func newTransaction(address types.Address, seqno types.Seqno, priorityFee uint64) *types.Transaction {
	return &types.Transaction{
		TransactionDigest: types.TransactionDigest{
			To:                   address,
			Seqno:                seqno,
			MaxPriorityFeePerGas: types.NewValueFromUint64(priorityFee),
			MaxFeePerGas:         types.NewValueFromUint64(defaultMaxFee),
		},
	}
}

func (suite *SuiteTxnPoolApi) SetupSuite() {
	suite.SuiteAccountsBase.SetupSuite()

	shardId := types.BaseShardId
	ctx := context.Background()

	var err error
	tx, err := suite.db.CreateRwTx(ctx)
	suite.Require().NoError(err)
	defer tx.Rollback()

	es, err := execution.NewExecutionState(tx, shardId, execution.StateParams{
		ConfigAccessor: config.GetStubAccessor(),
	})
	suite.Require().NoError(err)
	es.BaseFee = types.DefaultGasPrice

	suite.smcAddr = types.GenerateRandomAddress(shardId)
	suite.Require().NotEmpty(suite.smcAddr)

	suite.Require().NoError(es.CreateAccount(suite.smcAddr))
	suite.Require().NoError(es.SetCode(suite.smcAddr, []byte("some code")))
	suite.Require().NoError(es.SetState(suite.smcAddr, common.Hash{0x1}, common.IntToHash(2)))
	suite.Require().NoError(es.SetState(suite.smcAddr, common.Hash{0x3}, common.IntToHash(4)))

	suite.Require().NoError(es.SetBalance(suite.smcAddr, types.NewValueFromUint64(1234)))
	suite.Require().NoError(es.SetExtSeqno(suite.smcAddr, 567))

	blockRes, err := es.Commit(0, nil)
	suite.Require().NoError(err)
	suite.blockHash = blockRes.BlockHash

	err = execution.PostprocessBlock(tx, shardId, blockRes)
	suite.Require().NotNil(blockRes.Block)
	suite.Require().NoError(err)

	err = tx.Commit()
	suite.Require().NoError(err)

	//shardApi := rawapi.NewLocalShardApi(shardId, suite.db, nil)
	//localShardApis := map[types.ShardId]rawapi.ShardApi{
	//	shardId: shardApi,
	//}
	suite.pools = NewPools(suite.T(), ctx, 2)
	//localApi := rawapi.NewNodeApiOverShardApis(localShardApis)
	database, err := db.NewBadgerDbInMemory()
	suite.Require().NoError(err)
	defer database.Close()
	mainShardApi := rawapi.NewLocalShardApi(types.MainShardId, database, suite.pools[0])
	localShardApis := map[types.ShardId]rawapi.ShardApi{
		types.MainShardId: mainShardApi,
	}
	localApi := rawapi.NewNodeApiOverShardApis(localShardApis)

	suite.txnpoolApi = NewTxPoolAPI(localApi, logging.NewLogger("Test"))
	suite.Require().NoError(err)
}

func (suite *SuiteTxnPoolApi) TearDownSuite() {
	suite.SuiteAccountsBase.TearDownSuite()
}

func (suite *SuiteTxnPoolApi) TestGetTxnpoolStatus() {
	txAmount := 10
	ctx := context.Background()
	shardTxn := newTransaction(defaultAddress, 0, 123)
	shardTxn.To = types.ShardAndHexToAddress(0, "deadbeef")

	for i := 0; i < txAmount; i++ {
		_, err := suite.pools[0].Add(ctx, shardTxn)
		suite.Require().NoError(err)
	}
	txAmountRes, err := suite.txnpoolApi.GetTxpoolStatus(ctx, types.MainShardId)
	suite.Require().NoError(err)
	suite.Require().Equal(uint64(txAmount), txAmountRes)
}

func TestSuiteTxnPoolApi(t *testing.T) {
	t.Parallel()

	suite.Run(t, new(SuiteTxnPoolApi))
}
