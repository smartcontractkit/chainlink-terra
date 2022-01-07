// Code generated by mockery v2.8.0. DO NOT EDIT.

package mocks

import (
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	abcitypes "github.com/tendermint/tendermint/abci/types"

	client "github.com/smartcontractkit/chainlink-terra/pkg/terra/client"

	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"

	mock "github.com/stretchr/testify/mock"

	tx "github.com/cosmos/cosmos-sdk/types/tx"

	types "github.com/cosmos/cosmos-sdk/types"
)

// ReaderWriter is an autogenerated mock type for the ReaderWriter type
type ReaderWriter struct {
	mock.Mock
}

// Account provides a mock function with given fields: address
func (_m *ReaderWriter) Account(address types.AccAddress) (authtypes.AccountI, error) {
	ret := _m.Called(address)

	var r0 authtypes.AccountI
	if rf, ok := ret.Get(0).(func(types.AccAddress) authtypes.AccountI); ok {
		r0 = rf(address)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(authtypes.AccountI)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(types.AccAddress) error); ok {
		r1 = rf(address)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Block provides a mock function with given fields: height
func (_m *ReaderWriter) Block(height *int64) (*coretypes.ResultBlock, error) {
	ret := _m.Called(height)

	var r0 *coretypes.ResultBlock
	if rf, ok := ret.Get(0).(func(*int64) *coretypes.ResultBlock); ok {
		r0 = rf(height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultBlock)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*int64) error); ok {
		r1 = rf(height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GasPrice provides a mock function with given fields:
func (_m *ReaderWriter) GasPrice() types.DecCoin {
	ret := _m.Called()

	var r0 types.DecCoin
	if rf, ok := ret.Get(0).(func() types.DecCoin); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(types.DecCoin)
	}

	return r0
}

// QueryABCI provides a mock function with given fields: path, params
func (_m *ReaderWriter) QueryABCI(path string, params client.ABCIQueryParams) (abcitypes.ResponseQuery, error) {
	ret := _m.Called(path, params)

	var r0 abcitypes.ResponseQuery
	if rf, ok := ret.Get(0).(func(string, client.ABCIQueryParams) abcitypes.ResponseQuery); ok {
		r0 = rf(path, params)
	} else {
		r0 = ret.Get(0).(abcitypes.ResponseQuery)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, client.ABCIQueryParams) error); ok {
		r1 = rf(path, params)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SignAndBroadcast provides a mock function with given fields: msgs, accountNum, sequence, gasPrice, signer, mode
func (_m *ReaderWriter) SignAndBroadcast(msgs []types.Msg, accountNum uint64, sequence uint64, gasPrice types.DecCoin, signer cryptotypes.PrivKey, mode tx.BroadcastMode) (*types.TxResponse, error) {
	ret := _m.Called(msgs, accountNum, sequence, gasPrice, signer, mode)

	var r0 *types.TxResponse
	if rf, ok := ret.Get(0).(func([]types.Msg, uint64, uint64, types.DecCoin, cryptotypes.PrivKey, tx.BroadcastMode) *types.TxResponse); ok {
		r0 = rf(msgs, accountNum, sequence, gasPrice, signer, mode)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.TxResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]types.Msg, uint64, uint64, types.DecCoin, cryptotypes.PrivKey, tx.BroadcastMode) error); ok {
		r1 = rf(msgs, accountNum, sequence, gasPrice, signer, mode)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TxSearch provides a mock function with given fields: query
func (_m *ReaderWriter) TxSearch(query string) (*coretypes.ResultTxSearch, error) {
	ret := _m.Called(query)

	var r0 *coretypes.ResultTxSearch
	if rf, ok := ret.Get(0).(func(string) *coretypes.ResultTxSearch); ok {
		r0 = rf(query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*coretypes.ResultTxSearch)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(query)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
