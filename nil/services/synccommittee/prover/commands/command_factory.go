package commands

import (
	"github.com/NilFoundation/nil/nil/services/synccommittee/internal/types"
	"github.com/rs/zerolog"
)

type CommandFactory struct {
	config CommandConfig
	logger zerolog.Logger
}

func NewCommandFactory(config CommandConfig, logger zerolog.Logger) *CommandFactory {
	return &CommandFactory{
		config: config,
		logger: logger,
	}
}

func (factory *CommandFactory) MakeHandlerCommandForTaskType(taskType types.TaskType) (Command, error) {
	switch taskType {
	case types.PartialProve:
		return NewPartialProofCmd(factory.config, factory.logger), nil
	case types.AggregatedChallenge:
		return NewAggregateChallengesCmd(factory.config), nil
	case types.CombinedQ:
		return NewCombinedQCmd(factory.config), nil
	case types.AggregatedFRI:
		return NewAggregateFRICmd(factory.config), nil
	case types.FRIConsistencyChecks:
		return NewConsistencyCheckCmd(factory.config), nil
	case types.MergeProof:
		return NewMergeProofCmd(factory.config), nil
	case types.AggregateProofs:
		return NewAggregateProofCmd(factory.config), nil
	case types.ProofBlock, types.ProofBatch:
		return nil, types.NewTaskErrNotSupportedType(taskType)
	case types.TaskTypeNone:
		return nil, types.NewTaskExecErrorf(types.TaskErrInvalidTask, "TaskType cannot be None")
	default:
		return nil, types.NewTaskErrNotSupportedType(taskType)
	}
}
