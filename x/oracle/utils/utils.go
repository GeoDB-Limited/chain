package utils

import sdk "github.com/cosmos/cosmos-sdk/types"

func CalculateReward(data []byte, pricePerByte sdk.DecCoin) sdk.DecCoin {
	pricePerByte.Amount = pricePerByte.Amount.MulInt64(int64(len(data)))
	return pricePerByte
}
