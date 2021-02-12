package yoda

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	band "github.com/GeoDB-Limited/odin-core/app"
	"github.com/GeoDB-Limited/odin-core/x/oracle/types"
)

func TestGetSignBytesVerificationMessage(t *testing.T) {
	band.SetBech32AddressPrefixesAndBip44CoinType(sdk.GetConfig())
	validator, _ := sdk.ValAddressFromBech32("odinvaloper17rprjgtj0krfw3wyl9creueej6ca9dc4a65n80")
	vmsg := NewVerificationMessage("bandchain", validator, types.RequestID(1), types.ExternalID(1))
	expected := []byte(`{"chain_id":"bandchain","external_id":1,"request_id":1,"validator":"odinvaloper17rprjgtj0krfw3wyl9creueej6ca9dc4a65n80"}`)
	require.Equal(t, expected, vmsg.GetSignBytes())
}
