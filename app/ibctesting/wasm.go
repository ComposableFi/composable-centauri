package ibctesting

import (
	wasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
)

// ConstructUpdateWasmClientHeader will construct a valid 08-wasm Header with a zero height
// to update the light client on the source chain.
func (chain *TestChain) ConstructUpdateWasmClientHeader(counterparty *TestChain, clientID string) (*wasmtypes.ClientMessage, error) {
	return chain.ConstructUpdateWasmClientHeaderWithTrustedHeight(counterparty, clientID, clienttypes.ZeroHeight())
}

// ConstructUpdateWasmClientHeaderWithTrustedHeight will construct a valid 08-wasm Header
// to update the light client on the source chain.
func (chain *TestChain) ConstructUpdateWasmClientHeaderWithTrustedHeight(counterparty *TestChain, clientID string, trustedHeight clienttypes.Height) (*wasmtypes.ClientMessage, error) {
	tmHeader, err := chain.ConstructUpdateTMClientHeaderWithTrustedHeight(counterparty, clientID, trustedHeight)
	if err != nil {
		return nil, err
	}

	tmWasmHeaderData, err := chain.Codec.MarshalInterface(tmHeader)
	if err != nil {
		return nil, err
	}

	wasmHeader := wasmtypes.ClientMessage{
		Data: tmWasmHeaderData,
	}

	return &wasmHeader, nil
}

// func (chain *TestChain) CreateWasmClientHeader(chainID string, blockHeight int64, trustedHeight clienttypes.Height, timestamp time.Time, tmValSet, _, tmTrustedVals *tmtypes.ValidatorSet, signers []tmtypes.PrivValidator) *wasmtypes.Header {
//	tmHeader := chain.CreateTMClientHeader(chainID, blockHeight, trustedHeight, timestamp, tmValSet, tmTrustedVals, signers)
//	tmWasmHeaderData, err := chain.Codec.MarshalInterface(tmHeader)
//	require.NoError(chain.t, err)
//	height, ok := tmHeader.GetHeight().(clienttypes.Height)
//	require.True(chain.t, ok)
//	return &wasmtypes.Header{
//		Data:   tmWasmHeaderData,
//		Height: height,
//	}
//}
