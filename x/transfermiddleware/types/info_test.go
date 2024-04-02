package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ComposableFi/composable-cosmos/v6/x/transfermiddleware/types"
)

func TestValidateBasic(t *testing.T) {
	var (
		validInfo = types.ParachainIBCTokenInfo{
			IbcDenom:    "ibc-test-1",
			ChannelID:   "channel-1",
			NativeDenom: "native-1",
			AssetId:     "1",
		}
		invalidInfo = types.ParachainIBCTokenInfo{
			IbcDenom:    "ibc-test-1",
			ChannelID:   "channel-1",
			NativeDenom: "native-1",
			AssetId:     "asset-1",
		}
	)
	testCases := map[string]struct {
		info types.ParachainIBCTokenInfo

		expectedErr bool
	}{
		"valid parachain info": {
			info: validInfo,
		},
		"invalid parachain info": {
			info:        invalidInfo,
			expectedErr: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			err := tc.info.ValidateBasic()

			// Assertions.
			if tc.expectedErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
