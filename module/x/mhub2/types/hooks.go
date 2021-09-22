package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type MhubHooks interface {
	AfterContractCallExecutedEvent(ctx sdk.Context, event ContractCallExecutedEvent)
	AfterSignerSetExecutedEvent(ctx sdk.Context, event SignerSetTxExecutedEvent)
	AfterBatchExecutedEvent(ctx sdk.Context, event BatchExecutedEvent)
	AfterSendToHubEvent(ctx sdk.Context, event SendToHubEvent)
}

type MultiGravityHooks []MhubHooks

func NewMultiGravityHooks(hooks ...MhubHooks) MultiGravityHooks {
	return hooks
}

func (mghs MultiGravityHooks) AfterContractCallExecutedEvent(ctx sdk.Context, event ContractCallExecutedEvent) {
	for i := range mghs {
		mghs[i].AfterContractCallExecutedEvent(ctx, event)
	}
}

func (mghs MultiGravityHooks) AfterSignerSetExecutedEvent(ctx sdk.Context, event SignerSetTxExecutedEvent) {
	for i := range mghs {
		mghs[i].AfterSignerSetExecutedEvent(ctx, event)
	}
}

func (mghs MultiGravityHooks) AfterBatchExecutedEvent(ctx sdk.Context, event BatchExecutedEvent) {
	for i := range mghs {
		mghs[i].AfterBatchExecutedEvent(ctx, event)
	}
}

func (mghs MultiGravityHooks) AfterSendToHubEvent(ctx sdk.Context, event SendToHubEvent) {
	for i := range mghs {
		mghs[i].AfterSendToHubEvent(ctx, event)
	}
}
