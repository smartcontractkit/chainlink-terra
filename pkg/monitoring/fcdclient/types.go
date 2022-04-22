package fcdclient

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Response struct {
	Txs []Tx `json:"txs"`
}

type Tx struct {
	ID     uint64 `json:"id"`
	Height uint64 `json:"height"`
	Code   int    `json:"code"` // Error code if present
	Logs   []Log  `json:"logs"`
}

type Log struct {
	Events []Event `json:"events"`
}

type Event struct {
	Typ        string      `json:"type"`
	Attributes []Attribute `json:"attributes"`
}

type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

//go:generate mockery --name Client --output ./mocks/

type Client interface {
	GetTxList(context.Context, GetTxListParams) (Response, error)
}

type GetTxListParams struct {
	Account sdk.AccAddress
	Block   string
	Offset  int
	Limit   int
}
