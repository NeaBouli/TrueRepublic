package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	"truerepublic/x/dex"
	"truerepublic/x/truedemocracy"
)

func TestSupportedModuleGRPCQueryRoutesAreRegistered(t *testing.T) {
	app := newGenesisTestApp(t)
	routes := []string{
		"/truedemocracy.Query/Domain",
		"/truedemocracy.Query/Domains",
		"/truedemocracy.Query/Validator",
		"/truedemocracy.Query/Validators",
		"/truedemocracy.Query/Nullifier",
		"/truedemocracy.Query/PurgeSchedule",
		"/truedemocracy.Query/ZKPState",
		"/truedemocracy.Query/MerkleProof",
		"/truedemocracy.Query/PayToPut",
		"/dex.Query/Pool",
		"/dex.Query/Pools",
		"/dex.Query/RegisteredAssets",
		"/dex.Query/AssetByDenom",
		"/dex.Query/AssetBySymbol",
		"/dex.Query/EstimateSwap",
		"/dex.Query/PoolStats",
		"/dex.Query/SpotPrice",
		"/dex.Query/LiquidityDepth",
		"/dex.Query/LPPosition",
	}

	for _, route := range routes {
		if app.GRPCQueryRouter().Route(route) == nil {
			t.Errorf("supported gRPC query route %q is not registered", route)
		}
	}
}

func TestRetiredCustomABCIQueryPathsFailClosed(t *testing.T) {
	app := newGenesisTestApp(t)
	paths := []string{
		"custom/truedemocracy/domains",
		"/custom/truedemocracy/domains",
		"custom/dex/pools",
		"/custom/dex/pools",
	}

	for _, path := range paths {
		resp, err := app.Query(context.Background(), &abci.RequestQuery{Path: path})
		if err != nil {
			t.Fatalf("query %q returned transport error: %v", path, err)
		}
		if resp.Code == 0 {
			t.Errorf("retired query path %q unexpectedly succeeded", path)
		}
	}
}

func TestSupportedModuleGRPCQueriesExecuteThroughABCI(t *testing.T) {
	app := newGenesisTestApp(t)
	if err := initGenesisApp(app, exactlyBackedGenesisForApp(t, app)); err != nil {
		t.Fatal(err)
	}
	if _, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: 1,
		Time:   time.Date(2026, 8, 7, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Commit(); err != nil {
		t.Fatal(err)
	}

	tdRequest, err := app.appCodec.Marshal(&truedemocracy.QueryDomainsRequest{})
	if err != nil {
		t.Fatal(err)
	}
	tdResponse, err := app.Query(context.Background(), &abci.RequestQuery{
		Path: "/truedemocracy.Query/Domains",
		Data: tdRequest,
	})
	if err != nil {
		t.Fatal(err)
	}
	if tdResponse.Code != 0 {
		t.Fatalf("truedemocracy gRPC-over-ABCI query failed: code=%d log=%s", tdResponse.Code, tdResponse.Log)
	}
	var domainsResponse truedemocracy.QueryDomainsResponse
	if err := app.appCodec.Unmarshal(tdResponse.Value, &domainsResponse); err != nil {
		t.Fatal(err)
	}
	var domains []truedemocracy.Domain
	if err := json.Unmarshal(domainsResponse.Result, &domains); err != nil {
		t.Fatal(err)
	}
	if len(domains) != 1 || domains[0].Name != "Test" {
		t.Fatalf("unexpected domain query result: %+v", domains)
	}

	payToPutRequest, err := app.appCodec.Marshal(&truedemocracy.QueryPayToPutRequest{DomainName: "Test"})
	if err != nil {
		t.Fatal(err)
	}
	payToPutResponse, err := app.Query(context.Background(), &abci.RequestQuery{
		Path: "/truedemocracy.Query/PayToPut",
		Data: payToPutRequest,
	})
	if err != nil {
		t.Fatal(err)
	}
	if payToPutResponse.Code != 0 {
		t.Fatalf("PayToPut gRPC-over-ABCI query failed: code=%d log=%s", payToPutResponse.Code, payToPutResponse.Log)
	}
	var payToPutResult truedemocracy.QueryPayToPutResponse
	if err := app.appCodec.Unmarshal(payToPutResponse.Value, &payToPutResult); err != nil {
		t.Fatal(err)
	}
	var payToPut truedemocracy.PayToPutResult
	if err := json.Unmarshal(payToPutResult.Result, &payToPut); err != nil {
		t.Fatal(err)
	}
	// Genesis domain "Test": treasury 1000 upnyx, one member, so
	// base = 1000/1000 = 1 and final = 1 * min(15, 1) = 1.
	if payToPut.BaseCost != "1" || payToPut.FinalCost != "1" || payToPut.DomainMultiplier != 1 {
		t.Fatalf("unexpected PayToPut query result: %+v", payToPut)
	}

	merkleProofRequest, err := app.appCodec.Marshal(&truedemocracy.QueryMerkleProofRequest{
		DomainName: "Test",
		Commitment: "0000000000000000000000000000000000000000000000000000000000000000",
	})
	if err != nil {
		t.Fatal(err)
	}
	merkleProofResponse, err := app.Query(context.Background(), &abci.RequestQuery{
		Path: "/truedemocracy.Query/MerkleProof",
		Data: merkleProofRequest,
	})
	if err != nil {
		t.Fatal(err)
	}
	// Genesis domain "Test" has no identity commitments, so the proof
	// query must fail closed instead of fabricating a path.
	if merkleProofResponse.Code == 0 {
		t.Fatal("MerkleProof query unexpectedly succeeded for a domain without identity commitments")
	}

	dexRequest, err := app.appCodec.Marshal(&dex.QueryPoolsRequest{})
	if err != nil {
		t.Fatal(err)
	}
	dexResponse, err := app.Query(context.Background(), &abci.RequestQuery{
		Path: "/dex.Query/Pools",
		Data: dexRequest,
	})
	if err != nil {
		t.Fatal(err)
	}
	if dexResponse.Code != 0 {
		t.Fatalf("DEX gRPC-over-ABCI query failed: code=%d log=%s", dexResponse.Code, dexResponse.Log)
	}
	var poolsResponse dex.QueryPoolsResponse
	if err := app.appCodec.Unmarshal(dexResponse.Value, &poolsResponse); err != nil {
		t.Fatal(err)
	}
	var pools []dex.Pool
	if err := json.Unmarshal(poolsResponse.Result, &pools); err != nil {
		t.Fatal(err)
	}
	if len(pools) != 1 || pools[0].AssetDenom != "atom" {
		t.Fatalf("unexpected DEX query result: %+v", pools)
	}
}
