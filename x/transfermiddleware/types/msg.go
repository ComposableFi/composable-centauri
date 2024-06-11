package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
)

var _ sdk.Msg = &MsgAddParachainIBCTokenInfo{}

const (
	TypeMsgAddParachainIBCTokenInfo    = "add_para"
	TypeMsgRemoveParachainIBCTokenInfo = "remove_para"
	TypeMsgAddRlyAddress               = "add_rly_address"
)

func NewMsgAddParachainIBCTokenInfo(
	authority string,
	ibcDenom string,
	nativeDenom string,
	assetID string,
	channelID string,
) *MsgAddParachainIBCTokenInfo {
	return &MsgAddParachainIBCTokenInfo{
		Authority:   authority,
		IbcDenom:    ibcDenom,
		NativeDenom: nativeDenom,
		AssetId:     assetID,
		ChannelID:   channelID,
	}
}

// Route Implements Msg.
func (msg MsgAddParachainIBCTokenInfo) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgAddParachainIBCTokenInfo) Type() string { return TypeMsgAddParachainIBCTokenInfo }

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgAddParachainIBCTokenInfo) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgAddParachainIBCTokenInfo message.
func (msg *MsgAddParachainIBCTokenInfo) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgAddParachainIBCTokenInfo) ValidateBasic() error {
	// validate authority
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrap(err, "invalid authority address")
	}

	// validate channelIDs
	if err := host.ChannelIdentifierValidator(msg.ChannelID); err != nil {
		return err
	}

	// validate ibcDenom
	err := ibctransfertypes.ValidateIBCDenom(msg.IbcDenom)
	if err != nil {
		return err
	}

	return nil
}

var _ sdk.Msg = &MsgRemoveParachainIBCTokenInfo{}

func NewMsgRemoveParachainIBCTokenInfo(
	authority string,
	nativeDenom string,
) *MsgRemoveParachainIBCTokenInfo {
	return &MsgRemoveParachainIBCTokenInfo{
		Authority:   authority,
		NativeDenom: nativeDenom,
	}
}

// Route Implements Msg.
func (msg MsgRemoveParachainIBCTokenInfo) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgRemoveParachainIBCTokenInfo) Type() string { return TypeMsgRemoveParachainIBCTokenInfo }

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgRemoveParachainIBCTokenInfo) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgRemoveParachainIBCTokenInfo message.
func (msg *MsgRemoveParachainIBCTokenInfo) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgRemoveParachainIBCTokenInfo) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrap(err, "invalid authority address")
	}

	return nil
}

var _ sdk.Msg = &MsgAddRlyAddress{}

func NewMsgAddRlyAddress(
	authority string,
	rlyAdress string,
) *MsgAddRlyAddress {
	return &MsgAddRlyAddress{
		Authority:  authority,
		RlyAddress: rlyAdress,
	}
}

// Route Implements Msg.
func (msg MsgAddRlyAddress) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgAddRlyAddress) Type() string { return TypeMsgAddRlyAddress }

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgAddRlyAddress) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgRemoveParachainIBCTokenInfo message.
func (msg *MsgAddRlyAddress) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgAddRlyAddress) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return sdkerrors.Wrap(err, "invalid authority address")
	}

	if _, err := sdk.AccAddressFromBech32(msg.RlyAddress); err != nil {
		return sdkerrors.Wrap(err, "invalid authority address")
	}

	return nil
}
