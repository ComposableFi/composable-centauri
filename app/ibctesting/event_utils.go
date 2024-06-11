package ibctesting

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
)

func getAckPackets(evts []abci.Event) []PacketAck {
	var res []PacketAck
	for _, evt := range evts {
		if evt.Type == "write_acknowledgement" {
			packet := parsePacketFromEvent(evt)
			ack := PacketAck{
				Packet: packet,
				Ack:    []byte(getField(evt, "packet_ack")),
			}
			res = append(res, ack)
		}
	}
	return res
}

// Used for various debug statements above when needed... do not remove
// func showEvent(evt abci.Event) {
//	fmt.Printf("evt.Type: %s\n", evt.Type)
//	for _, attr := range evt.Attributes {
//		fmt.Printf("  %s = %s\n", string(attr.Key), string(attr.Value))
//	}
//}

func parsePacketFromEvent(evt abci.Event) channeltypes.Packet {
	return channeltypes.Packet{
		Sequence:           getUintField(evt, "packet_sequence"),
		SourcePort:         getField(evt, "packet_src_port"),
		SourceChannel:      getField(evt, "packet_src_channel"),
		DestinationPort:    getField(evt, "packet_dst_port"),
		DestinationChannel: getField(evt, "packet_dst_channel"),
		Data:               []byte(getField(evt, "packet_data")),
		TimeoutHeight:      parseTimeoutHeight(getField(evt, "packet_timeout_height")),
		TimeoutTimestamp:   getUintField(evt, "packet_timeout_timestamp"),
	}
}

// ParsePacketFromEvents parses events emitted from a MsgRecvPacket and returns the
// acknowledgement.
func ParsePacketFromEvents(events sdk.Events) (channeltypes.Packet, error) {
	for _, ev := range events {
		if ev.Type == channeltypes.EventTypeSendPacket {
			packet := channeltypes.Packet{}
			for _, attr := range ev.Attributes {
				switch attr.Key {
				case channeltypes.AttributeKeyData:
					packet.Data = []byte(attr.Value)

				case channeltypes.AttributeKeySequence:
					seq, err := strconv.ParseUint(attr.Value, 10, 64)
					if err != nil {
						return channeltypes.Packet{}, err
					}

					packet.Sequence = seq

				case channeltypes.AttributeKeySrcPort:
					packet.SourcePort = attr.Value

				case channeltypes.AttributeKeySrcChannel:
					packet.SourceChannel = attr.Value

				case channeltypes.AttributeKeyDstPort:
					packet.DestinationPort = attr.Value

				case channeltypes.AttributeKeyDstChannel:
					packet.DestinationChannel = attr.Value

				case channeltypes.AttributeKeyTimeoutHeight:
					height, err := clienttypes.ParseHeight(attr.Value)
					if err != nil {
						return channeltypes.Packet{}, err
					}

					packet.TimeoutHeight = height

				case channeltypes.AttributeKeyTimeoutTimestamp:
					timestamp, err := strconv.ParseUint(attr.Value, 10, 64)
					if err != nil {
						return channeltypes.Packet{}, err
					}

					packet.TimeoutTimestamp = timestamp

				default:
					continue
				}
			}

			return packet, nil
		}
	}
	return channeltypes.Packet{}, fmt.Errorf("acknowledgement event attribute not found")
}

// return the value for the attribute with the given name
func getField(evt abci.Event, key string) string {
	for _, attr := range evt.Attributes {
		if attr.Key == key {
			return attr.Value
		}
	}
	return ""
}

func getUintField(evt abci.Event, key string) uint64 {
	raw := getField(evt, key)
	return toUint64(raw)
}

func toUint64(raw string) uint64 {
	if raw == "" {
		return 0
	}
	i, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		panic(err)
	}
	return i
}

func parseTimeoutHeight(raw string) clienttypes.Height {
	chunks := strings.Split(raw, "-")
	return clienttypes.Height{
		RevisionNumber: toUint64(chunks[0]),
		RevisionHeight: toUint64(chunks[1]),
	}
}

// ParseAckFromEvents parses events emitted from a MsgRecvPacket and returns the
// acknowledgement.
func ParseAckFromEvents(events []abci.Event) ([]byte, error) {
	for _, ev := range events {
		if ev.Type == channeltypes.EventTypeWriteAck {
			for _, attr := range ev.Attributes {
				if attr.Key == channeltypes.AttributeKeyAckHex {
					bz, err := hex.DecodeString(attr.Value)
					if err != nil {
						panic(err)
					}
					return bz, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("acknowledgement event attribute not found")
}

// ParseClientIDFromEvents parses events emitted from a MsgCreateClient and returns the
// client identifier.
func ParseClientIDFromEvents(events []abci.Event) (string, error) {
	for _, ev := range events {
		if ev.Type == clienttypes.EventTypeCreateClient {
			for _, attr := range ev.Attributes {
				if attr.Key == clienttypes.AttributeKeyClientID {
					return attr.Value, nil
				}
			}
		}
	}
	return "", fmt.Errorf("client identifier event attribute not found")
}

// ParseConnectionIDFromEvents parses events emitted from a MsgConnectionOpenInit or
// MsgConnectionOpenTry and returns the connection identifier.
func ParseConnectionIDFromEvents(events []abci.Event) (string, error) {
	for _, ev := range events {
		if ev.Type == connectiontypes.EventTypeConnectionOpenInit ||
			ev.Type == connectiontypes.EventTypeConnectionOpenTry {
			for _, attr := range ev.Attributes {
				if attr.Key == connectiontypes.AttributeKeyConnectionID {
					return attr.Value, nil
				}
			}
		}
	}
	return "", fmt.Errorf("connection identifier event attribute not found")
}

func GetSendPackets(evts []abci.Event) []channeltypes.Packet {
	var res []channeltypes.Packet
	for _, evt := range evts {
		if evt.Type == channeltypes.EventTypeSendPacket {
			packet := ParsePacketFromEvent(evt)
			res = append(res, packet)
		}
	}
	return res
}

// Used for various debug statements above when needed... do not remove
// func showEvent(evt abci.Event) {
//	fmt.Printf("evt.Type: %s\n", evt.Type)
//	for _, attr := range evt.Attributes {
//		fmt.Printf("  %s = %s\n", string(attr.Key), string(attr.Value))
//	}
//}

func ParsePacketFromEvent(evt abci.Event) channeltypes.Packet {
	return channeltypes.Packet{
		Sequence:           getUintField(evt, channeltypes.AttributeKeySequence),
		SourcePort:         getField(evt, channeltypes.AttributeKeySrcPort),
		SourceChannel:      getField(evt, channeltypes.AttributeKeySrcChannel),
		DestinationPort:    getField(evt, channeltypes.AttributeKeyDstPort),
		DestinationChannel: getField(evt, channeltypes.AttributeKeyDstChannel),
		Data:               getHexField(evt, channeltypes.AttributeKeyDataHex),
		TimeoutHeight:      parseTimeoutHeight(getField(evt, channeltypes.AttributeKeyTimeoutHeight)),
		TimeoutTimestamp:   getUintField(evt, channeltypes.AttributeKeyTimeoutTimestamp),
	}
}

func getHexField(evt abci.Event, key string) []byte {
	got := getField(evt, key)
	if got == "" {
		return nil
	}
	bz, err := hex.DecodeString(got)
	if err != nil {
		panic(err)
	}
	return bz
}

// ParseChannelIDFromEvents parses events emitted from a MsgChannelOpenInit or
// MsgChannelOpenTry and returns the channel identifier.
func ParseChannelIDFromEvents(events []abci.Event) (string, error) {
	for _, ev := range events {
		if ev.Type == channeltypes.EventTypeChannelOpenInit || ev.Type == channeltypes.EventTypeChannelOpenTry {
			for _, attr := range ev.Attributes {
				if attr.Key == channeltypes.AttributeKeyChannelID {
					return attr.Value, nil
				}
			}
		}
	}
	return "", fmt.Errorf("channel identifier event attribute not found")
}
