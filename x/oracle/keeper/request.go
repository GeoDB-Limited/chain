package oraclekeeper

import (
	"github.com/GeoDB-Limited/odin-core/pkg/obi"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/pkg/errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	oracletypes "github.com/GeoDB-Limited/odin-core/x/oracle/types"
)

// HasRequest checks if the request of this ID exists in the storage.
func (k Keeper) HasRequest(ctx sdk.Context, id oracletypes.RequestID) bool {
	return ctx.KVStore(k.storeKey).Has(oracletypes.RequestStoreKey(id))
}

// GetRequest returns the request struct for the given ID or error if not exists.
func (k Keeper) GetRequest(ctx sdk.Context, id oracletypes.RequestID) (oracletypes.Request, error) {
	bz := ctx.KVStore(k.storeKey).Get(oracletypes.RequestStoreKey(id))
	if bz == nil {
		return oracletypes.Request{}, sdkerrors.Wrapf(oracletypes.ErrRequestNotFound, "id: %d", id)
	}
	var request oracletypes.Request
	k.cdc.MustUnmarshal(bz, &request)
	return request, nil
}

// MustGetRequest returns the request struct for the given ID. Panics error if not exists.
func (k Keeper) MustGetRequest(ctx sdk.Context, id oracletypes.RequestID) oracletypes.Request {
	request, err := k.GetRequest(ctx, id)
	if err != nil {
		panic(err)
	}
	return request
}

// GetPaginatedRequests returns all requests with pagination
func (k Keeper) GetPaginatedRequests(
	ctx sdk.Context,
	limit, offset uint64, reverse bool,
) ([]oracletypes.RequestResult, *query.PageResponse, error) {
	requests := make([]oracletypes.RequestResult, 0)
	requestsStore := prefix.NewStore(ctx.KVStore(k.storeKey), oracletypes.ResultStoreKeyPrefix)
	pagination := &query.PageRequest{
		Limit:   limit,
		Offset:  offset,
		Reverse: reverse,
	}

	pageRes, err := query.FilteredPaginate(requestsStore, pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
		var result oracletypes.Result
		obi.MustDecode(value, &result)

		request := oracletypes.RequestResult{
			RequestPacketData: &oracletypes.OracleRequestPacketData{
				ClientID:       result.ClientID,
				OracleScriptID: result.OracleScriptID,
				Calldata:       result.Calldata,
				AskCount:       result.AskCount,
				MinCount:       result.MinCount,
			},
			ResponsePacketData: &oracletypes.OracleResponsePacketData{
				RequestID:     result.RequestID,
				AnsCount:      result.AnsCount,
				RequestTime:   result.RequestTime,
				ResolveTime:   result.ResolveTime,
				ResolveStatus: result.ResolveStatus,
				Result:        result.Result,
			},
		}
		if accumulate {
			requests = append(requests, request)
		}
		return true, nil
	})
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to paginate requests")
	}

	return requests, pageRes, nil
}

// SetRequest saves the given data request to the store without performing any validation.
func (k Keeper) SetRequest(ctx sdk.Context, id oracletypes.RequestID, request oracletypes.Request) {
	request.ID = id
	ctx.KVStore(k.storeKey).Set(oracletypes.RequestStoreKey(id), k.cdc.MustMarshal(&request))
}

// DeleteRequest removes the given data request from the store.
func (k Keeper) DeleteRequest(ctx sdk.Context, id oracletypes.RequestID) {
	ctx.KVStore(k.storeKey).Delete(oracletypes.RequestStoreKey(id))
}

// AddRequest attempts to create and save a new request.
func (k Keeper) AddRequest(ctx sdk.Context, req oracletypes.Request) oracletypes.RequestID {
	id := k.GetNextRequestID(ctx)
	k.SetRequest(ctx, id, req)
	return id
}

// ProcessExpiredRequests resolves all expired requests and deactivates missed validators.
func (k Keeper) ProcessExpiredRequests(ctx sdk.Context) {
	currentReqID := k.GetRequestLastExpired(ctx) + 1
	lastReqID := oracletypes.RequestID(k.GetRequestCount(ctx))
	expirationBlockCount := int64(k.GetParamUint64(ctx, oracletypes.KeyExpirationBlockCount))
	// Loop through all data requests in chronological order. If a request reaches its
	// expiration height, we will deactivate validators that didn't report data on the
	// request. We also resolve requests to status EXPIRED if they are not yet resolved.
	for ; currentReqID <= lastReqID; currentReqID++ {
		req := k.MustGetRequest(ctx, currentReqID)
		// This request is not yet expired, so there's nothing to do here. Ditto for
		// all other requests that come after this. Thus we can just break the loop.
		if req.RequestHeight+expirationBlockCount > ctx.BlockHeight() {
			break
		}
		// If the request still does not have result, we resolve it as EXPIRED.
		if !k.HasResult(ctx, currentReqID) {
			k.ResolveExpired(ctx, currentReqID)
		}
		// Deactivate all validators that do not report to this request.
		for _, val := range req.RequestedValidators {
			v, _ := sdk.ValAddressFromBech32(val)
			if !k.HasReport(ctx, currentReqID, v) {
				k.MissReport(ctx, v, time.Unix(int64(req.RequestTime), 0))
			}
		}
		// Set last expired request ID to be this current request.
		k.SetRequestLastExpired(ctx, currentReqID)
	}
}

// AddPendingRequest adds the request to the pending list. DO NOT add same request more than once.
func (k Keeper) AddPendingRequest(ctx sdk.Context, id oracletypes.RequestID) {
	pendingList := k.GetPendingResolveList(ctx)
	pendingList = append(pendingList, id)
	k.SetPendingResolveList(ctx, pendingList)
}

// SetPendingResolveList saves the list of pending request that will be resolved at end block.
func (k Keeper) SetPendingResolveList(ctx sdk.Context, ids []oracletypes.RequestID) {
	intVs := make([]int64, len(ids))
	for idx, id := range ids {
		intVs[idx] = int64(id)
	}

	bz := k.cdc.MustMarshal(&oracletypes.PendingResolveList{RequestIds: intVs})
	if bz == nil {
		bz = []byte{}
	}
	ctx.KVStore(k.storeKey).Set(oracletypes.PendingResolveListStoreKey, bz)
}

// GetPendingResolveList returns the list of pending requests to be executed during EndBlock.
func (k Keeper) GetPendingResolveList(ctx sdk.Context) (ids []oracletypes.RequestID) {
	bz := ctx.KVStore(k.storeKey).Get(oracletypes.PendingResolveListStoreKey)
	if len(bz) == 0 { // Return an empty list if the key does not exist in the store.
		return []oracletypes.RequestID{}
	}
	pendingResolveList := oracletypes.PendingResolveList{}
	k.cdc.MustUnmarshal(bz, &pendingResolveList)
	for _, rid := range pendingResolveList.RequestIds {
		ids = append(ids, oracletypes.RequestID(rid))
	}
	return ids
}
