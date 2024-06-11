package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)

	// RegisterLegacyAminoCodec(authzcodec.Amino)
	// RegisterLegacyAminoCodec(govcodec.Amino)
	amino.Seal()
}

// RegisterLegacyAminoCodec registers concrete types on the LegacyAmino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(Params{}, "composable/x/mintParams", nil)
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "composable/MsgUpdateParams")
	cdc.RegisterConcrete(&MsgFundModuleAccount{}, "composable/FundModuleAccount", nil)
	cdc.RegisterConcrete(&MsgAddAccountToFundModuleSet{}, "composable/MsgAddAccountToFundModuleSet", nil)
}

// RegisterInterfaces registers the interfaces types with the interface registry.
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
