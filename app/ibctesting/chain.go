package ibctesting

import (
	"context"
	"fmt"
	"testing"
	"time"

	wasm08 "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/keeper"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/keeper"
	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	customibctransferkeeper "github.com/notional-labs/composable/v6/custom/ibc-transfer/keeper"
	transfermiddlewarekeeper "github.com/notional-labs/composable/v6/x/transfermiddleware/keeper"

	ratelimitmodulekeeper "github.com/notional-labs/composable/v6/x/ratelimit/keeper"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/tmhash"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	tmversion "github.com/cometbft/cometbft/version"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	commitmenttypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	"github.com/cosmos/ibc-go/v8/modules/core/types"
	ibctmtypes "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	cmtprotoversion "github.com/cometbft/cometbft/proto/tendermint/version"
	"github.com/notional-labs/composable/v6/app"
)

var MaxAccounts = 10

type SenderAccount struct {
	SenderPrivKey cryptotypes.PrivKey
	SenderAccount sdk.AccountI
}

// ChainApp Abstract chain app definition used for testing
type ChainApp interface {
	servertypes.ABCI
	AppCodec() codec.Codec
	GetContextForFinalizeBlock(txBytes []byte) sdk.Context
	NewContextLegacy(isCheckTx bool, header tmproto.Header) sdk.Context
	NewUncachedContext(isCheckTx bool, header tmproto.Header) sdk.Context
	LastBlockHeight() int64
	LastCommitID() storetypes.CommitID
	GetBaseApp() *baseapp.BaseApp

	TxConfig() client.TxConfig
	GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper
	GetIBCKeeper() *ibckeeper.Keeper
	GetBankKeeper() bankkeeper.Keeper
	GetStakingKeeper() *stakingkeeper.Keeper
	GetGovKeeper() *govkeeper.Keeper
	GetAccountKeeper() authkeeper.AccountKeeper
	GetWasmKeeper() wasmkeeper.Keeper
	GetWasm08Keeper() wasm08.Keeper
	GetPfmKeeper() packetforwardkeeper.Keeper
	GetRateLimitKeeper() ratelimitmodulekeeper.Keeper
	GetTransferMiddlewareKeeper() transfermiddlewarekeeper.Keeper
	GetTransferKeeper() customibctransferkeeper.Keeper
}

// TestChain is a testing struct that wraps a simapp with the last TM Header, the current ABCI
// header and the validators of the TestChain. It also contains a field called ChainID. This
// is the clientID that *other* chains use to refer to this TestChain. The SenderAccount
// is used for delivering transactions through the application state.
// NOTE: the actual application uses an empty chain-id for ease of testing.
type TestChain struct {
	t *testing.T

	Coordinator   *Coordinator
	App           ChainApp
	ChainID       string
	LastHeader    *ibctmtypes.Header // header for last block height committed
	CurrentHeader tmproto.Header     // header for current block height
	QueryServer   types.QueryServer
	TxConfig      client.TxConfig
	Codec         codec.BinaryCodec

	Vals     *cmttypes.ValidatorSet
	NextVals *cmttypes.ValidatorSet
	Signers  map[string]cmttypes.PrivValidator

	SenderPrivKey  cryptotypes.PrivKey
	SenderAccount  sdk.AccountI
	SenderAccounts []SenderAccount

	PendingSendPackets []channeltypes.Packet
	PendingAckPackets  []PacketAck

	DefaultMsgFees sdk.Coins

	// Use wasm client if true
	UseWasmClient bool
}

type PacketAck struct {
	Packet channeltypes.Packet
	Ack    []byte
}

// ChainAppFactory abstract factory method that usually implemented by app.SetupWithGenesisValSet
type ChainAppFactory func(t *testing.T, valSet *cmttypes.ValidatorSet, genAccs []authtypes.GenesisAccount, chainID string, balances ...banktypes.Balance) ChainApp

// NewTestChain initializes a new TestChain instance with a single validator set using a
// generated private key. It also creates a sender account to be used for delivering transactions.
//
// The first block height is committed to state in order to allow for client creations on
// counterparty chains. The TestChain will return with a block height starting at 2.
//
// Time management is handled by the Coordinator in order to ensure synchrony between chains.
// Each update of any chain increments the block header time for all chains by 5 seconds.
func NewTestChain(t *testing.T, coord *Coordinator, appFactory ChainAppFactory, chainID string) *TestChain {
	t.Helper()
	genAccs := []authtypes.GenesisAccount{}
	genBals := []banktypes.Balance{}
	senderAccs := []SenderAccount{}

	// generate validators private/public key
	var (
		validatorsPerChain = 4
		validators         []*cmttypes.Validator
		signersByAddress   = make(map[string]cmttypes.PrivValidator, validatorsPerChain)
	)

	for i := 0; i < validatorsPerChain; i++ {
		_, privVal := cmttypes.RandValidator(false, 100)
		pubKey, err := privVal.GetPubKey()
		require.NoError(t, err)
		validators = append(validators, cmttypes.NewValidator(pubKey, 1))
		signersByAddress[pubKey.Address().String()] = privVal
	}

	// construct validator set;
	// Note that the validators are sorted by voting power
	// or, if equal, by address lexical order
	valSet := cmttypes.NewValidatorSet(validators)

	// generate genesis accounts
	for i := 0; i < MaxAccounts; i++ {
		senderPrivKey := secp256k1.GenPrivKey()
		acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), uint64(i), 0)
		amount, ok := sdkmath.NewIntFromString("10000000000000000000")
		require.True(t, ok)

		// add sender account
		balance := banktypes.Balance{
			Address: acc.GetAddress().String(),
			Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amount)),
		}

		genAccs = append(genAccs, acc)
		genBals = append(genBals, balance)

		senderAcc := SenderAccount{
			SenderAccount: acc,
			SenderPrivKey: senderPrivKey,
		}

		senderAccs = append(senderAccs, senderAcc)
	}

	app := appFactory(t, valSet, genAccs, chainID, genBals...)

	// app := NewTestingAppDecorator(t, app.SetupWithGenesisValSet(t, valSet, []authtypes.GenesisAccount{acc}, "", nil, balance))

	// create current header and call begin block
	header := tmproto.Header{
		ChainID: chainID,
		Height:  1,
		Time:    coord.CurrentTime.UTC(),
	}

	txConfig := app.TxConfig()
	// create an account to send transactions from
	chain := &TestChain{
		t:              t,
		Coordinator:    coord,
		ChainID:        chainID,
		App:            app,
		CurrentHeader:  header,
		QueryServer:    app.GetIBCKeeper(),
		TxConfig:       txConfig,
		Codec:          app.AppCodec(),
		Vals:           valSet,
		NextVals:       valSet,
		Signers:        signersByAddress,
		SenderPrivKey:  senderAccs[0].SenderPrivKey,
		SenderAccount:  senderAccs[0].SenderAccount,
		SenderAccounts: senderAccs,
		DefaultMsgFees: sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.ZeroInt())),
	}

	baseapp.SetChainID(chain.ChainID)(chain.App.GetBaseApp())
	coord.CommitBlock(chain)

	return chain
}

// SetWasm
func (chain *TestChain) SetWasm(wasm bool) *TestChain {
	chain.UseWasmClient = wasm
	return chain
}

// GetContext returns the current context for the application.
func (chain *TestChain) GetContext() sdk.Context {
	return chain.App.GetBaseApp().NewUncachedContext(false, chain.CurrentHeader)
}

// QueryProof performs an abci query with the given key and returns the proto encoded merkle proof
// for the query and the height at which the proof will succeed on a tendermint verifier.
func (chain *TestChain) QueryProof(key []byte) ([]byte, clienttypes.Height) {
	return chain.QueryProofAtHeight(key, chain.App.LastBlockHeight())
}

// QueryProof performs an abci query with the given key and returns the proto encoded merkle proof
// for the query and the height at which the proof will succeed on a tendermint verifier.
func (chain *TestChain) QueryProofAtHeight(key []byte, height int64) ([]byte, clienttypes.Height) {
	return chain.QueryProofForStore(exported.StoreKey, key, height)
}

// QueryProofForStore performs an abci query with the given key and returns the proto encoded merkle proof
// for the query and the height at which the proof will succeed on a tendermint verifier.
func (chain *TestChain) QueryProofForStore(storeKey string, key []byte, height int64) ([]byte, clienttypes.Height) {
	res, err := chain.App.Query(context.TODO(), &abci.RequestQuery{
		Path:   fmt.Sprintf("store/%s/key", storeKey),
		Height: height - 1,
		Data:   key,
		Prove:  true,
	})
	require.NoError(chain.t, err)

	merkleProof, err := commitmenttypes.ConvertProofs(res.ProofOps)
	require.NoError(chain.t, err)

	proof, err := chain.App.AppCodec().Marshal(&merkleProof)
	require.NoError(chain.t, err)

	revision := clienttypes.ParseChainID(chain.ChainID)

	// proof height + 1 is returned as the proof created corresponds to the height the proof
	// was created in the IAVL tree. Tendermint and subsequently the clients that rely on it
	// have heights 1 above the IAVL tree. Thus we return proof height + 1
	return proof, clienttypes.NewHeight(revision, uint64(res.Height)+1)
}

// QueryUpgradeProof performs an abci query with the given key and returns the proto encoded merkle proof
// for the query and the height at which the proof will succeed on a tendermint verifier.
func (chain *TestChain) QueryUpgradeProof(key []byte, height uint64) ([]byte, clienttypes.Height) {
	res, err := chain.App.Query(context.TODO(), &abci.RequestQuery{
		Path:   "store/upgrade/key",
		Height: int64(height - 1),
		Data:   key,
		Prove:  true,
	})
	require.NoError(chain.t, err)

	merkleProof, err := commitmenttypes.ConvertProofs(res.ProofOps)
	require.NoError(chain.t, err)

	proof, err := chain.App.AppCodec().Marshal(&merkleProof)
	require.NoError(chain.t, err)

	revision := clienttypes.ParseChainID(chain.ChainID)

	// proof height + 1 is returned as the proof created corresponds to the height the proof
	// was created in the IAVL tree. Tendermint and subsequently the clients that rely on it
	// have heights 1 above the IAVL tree. Thus we return proof height + 1
	return proof, clienttypes.NewHeight(revision, uint64(res.Height+1))
}

// QueryConsensusStateProof performs an abci query for a consensus state
// stored on the given clientID. The proof and consensusHeight are returned.
func (chain *TestChain) QueryConsensusStateProof(clientID string) ([]byte, clienttypes.Height) {
	clientState := chain.GetClientState(clientID)

	consensusHeight := clientState.GetLatestHeight().(clienttypes.Height)
	consensusKey := host.FullConsensusStateKey(clientID, consensusHeight)
	proofConsensus, _ := chain.QueryProof(consensusKey)

	return proofConsensus, consensusHeight
}

// NextBlock sets the last header to the current header and increments the current header to be
// at the next block height. It does not update the time as that is handled by the Coordinator.
//
// CONTRACT: this function must only be called after app.Commit() occurs
func (chain *TestChain) NextBlock() {
	res, err := chain.App.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height:             chain.CurrentHeader.Height,
		Time:               chain.CurrentHeader.GetTime(), // todo (Alex): is this the correct time
		NextValidatorsHash: chain.NextVals.Hash(),
	})
	require.NoError(chain.t, err)
	chain.commitBlock(res)
}

func (chain *TestChain) commitBlock(res *abci.ResponseFinalizeBlock) {
	_, err := chain.App.Commit()
	require.NoError(chain.t, err)

	// set the last header to the current header
	// use nil trusted fields
	chain.LastHeader = chain.CurrentCmtClientHeader()

	// val set changes returned from previous block get applied to the next validators
	// of this block. See tendermint spec for details.
	chain.Vals = chain.NextVals
	chain.NextVals = ibctesting.ApplyValSetChanges(chain.t, chain.Vals, res.ValidatorUpdates)

	// increment the current header
	chain.CurrentHeader = tmproto.Header{
		ChainID: chain.ChainID,
		Height:  chain.App.LastBlockHeight() + 1,
		AppHash: chain.App.LastCommitID().Hash,
		// NOTE: the time is increased by the coordinator to maintain time synchrony amongst
		// chains.
		Time:               chain.CurrentHeader.Time,
		ValidatorsHash:     chain.Vals.Hash(),
		NextValidatorsHash: chain.NextVals.Hash(),
		ProposerAddress:    chain.CurrentHeader.ProposerAddress,
	}
}

// CurrentCmtClientHeader creates a CMT header using the current header parameters
// on the chain. The trusted fields in the header are set to nil.
func (chain *TestChain) CurrentCmtClientHeader() *ibctmtypes.Header {
	return chain.CreateCmtClientHeader(
		chain.ChainID,
		chain.CurrentHeader.Height,
		clienttypes.Height{},
		chain.CurrentHeader.Time,
		chain.Vals,
		chain.NextVals,
		nil,
		chain.Signers,
	)
}

// CreateCmtClientHeader creates a CMT header to update the CMT client. Args are passed in to allow
// caller flexibility to use params that differ from the chain.
func (chain *TestChain) CreateCmtClientHeader(chainID string, blockHeight int64, trustedHeight clienttypes.Height, timestamp time.Time, cmtValSet, nextVals, cmtTrustedVals *cmttypes.ValidatorSet, signers map[string]cmttypes.PrivValidator) *ibctmtypes.Header {
	var (
		valSet      *tmproto.ValidatorSet
		trustedVals *tmproto.ValidatorSet
	)
	require.NotNil(chain.t, cmtValSet)

	vsetHash := cmtValSet.Hash()
	nextValHash := nextVals.Hash()

	cmtHeader := cmttypes.Header{
		Version:            cmtprotoversion.Consensus{Block: tmversion.BlockProtocol, App: 2},
		ChainID:            chainID,
		Height:             blockHeight,
		Time:               timestamp,
		LastBlockID:        MakeBlockID(make([]byte, tmhash.Size), 10_000, make([]byte, tmhash.Size)),
		LastCommitHash:     chain.App.LastCommitID().Hash,
		DataHash:           tmhash.Sum([]byte("data_hash")),
		ValidatorsHash:     vsetHash,
		NextValidatorsHash: nextValHash,
		ConsensusHash:      tmhash.Sum([]byte("consensus_hash")),
		AppHash:            chain.CurrentHeader.AppHash,
		LastResultsHash:    tmhash.Sum([]byte("last_results_hash")),
		EvidenceHash:       tmhash.Sum([]byte("evidence_hash")),
		ProposerAddress:    cmtValSet.Proposer.Address, //nolint:staticcheck // SA5011: possible nil pointer dereference
	}

	hhash := cmtHeader.Hash()
	blockID := MakeBlockID(hhash, 3, tmhash.Sum([]byte("part_set")))
	voteSet := cmttypes.NewExtendedVoteSet(chainID, blockHeight, 1, tmproto.PrecommitType, cmtValSet)
	// MakeCommit expects a signer array in the same order as the validator array.
	// Thus we iterate over the ordered validator set and construct a signer array
	// from the signer map in the same order.
	signerArr := make([]cmttypes.PrivValidator, len(cmtValSet.Validators)) //nolint:staticcheck
	for i, v := range cmtValSet.Validators {                               //nolint:staticcheck
		signerArr[i] = signers[v.Address.String()]
	}
	extCommit, err := cmttypes.MakeExtCommit(blockID, blockHeight, 1, voteSet, signerArr, timestamp, true)
	require.NoError(chain.t, err)

	signedHeader := &tmproto.SignedHeader{
		Header: cmtHeader.ToProto(),
		Commit: extCommit.ToCommit().ToProto(),
	}

	if cmtValSet != nil { //nolint:staticcheck
		valSet, err = cmtValSet.ToProto()
		require.NoError(chain.t, err)
	}

	if cmtTrustedVals != nil {
		trustedVals, err = cmtTrustedVals.ToProto()
		require.NoError(chain.t, err)
	}

	// The trusted fields may be nil. They may be filled before relaying messages to a client.
	// The relayer is responsible for querying client and injecting appropriate trusted fields.
	return &ibctmtypes.Header{
		SignedHeader:      signedHeader,
		ValidatorSet:      valSet,
		TrustedHeight:     trustedHeight,
		TrustedValidators: trustedVals,
	}
}

// MakeBlockID copied unimported test functions from cmttypes to use them here
func MakeBlockID(hash []byte, partSetSize uint32, partSetHash []byte) cmttypes.BlockID {
	return cmttypes.BlockID{
		Hash: hash,
		PartSetHeader: cmttypes.PartSetHeader{
			Total: partSetSize,
			Hash:  partSetHash,
		},
	}
}

// sendMsgs delivers a transaction through the application without returning the result.
func (chain *TestChain) sendMsgs(msgs ...sdk.Msg) error {
	_, err := chain.SendMsgs(msgs...)
	return err
}

// SendMsgs delivers a transaction through the application. It updates the senders sequence
// number and updates the TestChain's headers. It returns the result and error if one
// occurred.
func (chain *TestChain) SendMsgs(msgs ...sdk.Msg) (*abci.ExecTxResult, error) {
	rsp, gotErr := chain.sendWithSigner(chain.SenderPrivKey, chain.SenderAccount, msgs...)

	return rsp, gotErr
}

// SendNonDefaultSenderMsgs is the same as SendMsgs but with a custom signer/account
func (chain *TestChain) SendNonDefaultSenderMsgs(senderPrivKey cryptotypes.PrivKey, msgs ...sdk.Msg) (*abci.ExecTxResult, error) {
	require.NotEqual(chain.t, chain.SenderPrivKey, senderPrivKey, "use SendMsgs method")

	addr := sdk.AccAddress(senderPrivKey.PubKey().Address().Bytes())
	account := chain.App.GetAccountKeeper().GetAccount(chain.GetContext(), addr)
	require.NotNil(chain.t, account)
	return chain.sendWithSigner(senderPrivKey, account, msgs...)
}

// sendWithSigner is a generic helper to send messages
func (chain *TestChain) sendWithSigner(
	senderPrivKey cryptotypes.PrivKey,
	senderAccount sdk.AccountI,
	msgs ...sdk.Msg,
) (*abci.ExecTxResult, error) {
	// ensure the chain has the latest time
	chain.Coordinator.UpdateTimeForChain(chain)

	// increment acc sequence regardless of success or failure tx execution
	defer func() {
		err := chain.SenderAccount.SetSequence(chain.SenderAccount.GetSequence() + 1)
		if err != nil {
			panic(err)
		}
	}()

	blockResp, gotErr := app.SignAndDeliver(
		chain.t,
		chain.TxConfig,
		chain.App.GetBaseApp(),
		msgs,
		chain.ChainID,
		[]uint64{senderAccount.GetAccountNumber()},
		[]uint64{senderAccount.GetSequence()},
		true,
		chain.CurrentHeader.GetTime(),
		chain.NextVals.Hash(),
		senderPrivKey,
	)
	if gotErr != nil {
		return nil, gotErr
	}
	chain.commitBlock(blockResp)

	require.Len(chain.t, blockResp.TxResults, 1)
	txResult := blockResp.TxResults[0]

	if txResult.Code != 0 {
		return txResult, fmt.Errorf("%s/%d: %q", txResult.Codespace, txResult.Code, txResult.Log)
	}

	chain.Coordinator.IncrementTime()
	chain.CaptureIBCEvents(txResult)
	return txResult, nil
}

func (chain *TestChain) SendMsgsWithExpPass(expPass bool, msgs ...sdk.Msg) (*abci.ExecTxResult, error) {
	// ensure the chain has the latest time
	chain.Coordinator.UpdateTimeForChain(chain)

	blockResp, err := app.SignAndDeliverWithoutCommit(chain.t, chain.TxConfig, chain.App.GetBaseApp(), msgs, chain.ChainID, []uint64{chain.SenderAccount.GetAccountNumber()}, []uint64{chain.SenderAccount.GetSequence()}, chain.CurrentHeader.GetTime(), chain.SenderPrivKey)
	if err != nil {
		return nil, err
	}

	// SignAndDeliver calls app.Commit()
	chain.NextBlock()

	// increment sequence for successful transaction execution
	err = chain.SenderAccount.SetSequence(chain.SenderAccount.GetSequence() + 1)
	if err != nil {
		return nil, err
	}

	chain.Coordinator.IncrementTime()

	txResult := blockResp.TxResults[0]
	if txResult.Code != 0 {
		return txResult, fmt.Errorf("%s/%d: %q", txResult.Codespace, txResult.Code, txResult.Log)
	}
	chain.CaptureIBCEvents(txResult)

	return txResult, nil
}

func (chain *TestChain) CaptureIBCEvents(r *abci.ExecTxResult) {
	toSend := GetSendPackets(r.Events)
	if len(toSend) > 0 {
		// Keep a queue on the chain that we can relay in tests
		chain.PendingSendPackets = append(chain.PendingSendPackets, toSend...)
	}
	toAck := getAckPackets(r.Events)
	if len(toAck) > 0 {
		// Keep a queue on the chain that we can relay in tests
		chain.PendingAckPackets = append(chain.PendingAckPackets, toAck...)
	}
}

// GetClientState retrieves the client state for the provided clientID. The client is
// expected to exist otherwise testing will fail.
func (chain *TestChain) GetClientState(clientID string) exported.ClientState {
	clientState, found := chain.App.GetIBCKeeper().ClientKeeper.GetClientState(chain.GetContext(), clientID)
	require.True(chain.t, found)

	return clientState
}

// GetConsensusState retrieves the consensus state for the provided clientID and height.
// It will return a success boolean depending on if consensus state exists or not.
func (chain *TestChain) GetConsensusState(clientID string, height exported.Height) (exported.ConsensusState, bool) {
	return chain.App.GetIBCKeeper().ClientKeeper.GetClientConsensusState(chain.GetContext(), clientID, height)
}

// GetValsAtHeight will return the validator set of the chain at a given height. It will return
// a success boolean depending on if the validator set exists or not at that height.
func (chain *TestChain) GetValsAtHeight(height int64) (*cmttypes.ValidatorSet, bool) {
	// if the current uncommitted header equals the requested height, then we can return
	// the current validator set as this validator set will be stored in the historical info
	// when the block height is executed
	if height == chain.CurrentHeader.Height {
		return chain.Vals, true
	}

	histInfo, err := chain.App.GetStakingKeeper().GetHistoricalInfo(chain.GetContext(), height)
	if err != nil {
		return nil, false
	}

	valSet := stakingtypes.Validators{
		Validators: histInfo.Valset,
	}

	cmtValidators, err := testutil.ToCmtValidators(valSet, sdk.DefaultPowerReduction)
	if err != nil {
		panic(err)
	}
	return cmttypes.NewValidatorSet(cmtValidators), true
}

// GetAcknowledgement retrieves an acknowledgement for the provided packet. If the
// acknowledgement does not exist then testing will fail.
func (chain *TestChain) GetAcknowledgement(packet exported.PacketI) []byte {
	ack, found := chain.App.GetIBCKeeper().ChannelKeeper.GetPacketAcknowledgement(chain.GetContext(), packet.GetDestPort(), packet.GetDestChannel(), packet.GetSequence())
	require.True(chain.t, found)

	return ack
}

// GetPrefix returns the prefix for used by a chain in connection creation
func (chain *TestChain) GetPrefix() commitmenttypes.MerklePrefix {
	return commitmenttypes.NewMerklePrefix(chain.App.GetIBCKeeper().ConnectionKeeper.GetCommitmentPrefix().Bytes())
}

// ConstructUpdateTMClientHeader will construct a valid 07-tendermint Header to update the
// light client on the source chain.
func (chain *TestChain) ConstructUpdateTMClientHeader(counterparty *TestChain, clientID string) (*ibctmtypes.Header, error) {
	return chain.ConstructUpdateTMClientHeaderWithTrustedHeight(counterparty, clientID, clienttypes.ZeroHeight())
}

// ConstructUpdateTMClientHeader will construct a valid 07-tendermint Header to update the
// light client on the source chain.
func (chain *TestChain) ConstructUpdateTMClientHeaderWithTrustedHeight(counterparty *TestChain, clientID string, trustedHeight clienttypes.Height) (*ibctmtypes.Header, error) {
	header := counterparty.LastHeader
	// Relayer must query for LatestHeight on client to get TrustedHeight if the trusted height is not set
	if trustedHeight.IsZero() {
		trustedHeight = chain.GetClientState(clientID).GetLatestHeight().(clienttypes.Height)
	}
	var (
		tmTrustedVals *cmttypes.ValidatorSet
		ok            bool
	)
	// Once we get TrustedHeight from client, we must query the validators from the counterparty chain
	// If the LatestHeight == LastHeader.Height, then TrustedValidators are current validators
	// If LatestHeight < LastHeader.Height, we can query the historical validator set from HistoricalInfo
	if trustedHeight == counterparty.LastHeader.GetHeight() {
		tmTrustedVals = counterparty.Vals
	} else {
		// NOTE: We need to get validators from counterparty at height: trustedHeight+1
		// since the last trusted validators for a header at height h
		// is the NextValidators at h+1 committed to in header h by
		// NextValidatorsHash
		tmTrustedVals, ok = counterparty.GetValsAtHeight(int64(trustedHeight.RevisionHeight + 1))
		if !ok {
			return nil, errors.Wrapf(ibctmtypes.ErrInvalidHeaderHeight, "could not retrieve trusted validators at trustedHeight: %d", trustedHeight)
		}
	}
	// inject trusted fields into last header
	// for now assume revision number is 0
	header.TrustedHeight = trustedHeight

	trustedVals, err := tmTrustedVals.ToProto()
	if err != nil {
		return nil, err
	}
	header.TrustedValidators = trustedVals

	return header, nil
}

// ExpireClient fast forwards the chain's block time by the provided amount of time which will
// expire any clients with a trusting period less than or equal to this amount of time.
func (chain *TestChain) ExpireClient(amount time.Duration) {
	chain.Coordinator.IncrementTimeBy(amount)
}

// CreatePortCapability binds and claims a capability for the given portID if it does not
// already exist. This function will fail testing on any resulting error.
// NOTE: only creation of a capbility for a transfer or mock port is supported
// Other applications must bind to the port in InitGenesis or modify this code.
func (chain *TestChain) CreatePortCapability(scopedKeeper capabilitykeeper.ScopedKeeper, portID string) {
	// check if the portId is already binded, if not bind it
	_, ok := chain.App.GetScopedIBCKeeper().GetCapability(chain.GetContext(), host.PortPath(portID))
	if !ok {
		// create capability using the IBC capability keeper
		capability, err := chain.App.GetScopedIBCKeeper().NewCapability(chain.GetContext(), host.PortPath(portID))
		require.NoError(chain.t, err)

		// claim capability using the scopedKeeper
		err = scopedKeeper.ClaimCapability(chain.GetContext(), capability, host.PortPath(portID))
		require.NoError(chain.t, err)
	}

	chain.App.Commit()

	chain.NextBlock()
}

// GetPortCapability returns the port capability for the given portID. The capability must
// exist, otherwise testing will fail.
func (chain *TestChain) GetPortCapability(portID string) *capabilitytypes.Capability {
	capability, ok := chain.App.GetScopedIBCKeeper().GetCapability(chain.GetContext(), host.PortPath(portID))
	require.True(chain.t, ok)

	return capability
}

// CreateChannelCapability binds and claims a capability for the given portID and channelID
// if it does not already exist. This function will fail testing on any resulting error. The
// scoped keeper passed in will claim the new capability.
func (chain *TestChain) CreateChannelCapability(scopedKeeper capabilitykeeper.ScopedKeeper, portID, channelID string) {
	capName := host.ChannelCapabilityPath(portID, channelID)
	// check if the portId is already binded, if not bind it
	_, ok := chain.App.GetScopedIBCKeeper().GetCapability(chain.GetContext(), capName)
	if !ok {
		capability, err := chain.App.GetScopedIBCKeeper().NewCapability(chain.GetContext(), capName)
		require.NoError(chain.t, err)
		err = scopedKeeper.ClaimCapability(chain.GetContext(), capability, capName)
		require.NoError(chain.t, err)
	}

	chain.App.Commit()

	chain.NextBlock()
}

// GetChannelCapability returns the channel capability for the given portID and channelID.
// The capability must exist, otherwise testing will fail.
func (chain *TestChain) GetChannelCapability(portID, channelID string) *capabilitytypes.Capability {
	capability, ok := chain.App.GetScopedIBCKeeper().GetCapability(chain.GetContext(), host.ChannelCapabilityPath(portID, channelID))
	require.True(chain.t, ok)

	return capability
}

func (chain *TestChain) TransferMiddleware() transfermiddlewarekeeper.Keeper {
	return chain.App.GetTransferMiddlewareKeeper()
}

func (chain *TestChain) RateLimit() ratelimitmodulekeeper.Keeper {
	return chain.App.GetRateLimitKeeper()
}

func (chain *TestChain) TransferKeeper() customibctransferkeeper.Keeper {
	return chain.App.GetTransferKeeper()
}

func (chain *TestChain) Balance(acc sdk.AccAddress, denom string) sdk.Coin {
	return chain.App.GetBankKeeper().GetBalance(chain.GetContext(), acc, denom)
}

func (chain *TestChain) AllBalances(acc sdk.AccAddress) sdk.Coins {
	return chain.App.GetBankKeeper().GetAllBalances(chain.GetContext(), acc)
}

func (chain *TestChain) GetBankKeeper() bankkeeper.Keeper {
	return chain.App.GetBankKeeper()
}

func (chain TestChain) Wasm08Keeper() wasm08.Keeper {
	return chain.App.GetWasm08Keeper()
}

func (chain *TestChain) QueryContract(suite *suite.Suite, contract sdk.AccAddress, key []byte) string {
	wasmKeeper := chain.App.GetWasmKeeper()
	state, err := wasmKeeper.QuerySmart(chain.GetContext(), contract, key)
	suite.Require().NoError(err)
	return string(state)
}

//
// func (chain *TestChain) StoreContractCode(suite *suite.Suite, path string) {
//	govModuleAddress := chain.App.GetAccountKeeper().GetModuleAddress(govtypes.ModuleName)
//	wasmCode, err := os.ReadFile(path)
//	suite.Require().NoError(err)
//
//	src := wasmtypes.New(func(p *wasmtypes.StoreCodeProposal) { //nolint: staticcheck
//		p.RunAs = govModuleAddress.String()
//		p.WASMByteCode = wasmCode
//		checksum := sha256.Sum256(wasmCode)
//		p.CodeHash = checksum[:]
//	})
//
//	govKeeper := chain.App.GetGovKeeper()
//	// when
//	mustSubmitAndExecuteLegacyProposal(suite.T(), chain.GetContext(), src, chain.SenderAccount.GetAddress().String(), govKeeper, govModuleAddress.String())
//	suite.Require().NoError(err)
//}

func (chain *TestChain) InstantiateContract(suite *suite.Suite, msg string, codeID uint64) sdk.AccAddress {
	wasmKeeper := chain.App.GetWasmKeeper()
	govModuleAddress := chain.App.GetAccountKeeper().GetModuleAddress(govtypes.ModuleName)

	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(wasmKeeper)
	addr, _, err := contractKeeper.Instantiate(chain.GetContext(), codeID, govModuleAddress, govModuleAddress, []byte(msg), "contract", nil)
	suite.Require().NoError(err)
	return addr
}
