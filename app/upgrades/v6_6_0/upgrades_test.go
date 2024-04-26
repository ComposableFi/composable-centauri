package v6_6_0_test

import (
	"encoding/json"
	ibchookskeeper "github.com/notional-labs/composable/v6/x/ibc-hooks/keeper"
	ibctransfermiddlewaretypes "github.com/notional-labs/composable/v6/x/ibctransfermiddleware/types"
	"strings"
	"testing"
	"time"

	"github.com/notional-labs/composable/v6/app/upgrades/v6_6_0"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	routertypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	ibchookstypes "github.com/notional-labs/composable/v6/x/ibc-hooks/types"
	ibctransfemiddlewaretypes "github.com/notional-labs/composable/v6/x/ibctransfermiddleware/types"

	apptesting "github.com/notional-labs/composable/v6/app"
	"github.com/notional-labs/composable/v6/bech32-migration/utils"
	"github.com/stretchr/testify/suite"
	alliancetypes "github.com/terra-money/alliance/x/alliance/types"
)

const (
	COIN_DENOM   = "stake"
	CONNECTION_0 = "connection-0"
	PORT_0       = "port-0"
	CHANNEL_0    = "channel-0"
)

type UpgradeTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

// Ensures the test does not error out.
func (s *UpgradeTestSuite) TestForMigratingNewPrefix() {
	// DEFAULT PREFIX: centauri
	sdk.SetAddrCacheEnabled(false)

	sdk.GetConfig().SetBech32PrefixForAccount(utils.OldBech32PrefixAccAddr, utils.OldBech32PrefixAccPub)
	sdk.GetConfig().SetBech32PrefixForValidator(utils.OldBech32PrefixValAddr, utils.OldBech32PrefixValPub)
	sdk.GetConfig().SetBech32PrefixForConsensusNode(utils.OldBech32PrefixConsAddr, utils.OldBech32PrefixConsPub)

	s.Setup(s.T())

	acc1, proposal := prepareForTestingGovModule(s)
	oldConsAddress := prepareForTestingSlashingModule(s)
	oldValAddress, oldValAddress2, acc3, afterOneDay := prepareForTestingStakingModule(s)
	baseAccount, stakingModuleAccount, baseVestingAccount, continuousVestingAccount, delayedVestingAccount, periodicVestingAccount, permanentLockedAccount := prepareForTestingAuthModule(s)
	prepareForTestingAllianceModule(s)
	prepareForTestingICAHostModule(s)
	prepareForTestingMintModule(s)
	prepareForTestingTransferMiddlewareModule(s)
	prepareForTestingPfmMiddlewareModule(s)
	prepareForTestingIbcTransferMiddlewareModule(s)
	prepareForTestingIbcHooksModule(s)

	/* == UPGRADE == */
	upgradeHeight := int64(5)
	s.ConfirmUpgradeSucceeded(v6_6_0.UpgradeName, upgradeHeight)

	/* == CHECK AFTER UPGRADE == */
	checkUpgradeGovModule(s, acc1, proposal)
	checkUpgradeSlashingModule(s, oldConsAddress)
	checkUpgradeStakingModule(s, oldValAddress, oldValAddress2, acc3, afterOneDay)
	checkUpgradeAuthModule(s, baseAccount, stakingModuleAccount, baseVestingAccount, continuousVestingAccount, delayedVestingAccount, periodicVestingAccount, permanentLockedAccount)
	checkUpgradeAllianceModule(s)
	checkUpgradeICAHostModule(s)
	checkUpgradeMintModule(s)
	checkUpgradeTransferMiddlewareModule(s)
	checkUpgradePfmMiddlewareModule(s)
	checkUpgradeIbcTransferMiddlewareModule(s)
	checkUpgradeIbcHooksMiddlewareModule(s)
}

func prepareForTestingGovModule(s *UpgradeTestSuite) (sdk.AccAddress, govtypes.Proposal) {
	/* PREPARE FOR TESTING GOV MODULE */
	acc1 := s.TestAccs[0]

	// MINT NEW TOKEN FOR BALANCE CHECKING
	s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(COIN_DENOM, math.NewIntFromUint64(100000000))))

	// VOTE AND DEPOSIT
	proposal, err := s.App.GovKeeper.SubmitProposal(s.Ctx, []sdk.Msg{}, "", "test", "description", acc1)
	s.Suite.Equal(err, nil)

	s.App.GovKeeper.SetVote(s.Ctx, govtypes.Vote{
		ProposalId: proposal.Id,
		Voter:      acc1.String(),
		Options:    nil,
		Metadata:   "",
	})

	s.App.GovKeeper.SetDeposit(s.Ctx, govtypes.Deposit{
		ProposalId: proposal.Id,
		Depositor:  acc1.String(),
		Amount:     sdk.NewCoins(sdk.NewCoin(COIN_DENOM, math.NewIntFromUint64(1))),
	})

	return acc1, proposal
}

func prepareForTestingSlashingModule(s *UpgradeTestSuite) sdk.ConsAddress {
	/* PREPARE FOR TESTING SLASHING MODULE */
	acc2 := s.TestAccs[1]

	oldConsAddress, err := utils.ConsAddressFromOldBech32(acc2.String(), utils.OldBech32PrefixAccAddr)
	s.Suite.Equal(err, nil)

	// CHECK ValidatorSigningInfo
	s.App.SlashingKeeper.SetValidatorSigningInfo(s.Ctx, oldConsAddress, slashingtypes.ValidatorSigningInfo{
		Address: oldConsAddress.String(),
	})
	return oldConsAddress
}

func prepareForTestingStakingModule(s *UpgradeTestSuite) (sdk.ValAddress, sdk.ValAddress, sdk.AccAddress, time.Time) {
	/* PREPARE FOR TESTING SLASHING MODULE */
	acc3 := s.TestAccs[2]

	// MINT NEW TOKEN FOR BALANCE CHECKING
	s.App.BankKeeper.MintCoins(s.Ctx, minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin(COIN_DENOM, math.NewIntFromUint64(100000000))))
	s.App.BankKeeper.SendCoinsFromModuleToAccount(s.Ctx, minttypes.ModuleName, acc3, sdk.NewCoins(sdk.NewCoin(COIN_DENOM, math.NewIntFromUint64(100000000))))

	// validator.OperatorAddress
	oldValAddress := s.SetupValidator(stakingtypes.Bonded)
	oldValAddress2 := s.SetupValidator(stakingtypes.Bonded)

	// delegation.DelegatorAddress & delegation.ValidatorAddress
	s.StakingHelper.Delegate(acc3, oldValAddress, sdk.NewInt(300))

	// redelegation.DelegatorAddress & redelegation.ValidatorSrcAddress & redelegation.ValidatorDstAddress
	completionTime, err := s.App.StakingKeeper.BeginRedelegation(s.Ctx, acc3, oldValAddress, oldValAddress2, sdk.NewDec(100))
	afterOneDay := completionTime.AddDate(0, 0, 1)
	s.Require().NoError(err)

	// Undelegate part of the tokens from val2 (test instant unbonding on undelegation started before upgrade)
	s.StakingHelper.Undelegate(acc3, oldValAddress, sdk.NewInt(10), true)

	s.App.StakingKeeper.SetRedelegationQueueTimeSlice(s.Ctx, time.Date(2024, time.March, 4, 12, 0, 0, 0, time.UTC), []stakingtypes.DVVTriplet{
		{
			DelegatorAddress:    s.TestAccs[2].String(),
			ValidatorDstAddress: oldValAddress.String(),
			ValidatorSrcAddress: oldValAddress2.String(),
		},
	})

	return oldValAddress, oldValAddress2, acc3, afterOneDay
}

func prepareForTestingAuthModule(s *UpgradeTestSuite) (sdk.AccAddress, sdk.AccAddress, sdk.AccAddress, sdk.AccAddress, sdk.AccAddress, sdk.AccAddress, sdk.AccAddress) {
	addr0 := s.TestAccs[0]
	baseAccount := authtypes.NewBaseAccount(addr0, nil, 0, 0)
	s.App.AccountKeeper.SetAccount(s.Ctx, baseAccount)

	addr6 := s.TestAccs[6]
	baseAccount6 := authtypes.NewBaseAccount(addr6, nil, 0, 0)
	stakingModuleAccount := authtypes.NewModuleAccount(baseAccount6, "name", "name")
	s.App.AccountKeeper.SetAccount(s.Ctx, stakingModuleAccount)

	addr2 := s.TestAccs[2]
	baseAccount2 := authtypes.NewBaseAccount(addr2, nil, 0, 0)
	baseVestingAccount := vestingtypes.NewBaseVestingAccount(baseAccount2, sdk.NewCoins(sdk.NewCoin(COIN_DENOM, math.NewIntFromUint64(1))), 60)
	s.App.AccountKeeper.SetAccount(s.Ctx, baseVestingAccount)

	continuousVestingAccount := CreateVestingAccount(s)

	addr3 := s.TestAccs[3]
	baseAccount3 := authtypes.NewBaseAccount(addr3, nil, 0, 0)
	baseVestingAccount2 := vestingtypes.NewBaseVestingAccount(baseAccount3, sdk.NewCoins(sdk.NewCoin(COIN_DENOM, math.NewIntFromUint64(1))), 60)
	delayedVestingAccount := vestingtypes.NewDelayedVestingAccountRaw(baseVestingAccount2)
	s.App.AccountKeeper.SetAccount(s.Ctx, delayedVestingAccount)

	addr4 := s.TestAccs[4]
	baseAccount4 := authtypes.NewBaseAccount(addr4, nil, 0, 0)
	baseVestingAccount3 := vestingtypes.NewBaseVestingAccount(baseAccount4, sdk.NewCoins(sdk.NewCoin(COIN_DENOM, math.NewIntFromUint64(1))), 60)
	periodicVestingAccount := vestingtypes.NewPeriodicVestingAccountRaw(baseVestingAccount3, 0, vestingtypes.Periods{})
	s.App.AccountKeeper.SetAccount(s.Ctx, periodicVestingAccount)

	addr5 := s.TestAccs[5]
	baseAccount5 := authtypes.NewBaseAccount(addr5, nil, 0, 0)
	permanentLockedAccount := vestingtypes.NewPermanentLockedAccount(baseAccount5, sdk.NewCoins(sdk.NewCoin(COIN_DENOM, math.NewIntFromUint64(1))))
	s.App.AccountKeeper.SetAccount(s.Ctx, permanentLockedAccount)

	return baseAccount.GetAddress(), stakingModuleAccount.GetAddress(), baseVestingAccount.GetAddress(), continuousVestingAccount.GetAddress(), delayedVestingAccount.GetAddress(), periodicVestingAccount.GetAddress(), permanentLockedAccount.GetAddress()
}

func prepareForTestingICAHostModule(s *UpgradeTestSuite) {
	acc1 := s.TestAccs[0]
	s.App.ICAHostKeeper.SetInterchainAccountAddress(s.Ctx, CONNECTION_0, PORT_0, acc1.String())
}

func prepareForTestingMintModule(s *UpgradeTestSuite) {
	acc1 := s.TestAccs[0]
	s.App.MintKeeper.SetAllowedAddress(s.Ctx, acc1.String())
}

func prepareForTestingAllianceModule(s *UpgradeTestSuite) {
	oldValAddress := s.SetupValidator(stakingtypes.Bonded)
	_, bz, _ := bech32.DecodeAndConvert(oldValAddress.String())
	oldBech32Addr, _ := bech32.ConvertAndEncode(utils.OldBech32PrefixValAddr, bz)

	s.App.AllianceKeeper.InitGenesis(s.Ctx, &alliancetypes.GenesisState{
		ValidatorInfos: []alliancetypes.ValidatorInfoState{{
			ValidatorAddress: oldBech32Addr,
			Validator:        alliancetypes.NewAllianceValidatorInfo(),
		}},
	})
}

func prepareForTestingTransferMiddlewareModule(s *UpgradeTestSuite) {
	acc1 := s.TestAccs[0]
	s.App.TransferMiddlewareKeeper.SetAllowRlyAddress(s.Ctx, acc1.String())
}

func prepareForTestingPfmMiddlewareModule(s *UpgradeTestSuite) {
	store := s.Ctx.KVStore(s.App.GetKey(routertypes.StoreKey))
	inFlightPacket := routertypes.InFlightPacket{
		PacketData:            []byte("{\"amount\":\"10000\",\"denom\":\"transfer/channel-6660/ppica\",\"memo\":\"{\\\"forward\\\":{\\\"receiver\\\":\\\"osmo1wkjvpgkuchq0r8425g4z4sf6n85zj5wth3u77y\\\",\\\"port\\\":\\\"transfer\\\",\\\"channel\\\":\\\"channel-9\\\",\\\"timeout\\\":600000000000,\\\"retries\\\":0}}\",\"receiver\":\"centauri1wkjvpgkuchq0r8425g4z4sf6n85zj5wtmqzjv9\",\"sender\":\"osmo1wkjvpgkuchq0r8425g4z4sf6n85zj5wth3u77y\"}"),
		OriginalSenderAddress: "centauri1wkjvpgkuchq0r8425g4z4sf6n85zj5wtmqzjv9",
		RefundChannelId:       "channel-9",
		RefundPortId:          "transfer",
		RefundSequence:        18,
		PacketSrcPortId:       "transfer",
		PacketSrcChannelId:    "channel-66660",

		PacketTimeoutTimestamp: 1712153063084849609,
		PacketTimeoutHeight:    "5-123",

		RetriesRemaining: int32(0),
		Timeout:          uint64(600000000000),
		Nonrefundable:    false,
	}

	encCdc := apptesting.MakeEncodingConfig()

	key := routertypes.RefundPacketKey("channel-9", "transfer", 0)
	bz := encCdc.Amino.MustMarshal(&inFlightPacket)
	store.Set(key, bz)

	key = routertypes.RefundPacketKey("channel-9", "transfer", 2)
	inFlightPacket.OriginalSenderAddress = "centauri1hj5fveer5cjtn4wd6wstzugjfdxzl0xpzxlwgs"
	bz = encCdc.Amino.MustMarshal(&inFlightPacket)
	store.Set(key, bz)
}

func prepareForTestingIbcTransferMiddlewareModule(s *UpgradeTestSuite) {
	store := s.Ctx.KVStore(s.App.GetKey(ibctransfermiddlewaretypes.StoreKey))
	var fees []*ibctransfemiddlewaretypes.ChannelFee
	fees = append(fees, &ibctransfemiddlewaretypes.ChannelFee{
		Channel: "channel-7",
		AllowedTokens: []*ibctransfemiddlewaretypes.CoinItem{{
			MinFee:     sdk.Coin{},
			Percentage: 20,
		}},
		FeeAddress:          "centauri1hj5fveer5cjtn4wd6wstzugjfdxzl0xpzxlwgs",
		MinTimeoutTimestamp: 0,
	})
	fees = append(fees, &ibctransfemiddlewaretypes.ChannelFee{
		Channel: "channel-9",
		AllowedTokens: []*ibctransfemiddlewaretypes.CoinItem{{
			MinFee:     sdk.Coin{},
			Percentage: 10,
		}},
		FeeAddress:          "centauri1hj5fveer5cjtn4wd6wstzugjfdxzl0xpzxlwgs",
		MinTimeoutTimestamp: 0,
	})
	params := ibctransfemiddlewaretypes.Params{ChannelFees: fees}
	encCdc := apptesting.MakeEncodingConfig()
	bz := encCdc.Amino.MustMarshal(&params)
	store.Set(ibctransfemiddlewaretypes.ParamsKey, bz)
}

func prepareForTestingIbcHooksModule(s *UpgradeTestSuite) {
	store := s.Ctx.KVStore(s.App.GetKey(ibchookstypes.StoreKey))
	store.Set(ibchookskeeper.GetPacketKey("channel-2", 2), []byte("centauri1hj5fveer5cjtn4wd6wstzugjfdxzl0xpzxlwgs"))
	store.Set(ibchookskeeper.GetPacketKey("channel-4", 2), []byte("centauri1wkjvpgkuchq0r8425g4z4sf6n85zj5wtmqzjv9"))
}

func checkUpgradeGovModule(s *UpgradeTestSuite, acc1 sdk.AccAddress, proposal govtypes.Proposal) {
	// CONVERT ACC TO NEW PREFIX
	_, bz, _ := bech32.DecodeAndConvert(acc1.String())
	newBech32Addr, _ := bech32.ConvertAndEncode(utils.NewBech32PrefixAccAddr, bz)
	newAddr, err := utils.AccAddressFromOldBech32(newBech32Addr, utils.NewBech32PrefixAccAddr)
	s.Suite.Equal(err, nil)

	// CHECK PROPOSAL
	proposal, found := s.App.GovKeeper.GetProposal(s.Ctx, proposal.Id)
	s.Suite.Equal(found, true)
	s.Suite.Equal(proposal.Proposer, newBech32Addr)

	// CHECK VOTER AND DEPOSITER OF NEW ADDRESS
	existed_proposal, _ := s.App.GovKeeper.GetProposal(s.Ctx, proposal.Id)
	s.Suite.Equal(existed_proposal.Proposer, newBech32Addr)

	vote, found := s.App.GovKeeper.GetVote(s.Ctx, proposal.Id, newAddr)
	s.Suite.Equal(found, true)
	s.Suite.Equal(vote.Voter, newBech32Addr)

	deposit, found := s.App.GovKeeper.GetDeposit(s.Ctx, proposal.Id, newAddr)
	s.Suite.Equal(found, true)
	s.Suite.Equal(deposit.Depositor, newBech32Addr)
}

func checkUpgradeSlashingModule(s *UpgradeTestSuite, oldConsAddress sdk.ConsAddress) {
	// CONVERT TO ACC TO NEW PREFIX
	_, bz, _ := bech32.DecodeAndConvert(oldConsAddress.String())
	newBech32Addr, _ := bech32.ConvertAndEncode(utils.NewBech32PrefixConsAddr, bz)
	newAddr, err := utils.ConsAddressFromOldBech32(newBech32Addr, utils.NewBech32PrefixConsAddr)
	s.Suite.Equal(err, nil)

	valSigningInfo, found := s.App.SlashingKeeper.GetValidatorSigningInfo(s.Ctx, newAddr)
	s.Suite.Equal(found, true)
	s.Suite.Equal(valSigningInfo.Address, newBech32Addr)
}

func checkUpgradeStakingModule(s *UpgradeTestSuite, oldValAddress, oldValAddress2 sdk.ValAddress, acc1 sdk.AccAddress, afterOneDay time.Time) {
	// CONVERT TO ACC TO NEW PREFIX
	_, bz, _ := bech32.DecodeAndConvert(oldValAddress.String())
	newBech32Addr, _ := bech32.ConvertAndEncode(utils.NewBech32PrefixValAddr, bz)
	newValAddr, err := utils.ValAddressFromOldBech32(newBech32Addr, utils.NewBech32PrefixValAddr)
	s.Suite.Equal(err, nil)

	_, bzVal2, _ := bech32.DecodeAndConvert(oldValAddress2.String())
	newBech32AddrVal2, _ := bech32.ConvertAndEncode(utils.NewBech32PrefixValAddr, bzVal2)
	newValAddr2, err := utils.ValAddressFromOldBech32(newBech32AddrVal2, utils.NewBech32PrefixValAddr)
	s.Suite.Equal(err, nil)

	_, bz1, _ := bech32.DecodeAndConvert(acc1.String())
	newBech32DelAddr, _ := bech32.ConvertAndEncode(utils.NewBech32PrefixAccAddr, bz1)
	newAccAddr, err := utils.AccAddressFromOldBech32(newBech32DelAddr, utils.NewBech32PrefixAccAddr)
	s.Suite.Equal(err, nil)

	val, found := s.App.StakingKeeper.GetValidator(s.Ctx, newValAddr)
	s.Suite.Equal(found, true)
	s.Suite.Equal(val.OperatorAddress, newBech32Addr)

	delegation, found := s.App.StakingKeeper.GetDelegation(s.Ctx, newAccAddr, newValAddr)
	s.Suite.Equal(found, true)
	s.Suite.Equal(delegation.DelegatorAddress, newBech32DelAddr)
	s.Suite.Equal(delegation.ValidatorAddress, newBech32Addr)

	unbonding, found := s.App.StakingKeeper.GetUnbondingDelegation(s.Ctx, newAccAddr, newValAddr)
	s.Suite.Equal(found, true)
	s.Suite.Equal(unbonding.DelegatorAddress, newBech32DelAddr)
	s.Suite.Equal(unbonding.ValidatorAddress, newBech32Addr)

	s.Ctx = s.Ctx.WithBlockTime(afterOneDay)

	redelegation, found := s.App.StakingKeeper.GetRedelegation(s.Ctx, newAccAddr, newValAddr, newValAddr2)
	s.Suite.Equal(found, true)
	s.Suite.Equal(redelegation.DelegatorAddress, newBech32DelAddr)
	s.Suite.Equal(redelegation.ValidatorSrcAddress, newBech32Addr)
	s.Suite.Equal(redelegation.ValidatorDstAddress, newBech32AddrVal2)

	RedelegationQueueTimeSlice := s.App.StakingKeeper.GetRedelegationQueueTimeSlice(s.Ctx, time.Date(2024, time.March, 4, 12, 0, 0, 0, time.UTC))
	s.Suite.Equal(strings.Contains(RedelegationQueueTimeSlice[0].DelegatorAddress, "pica"), true)
	s.Suite.Equal(strings.Contains(RedelegationQueueTimeSlice[0].ValidatorDstAddress, "pica"), true)
	s.Suite.Equal(strings.Contains(RedelegationQueueTimeSlice[0].ValidatorSrcAddress, "pica"), true)
}

func checkUpgradeAuthModule(s *UpgradeTestSuite, baseAccount, stakingModuleAccount, baseVestingAccount, continuousVestingAccount, delayedVestingAccount, periodicVestingAccount, permanentLockedAccount sdk.AccAddress) {
	/* CHECK BASE ACCOUNT */
	_, bz, _ := bech32.DecodeAndConvert(baseAccount.String())
	newBech32AddrBaseAccount, _ := bech32.ConvertAndEncode(utils.NewBech32PrefixAccAddr, bz)
	var newPrefixAddr authtypes.AccountI
	newPrefixAddr = s.App.AccountKeeper.GetAccount(s.Ctx, baseAccount)
	switch acci := newPrefixAddr.(type) {
	case *authtypes.BaseAccount:
		acc := acci
		s.Suite.Equal(acc.Address, newBech32AddrBaseAccount)
	default:
		s.Suite.NotNil(nil)
	}

	/* CHECK MODULE ACCOUNT */
	_, bz, _ = bech32.DecodeAndConvert(stakingModuleAccount.String())
	newBech32AddrModuleAccount, _ := bech32.ConvertAndEncode(utils.NewBech32PrefixAccAddr, bz)
	newPrefixAddr = s.App.AccountKeeper.GetAccount(s.Ctx, stakingModuleAccount)
	switch acci := newPrefixAddr.(type) {
	case *authtypes.ModuleAccount:
		acc := acci
		s.Suite.Equal(acc.Address, newBech32AddrModuleAccount)
	default:
		s.Suite.NotNil(nil)
	}

	/* CHECK BASE VESTING ACCOUNT */
	_, bz, _ = bech32.DecodeAndConvert(baseVestingAccount.String())
	newBech32AddrBaseVestingAccount, _ := bech32.ConvertAndEncode(utils.NewBech32PrefixAccAddr, bz)
	newPrefixAddr = s.App.AccountKeeper.GetAccount(s.Ctx, baseVestingAccount)
	switch acci := newPrefixAddr.(type) {
	case *vestingtypes.BaseVestingAccount:
		acc := acci
		s.Suite.Equal(acc.Address, newBech32AddrBaseVestingAccount)
	default:
		s.Suite.NotNil(nil)
	}

	// CHECK CONTINUOUS VESTING ACCOUNT AND MULTISIG
	_, bz, _ = bech32.DecodeAndConvert(continuousVestingAccount.String())
	newBech32AddrConVestingAccount, _ := bech32.ConvertAndEncode(utils.NewBech32PrefixAccAddr, bz)
	newPrefixAddr = s.App.AccountKeeper.GetAccount(s.Ctx, continuousVestingAccount)
	switch acci := newPrefixAddr.(type) {
	case *vestingtypes.ContinuousVestingAccount:
		acc := acci
		s.Suite.Equal(acc.Address, newBech32AddrConVestingAccount)
	default:
		s.Suite.NotNil(nil)
	}

	// CHECK DELAYED VESTING ACCOUNT
	_, bz, _ = bech32.DecodeAndConvert(delayedVestingAccount.String())
	newBech32AddrDelayedVestingAccount, _ := bech32.ConvertAndEncode(utils.NewBech32PrefixAccAddr, bz)
	newPrefixAddr = s.App.AccountKeeper.GetAccount(s.Ctx, delayedVestingAccount)
	switch acci := newPrefixAddr.(type) {
	case *vestingtypes.DelayedVestingAccount:
		acc := acci
		s.Suite.Equal(acc.Address, newBech32AddrDelayedVestingAccount)
	default:
		s.Suite.NotNil(nil)
	}

	// CHECK PERIODIC VESTING ACCOUNT
	_, bz, _ = bech32.DecodeAndConvert(periodicVestingAccount.String())
	newBech32AddrPeriodicVestingAccount, _ := bech32.ConvertAndEncode(utils.NewBech32PrefixAccAddr, bz)
	newPrefixAddr = s.App.AccountKeeper.GetAccount(s.Ctx, periodicVestingAccount)
	switch acci := newPrefixAddr.(type) {
	case *vestingtypes.PeriodicVestingAccount:
		acc := acci
		s.Suite.Equal(acc.Address, newBech32AddrPeriodicVestingAccount)
	default:
		s.Suite.NotNil(nil)
	}

	// CHECK PERMANENT LOCKED ACCOUNT
	_, bz, _ = bech32.DecodeAndConvert(permanentLockedAccount.String())
	newBech32AddrPermanentVestingAccount, _ := bech32.ConvertAndEncode(utils.NewBech32PrefixAccAddr, bz)
	newPrefixAddr = s.App.AccountKeeper.GetAccount(s.Ctx, permanentLockedAccount)
	switch acci := newPrefixAddr.(type) {
	case *vestingtypes.PermanentLockedAccount:
		acc := acci
		s.Suite.Equal(acc.Address, newBech32AddrPermanentVestingAccount)
	default:
		s.Suite.NotNil(nil)
	}
}

func checkUpgradeAllianceModule(s *UpgradeTestSuite) {
	// the validator address in alliance genesis file is converted into accAddr type
	// and then used for key storage
	// so the migration do not affect this module
	genesis := s.App.AllianceKeeper.ExportGenesis(s.Ctx)
	s.Suite.Equal(strings.Contains(genesis.ValidatorInfos[0].ValidatorAddress, "pica"), true)
}

func checkUpgradeICAHostModule(s *UpgradeTestSuite) {
	acc1 := s.TestAccs[0]
	interchainAccount, _ := s.App.ICAHostKeeper.GetInterchainAccountAddress(s.Ctx, CONNECTION_0, PORT_0)
	s.Suite.Equal(acc1.String(), interchainAccount)
}

func checkUpgradeMintModule(s *UpgradeTestSuite) {
	acc1 := s.TestAccs[0]
	found := s.App.MintKeeper.IsAllowedAddress(s.Ctx, acc1.String())
	s.Suite.Equal(found, true)
}

func checkUpgradeTransferMiddlewareModule(s *UpgradeTestSuite) {
	acc1 := s.TestAccs[0]
	found := s.App.TransferMiddlewareKeeper.HasAllowRlyAddress(s.Ctx, acc1.String())
	s.Suite.Equal(found, true)
}

func checkUpgradePfmMiddlewareModule(s *UpgradeTestSuite) {
	data := s.App.RouterKeeper.GetAndClearInFlightPacket(s.Ctx, "channel-9", "transfer", 0)
	s.Suite.Equal("pica1wkjvpgkuchq0r8425g4z4sf6n85zj5wtykvtv3", data.OriginalSenderAddress)

	data = s.App.RouterKeeper.GetAndClearInFlightPacket(s.Ctx, "channel-9", "transfer", 2)
	s.Suite.Equal("pica1hj5fveer5cjtn4wd6wstzugjfdxzl0xpas3hgy", data.OriginalSenderAddress)
}

func checkUpgradeIbcTransferMiddlewareModule(s *UpgradeTestSuite) {
	data := s.App.IbcTransferMiddlewareKeeper.GetChannelFeeAddress(s.Ctx, "channel-9")
	s.Suite.Equal("pica1hj5fveer5cjtn4wd6wstzugjfdxzl0xpas3hgy", data)

	data = s.App.IbcTransferMiddlewareKeeper.GetChannelFeeAddress(s.Ctx, "channel-7")
	s.Suite.Equal("pica1hj5fveer5cjtn4wd6wstzugjfdxzl0xpas3hgy", data)
	data = s.App.IbcTransferMiddlewareKeeper.GetChannelFeeAddress(s.Ctx, "channel-1")
	s.Suite.Equal("", data)
}

func checkUpgradeIbcHooksMiddlewareModule(s *UpgradeTestSuite) {
	data := s.App.IBCHooksKeeper.GetPacketCallback(s.Ctx, "channel-2", 2)
	s.Suite.Equal("pica1hj5fveer5cjtn4wd6wstzugjfdxzl0xpas3hgy", data)

	data = s.App.IBCHooksKeeper.GetPacketCallback(s.Ctx, "channel-4", 2)
	s.Suite.Equal("pica1wkjvpgkuchq0r8425g4z4sf6n85zj5wtykvtv3", data)

	data = s.App.IBCHooksKeeper.GetPacketCallback(s.Ctx, "channel-2", 1)
	s.Suite.Equal("", data)
}

func CreateVestingAccount(s *UpgradeTestSuite,
) vestingtypes.ContinuousVestingAccount {
	str := `{"@type":"/cosmos.vesting.v1beta1.ContinuousVestingAccount","base_vesting_account":{"base_account":{"address":"centauri1alga5e8vr6ccr9yrg0kgxevpt5xgmgrvfkc5p8","pub_key":{"@type":"/cosmos.crypto.multisig.LegacyAminoPubKey","threshold":4,"public_keys":[{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AlnzK22KrkylnvTCvZZc8eZnydtQuzCWLjJJSMFUvVHf"},{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Aiw2Ftg+fnoHDU7M3b0VMRsI0qurXlerW0ahtfzSDZA4"},{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AvEHv+MVYRVau8FbBcJyG0ql85Tbbn7yhSA0VGmAY4ku"},{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Az5VHWqi3zMJu1rLGcu2EgNXLLN+al4Dy/lj6UZTzTCl"},{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"Ai4GlSH3uG+joMnAFbQC3jQeHl9FPvVTlRmwIFt7d7TI"},{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A2kAzH2bZr530jmFq/bRFrT2q8SRqdnfIebba+YIBqI1"}]},"account_number":46,"sequence":27},"original_vesting":[{"denom":"stake","amount":"22165200000000"}],"delegated_free":[{"denom":"stake","amount":"443382497453"}],"delegated_vesting":[{"denom":"stake","amount":"22129422502547"}],"end_time":1770994800},"start_time":1676300400}`

	var acc vestingtypes.ContinuousVestingAccount
	if err := json.Unmarshal([]byte(str), &acc); err != nil {
		panic(err)
	}

	err := banktestutil.FundAccount(s.App.BankKeeper, s.Ctx, acc.BaseAccount.GetAddress(),
		acc.GetOriginalVesting())
	if err != nil {
		panic(err)
	}

	err = banktestutil.FundAccount(s.App.BankKeeper, s.Ctx, acc.BaseAccount.GetAddress(),
		sdk.NewCoins(sdk.NewCoin(COIN_DENOM, math.NewIntFromUint64(1))))
	if err != nil {
		panic(err)
	}

	s.App.AccountKeeper.SetAccount(s.Ctx, &acc)
	return acc
}
