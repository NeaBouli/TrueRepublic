package dex

import (
	"encoding/json"
	"testing"

	"cosmossdk.io/math"
)

// ---------- RegisterAsset ----------

func TestRegisterAsset(t *testing.T) {
	k, ctx := setupKeeper(t)

	asset := RegisteredAsset{
		IBCDenom:       "ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2",
		Symbol:         "BTC",
		Name:           "Bitcoin",
		Decimals:       8,
		OriginChain:    "cosmoshub-4",
		IBCChannel:     "channel-0",
		TradingEnabled: true,
	}

	err := k.RegisterAsset(ctx, asset)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	retrieved, found := k.GetAssetByDenom(ctx, asset.IBCDenom)
	if !found {
		t.Fatal("asset not found after registration")
	}
	if retrieved.Symbol != "BTC" {
		t.Errorf("symbol = %s, want BTC", retrieved.Symbol)
	}
	if retrieved.Decimals != 8 {
		t.Errorf("decimals = %d, want 8", retrieved.Decimals)
	}
	if retrieved.OriginChain != "cosmoshub-4" {
		t.Errorf("origin_chain = %s, want cosmoshub-4", retrieved.OriginChain)
	}
	if retrieved.IBCChannel != "channel-0" {
		t.Errorf("ibc_channel = %s, want channel-0", retrieved.IBCChannel)
	}
}

func TestRegisterAssetDuplicate(t *testing.T) {
	k, ctx := setupKeeper(t)

	asset := RegisteredAsset{
		IBCDenom: "ibc/BTC123",
		Symbol:   "BTC",
		Name:     "Bitcoin",
		Decimals: 8,
	}

	err := k.RegisterAsset(ctx, asset)
	if err != nil {
		t.Fatalf("first registration failed: %v", err)
	}

	err = k.RegisterAsset(ctx, asset)
	if err == nil {
		t.Fatal("expected error for duplicate registration")
	}
}

func TestRegisterMultipleAssets(t *testing.T) {
	k, ctx := setupKeeper(t)

	assets := []RegisteredAsset{
		{IBCDenom: "ibc/BTC", Symbol: "BTC", Name: "Bitcoin", Decimals: 8, OriginChain: "cosmoshub-4"},
		{IBCDenom: "ibc/ETH", Symbol: "ETH", Name: "Ethereum", Decimals: 18, OriginChain: "ethereum"},
		{IBCDenom: "ibc/LUSD", Symbol: "LUSD", Name: "Liquity USD", Decimals: 18, OriginChain: "ethereum"},
	}

	for _, a := range assets {
		if err := k.RegisterAsset(ctx, a); err != nil {
			t.Fatalf("failed to register %s: %v", a.Symbol, err)
		}
	}

	allAssets := k.GetAllAssets(ctx)
	if len(allAssets) != 3 {
		t.Fatalf("expected 3 assets, got %d", len(allAssets))
	}

	symbols := make(map[string]bool)
	for _, a := range allAssets {
		symbols[a.Symbol] = true
	}
	for _, expected := range []string{"BTC", "ETH", "LUSD"} {
		if !symbols[expected] {
			t.Errorf("missing asset: %s", expected)
		}
	}
}

// ---------- GetAssetBySymbol ----------

func TestGetAssetBySymbol(t *testing.T) {
	k, ctx := setupKeeper(t)

	k.RegisterAsset(ctx, RegisteredAsset{IBCDenom: "ibc/BTC", Symbol: "BTC", Decimals: 8})
	k.RegisterAsset(ctx, RegisteredAsset{IBCDenom: "ibc/ETH", Symbol: "ETH", Decimals: 18})

	btc, found := k.GetAssetBySymbol(ctx, "BTC")
	if !found {
		t.Fatal("BTC not found by symbol")
	}
	if btc.IBCDenom != "ibc/BTC" {
		t.Errorf("denom = %s, want ibc/BTC", btc.IBCDenom)
	}

	eth, found := k.GetAssetBySymbol(ctx, "ETH")
	if !found {
		t.Fatal("ETH not found by symbol")
	}
	if eth.Decimals != 18 {
		t.Errorf("ETH decimals = %d, want 18", eth.Decimals)
	}

	_, found = k.GetAssetBySymbol(ctx, "UNKNOWN")
	if found {
		t.Fatal("should not find non-existent symbol")
	}
}

// ---------- UpdateAssetTradingStatus ----------

func TestUpdateAssetTradingStatus(t *testing.T) {
	k, ctx := setupKeeper(t)

	k.RegisterAsset(ctx, RegisteredAsset{
		IBCDenom:       "ibc/BTC",
		Symbol:         "BTC",
		TradingEnabled: true,
	})

	// Disable trading.
	err := k.UpdateAssetTradingStatus(ctx, "ibc/BTC", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	asset, _ := k.GetAssetByDenom(ctx, "ibc/BTC")
	if asset.TradingEnabled {
		t.Error("trading should be disabled")
	}

	// Re-enable.
	err = k.UpdateAssetTradingStatus(ctx, "ibc/BTC", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	asset, _ = k.GetAssetByDenom(ctx, "ibc/BTC")
	if !asset.TradingEnabled {
		t.Error("trading should be enabled")
	}
}

func TestUpdateAssetTradingStatusNotFound(t *testing.T) {
	k, ctx := setupKeeper(t)

	err := k.UpdateAssetTradingStatus(ctx, "ibc/NONEXISTENT", true)
	if err == nil {
		t.Fatal("expected error for non-existent asset")
	}
}

// ---------- DeregisterAsset ----------

func TestDeregisterAsset(t *testing.T) {
	k, ctx := setupKeeper(t)

	k.RegisterAsset(ctx, RegisteredAsset{IBCDenom: "ibc/BTC", Symbol: "BTC"})

	_, found := k.GetAssetByDenom(ctx, "ibc/BTC")
	if !found {
		t.Fatal("asset should exist before deregistration")
	}

	err := k.DeregisterAsset(ctx, "ibc/BTC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, found = k.GetAssetByDenom(ctx, "ibc/BTC")
	if found {
		t.Fatal("asset should be removed after deregistration")
	}
}

func TestDeregisterAssetNotFound(t *testing.T) {
	k, ctx := setupKeeper(t)

	err := k.DeregisterAsset(ctx, "ibc/NONEXISTENT")
	if err == nil {
		t.Fatal("expected error for non-existent asset")
	}
}

// ---------- Validation ----------

func TestRegisteredAssetValidation(t *testing.T) {
	tests := []struct {
		name    string
		asset   RegisteredAsset
		wantErr string
	}{
		{
			name:    "empty denom",
			asset:   RegisteredAsset{IBCDenom: "", Symbol: "BTC"},
			wantErr: "ibc_denom",
		},
		{
			name:    "empty symbol",
			asset:   RegisteredAsset{IBCDenom: "ibc/BTC", Symbol: ""},
			wantErr: "symbol",
		},
		{
			name:    "decimals too high",
			asset:   RegisteredAsset{IBCDenom: "ibc/BTC", Symbol: "BTC", Decimals: 19},
			wantErr: "decimals",
		},
		{
			name:  "valid asset",
			asset: RegisteredAsset{IBCDenom: "ibc/BTC", Symbol: "BTC", Decimals: 8},
		},
		{
			name:  "valid with max decimals",
			asset: RegisteredAsset{IBCDenom: "ibc/ETH", Symbol: "ETH", Decimals: 18},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.asset.ValidateBasic()
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if !containsStr(err.Error(), tt.wantErr) {
					t.Errorf("error %q should contain %q", err.Error(), tt.wantErr)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestRegisterAssetValidationRejectsInvalid(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Empty denom should fail at keeper level too.
	err := k.RegisterAsset(ctx, RegisteredAsset{IBCDenom: "", Symbol: "BTC"})
	if err == nil {
		t.Fatal("expected error for empty denom")
	}

	// Empty symbol should fail.
	err = k.RegisterAsset(ctx, RegisteredAsset{IBCDenom: "ibc/BTC", Symbol: ""})
	if err == nil {
		t.Fatal("expected error for empty symbol")
	}
}

// ---------- Genesis round-trip ----------

func TestAssetRegistryGenesisRoundTrip(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Register several assets.
	originals := []RegisteredAsset{
		{IBCDenom: "pnyx", Symbol: "PNYX", Name: "TrueRepublic Native Token", Decimals: 6, OriginChain: "truerepublic-1", TradingEnabled: true},
		{IBCDenom: "ibc/BTC", Symbol: "BTC", Name: "Bitcoin", Decimals: 8, OriginChain: "cosmoshub-4", IBCChannel: "channel-0", TradingEnabled: true},
		{IBCDenom: "ibc/ETH", Symbol: "ETH", Name: "Ethereum", Decimals: 18, OriginChain: "ethereum", IBCChannel: "channel-1", TradingEnabled: false},
	}

	for _, a := range originals {
		if err := k.RegisterAsset(ctx, a); err != nil {
			t.Fatalf("register %s: %v", a.Symbol, err)
		}
	}

	// Export.
	exported := k.GetAllAssets(ctx)
	if len(exported) != 3 {
		t.Fatalf("expected 3 exported assets, got %d", len(exported))
	}

	// Serialize and deserialize (simulating genesis export/import).
	genesis := GenesisState{
		Pools:            []Pool{},
		RegisteredAssets: exported,
	}
	bz, err := json.Marshal(genesis)
	if err != nil {
		t.Fatalf("marshal genesis: %v", err)
	}

	var restored GenesisState
	if err := json.Unmarshal(bz, &restored); err != nil {
		t.Fatalf("unmarshal genesis: %v", err)
	}

	if len(restored.RegisteredAssets) != 3 {
		t.Fatalf("expected 3 restored assets, got %d", len(restored.RegisteredAssets))
	}

	// Verify data integrity.
	symbolMap := make(map[string]RegisteredAsset)
	for _, a := range restored.RegisteredAssets {
		symbolMap[a.Symbol] = a
	}

	btc := symbolMap["BTC"]
	if btc.Decimals != 8 {
		t.Errorf("BTC decimals = %d, want 8", btc.Decimals)
	}
	if btc.OriginChain != "cosmoshub-4" {
		t.Errorf("BTC origin = %s, want cosmoshub-4", btc.OriginChain)
	}

	eth := symbolMap["ETH"]
	if eth.TradingEnabled {
		t.Error("ETH should have trading disabled")
	}
}

// ---------- DEX pool creation with IBC denom ----------

func TestCreatePoolWithIBCDenom(t *testing.T) {
	k, ctx := setupKeeper(t)

	// Register an IBC asset.
	k.RegisterAsset(ctx, RegisteredAsset{
		IBCDenom:       "ibc/BTC",
		Symbol:         "BTC",
		Decimals:       8,
		TradingEnabled: true,
	})

	// Create a pool using the IBC denom.
	err := k.CreatePool(ctx, "ibc/BTC", math.NewInt(500_000), math.NewInt(100_000))
	if err != nil {
		t.Fatalf("unexpected error creating IBC pool: %v", err)
	}

	pool, found := k.GetPool(ctx, "ibc/BTC")
	if !found {
		t.Fatal("IBC pool not found")
	}
	if !pool.PnyxReserve.Equal(math.NewInt(500_000)) {
		t.Errorf("pnyx reserve = %s, want 500000", pool.PnyxReserve)
	}
	if !pool.AssetReserve.Equal(math.NewInt(100_000)) {
		t.Errorf("asset reserve = %s, want 100000", pool.AssetReserve)
	}
}

func TestSwapWithIBCDenom(t *testing.T) {
	k, ctx := setupKeeper(t)

	k.CreatePool(ctx, "ibc/BTC", math.NewInt(1_000_000), math.NewInt(1_000_000))

	// Swap PNYX for IBC BTC.
	out, err := k.Swap(ctx, "pnyx", math.NewInt(10_000), "ibc/BTC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.IsPositive() {
		t.Fatal("output should be positive")
	}

	// Swap IBC BTC for PNYX.
	out, err = k.Swap(ctx, "ibc/BTC", math.NewInt(5_000), "pnyx")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !out.IsPositive() {
		t.Fatal("output should be positive")
	}
}

// ---------- Default genesis ----------

func TestDefaultGenesisIncludesPNYX(t *testing.T) {
	genesis := DefaultGenesisState()

	if len(genesis.RegisteredAssets) == 0 {
		t.Fatal("default genesis should include PNYX asset")
	}

	pnyx := genesis.RegisteredAssets[0]
	if pnyx.Symbol != "PNYX" {
		t.Errorf("first default asset symbol = %s, want PNYX", pnyx.Symbol)
	}
	if pnyx.IBCDenom != "pnyx" {
		t.Errorf("PNYX denom = %s, want pnyx", pnyx.IBCDenom)
	}
	if pnyx.Decimals != 6 {
		t.Errorf("PNYX decimals = %d, want 6", pnyx.Decimals)
	}
	if !pnyx.TradingEnabled {
		t.Error("PNYX trading should be enabled by default")
	}
}

// ---------- helpers ----------

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && searchStr(s, substr)
}

func searchStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
