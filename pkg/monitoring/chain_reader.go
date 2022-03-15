package monitoring

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
)

// ChainReader is a subset of the pkg/terra/client.Reader interface enhanced with context support.
type ChainReader interface {
	TxsEvents(ctx context.Context, events []string, paginationParams *query.PageRequest) (*txtypes.GetTxsEventResponse, error)
	ContractStore(ctx context.Context, contractAddress sdk.AccAddress, queryMsg []byte) ([]byte, error)
}
