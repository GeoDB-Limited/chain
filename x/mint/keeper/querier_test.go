package keeper_test

import (
	"github.com/GeoDB-Limited/odin-core/x/common/testapp"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/stretchr/testify/require"

	mintkeeper "github.com/GeoDB-Limited/odin-core/x/mint/keeper"
	minttypes "github.com/GeoDB-Limited/odin-core/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

func TestNewQuerier(t *testing.T) {
	app, ctx, _ := testapp.CreateTestInput(true)
	legacyQuerierCdc := codec.NewAminoCodec(app.LegacyAmino())
	querier := mintkeeper.NewQuerier(app.MintKeeper, legacyQuerierCdc.LegacyAmino)

	query := abci.RequestQuery{
		Path: "",
		Data: []byte{},
	}

	_, err := querier(ctx, []string{minttypes.QueryParams}, query)
	require.NoError(t, err)

	_, err = querier(ctx, []string{minttypes.QueryInflation}, query)
	require.NoError(t, err)

	_, err = querier(ctx, []string{minttypes.QueryAnnualProvisions}, query)
	require.NoError(t, err)

	_, err = querier(ctx, []string{minttypes.QueryIntegrationAddresses, "bsc"}, query)
	require.Error(t, err, "integration address not supported")

	_, err = querier(ctx, []string{"foo"}, query)
	require.Error(t, err)
}

func TestQueryParams(t *testing.T) {
	app, ctx, _ := testapp.CreateTestInput(true)
	legacyQuerierCdc := codec.NewAminoCodec(app.LegacyAmino())
	querier := mintkeeper.NewQuerier(app.MintKeeper, legacyQuerierCdc.LegacyAmino)

	var params minttypes.Params

	res, sdkErr := querier(ctx, []string{minttypes.QueryParams}, abci.RequestQuery{})
	require.NoError(t, sdkErr)

	err := app.LegacyAmino().UnmarshalJSON(res, &params)
	require.NoError(t, err)

	require.Equal(t, app.MintKeeper.GetParams(ctx), params)
}

func TestQueryInflation(t *testing.T) {
	app, ctx, _ := testapp.CreateTestInput(true)
	legacyQuerierCdc := codec.NewAminoCodec(app.LegacyAmino())
	querier := mintkeeper.NewQuerier(app.MintKeeper, legacyQuerierCdc.LegacyAmino)

	var inflation sdk.Dec

	res, sdkErr := querier(ctx, []string{minttypes.QueryInflation}, abci.RequestQuery{})
	require.NoError(t, sdkErr)

	err := app.LegacyAmino().UnmarshalJSON(res, &inflation)
	require.NoError(t, err)

	require.Equal(t, app.MintKeeper.GetMinter(ctx).Inflation, inflation)
}

func TestQueryAnnualProvisions(t *testing.T) {
	app, ctx, _ := testapp.CreateTestInput(true)
	legacyQuerierCdc := codec.NewAminoCodec(app.LegacyAmino())
	querier := mintkeeper.NewQuerier(app.MintKeeper, legacyQuerierCdc.LegacyAmino)

	var annualProvisions sdk.Dec

	res, sdkErr := querier(ctx, []string{minttypes.QueryAnnualProvisions}, abci.RequestQuery{})
	require.NoError(t, sdkErr)

	err := app.LegacyAmino().UnmarshalJSON(res, &annualProvisions)
	require.NoError(t, err)

	require.Equal(t, app.MintKeeper.GetMinter(ctx).AnnualProvisions, annualProvisions)
}

func TestQueryIntegrationAddresses(t *testing.T) {
	app, ctx, _ := testapp.CreateTestInput(true)
	legacyQuerierCdc := codec.NewAminoCodec(app.LegacyAmino())
	querier := mintkeeper.NewQuerier(app.MintKeeper, legacyQuerierCdc.LegacyAmino)

	_, sdkErr := querier(ctx, []string{minttypes.QueryIntegrationAddresses, "eth"}, abci.RequestQuery{})
	require.Error(t, sdkErr, "integration address not supported")
}
