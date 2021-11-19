package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type MhubHooks interface {
	AfterContractCallExecutedEvent(ctx sdk.Context, event ContractCallExecutedEvent)
	AfterSignerSetExecutedEvent(ctx sdk.Context, event SignerSetTxExecutedEvent)
	AfterBatchExecutedEvent(ctx sdk.Context, event BatchExecutedEvent)
	AfterSendToHubEvent(ctx sdk.Context, event SendToHubEvent)
}

type MultiMhub2Hooks []MhubHooks

func NewMultiMhub2Hooks(hooks ...MhubHooks) MultiMhub2Hooks {
	return hooks
}

func (mghs MultiMhub2Hooks) AfterContractCallExecutedEvent(ctx sdk.Context, event ContractCallExecutedEvent) {
	for i := range mghs {
		mghs[i].AfterContractCallExecutedEvent(ctx, event)
	}
}

func (mghs MultiMhub2Hooks) AfterSignerSetExecutedEvent(ctx sdk.Context, event SignerSetTxExecutedEvent) {
	for i := range mghs {
		mghs[i].AfterSignerSetExecutedEvent(ctx, event)
	}
}

func (mghs MultiMhub2Hooks) AfterBatchExecutedEvent(ctx sdk.Context, event BatchExecutedEvent) {
	for i := range mghs {
		mghs[i].AfterBatchExecutedEvent(ctx, event)
	}
}

func (mghs MultiMhub2Hooks) AfterSendToHubEvent(ctx sdk.Context, event SendToHubEvent) {
	for i := range mghs {
		mghs[i].AfterSendToHubEvent(ctx, event)
	}
}
