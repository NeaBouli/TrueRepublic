package observability

import (
	"strings"
	"sync"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"truerepublic/token"
)

// expectedFamilies pins the exact metric contract: name and type. Every
// family must carry zero labels so cardinality stays bounded regardless of
// chain activity.
var expectedFamilies = map[string]dto.MetricType{
	MetricLastSuccessfulBlockHeight:          dto.MetricType_GAUGE,
	MetricCompletedBlocksTotal:               dto.MetricType_COUNTER,
	MetricLastSuccessfulInvariantCycleHeight: dto.MetricType_GAUGE,
	MetricPNYXSupplyBaseUnits:                dto.MetricType_GAUGE,
	MetricPNYXSupplyHeadroomBaseUnits:        dto.MetricType_GAUGE,
}

func gather(t *testing.T, gatherer prometheus.Gatherer) map[string]*dto.MetricFamily {
	t.Helper()
	families, err := gatherer.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}
	byName := make(map[string]*dto.MetricFamily, len(families))
	for _, family := range families {
		byName[family.GetName()] = family
	}
	return byName
}

func familyValue(t *testing.T, families map[string]*dto.MetricFamily, name string) float64 {
	t.Helper()
	family, ok := families[name]
	if !ok {
		t.Fatalf("metric family %s missing", name)
	}
	if len(family.GetMetric()) != 1 {
		t.Fatalf("metric family %s has %d metrics, want exactly 1", name, len(family.GetMetric()))
	}
	metric := family.GetMetric()[0]
	switch family.GetType() {
	case dto.MetricType_GAUGE:
		return metric.GetGauge().GetValue()
	case dto.MetricType_COUNTER:
		return metric.GetCounter().GetValue()
	default:
		t.Fatalf("metric family %s has unexpected type %s", name, family.GetType())
		return 0
	}
}

func TestNewAppMetricsRegistersExactFamilies(t *testing.T) {
	registry := prometheus.NewRegistry()
	if _, err := NewAppMetrics(registry); err != nil {
		t.Fatalf("NewAppMetrics: %v", err)
	}

	families := gather(t, registry)
	if len(families) != len(expectedFamilies) {
		t.Fatalf("registered families = %d, want %d: %v", len(families), len(expectedFamilies), families)
	}
	for name, wantType := range expectedFamilies {
		family, ok := families[name]
		if !ok {
			t.Fatalf("metric family %s missing", name)
		}
		if family.GetType() != wantType {
			t.Fatalf("metric family %s type = %s, want %s", name, family.GetType(), wantType)
		}
		if family.GetHelp() == "" {
			t.Fatalf("metric family %s has no help text", name)
		}
		if len(family.GetMetric()) != 1 {
			t.Fatalf("metric family %s has %d metrics, want exactly 1", name, len(family.GetMetric()))
		}
		if labels := family.GetMetric()[0].GetLabel(); len(labels) != 0 {
			t.Fatalf("metric family %s has unbounded-label risk: %v", name, labels)
		}
	}
}

func TestRecordSuccessfulEndBlockProgressionAndValues(t *testing.T) {
	registry := prometheus.NewRegistry()
	metrics, err := NewAppMetrics(registry)
	if err != nil {
		t.Fatalf("NewAppMetrics: %v", err)
	}

	// Before any successful block, every series reads zero and the counter
	// has not moved.
	families := gather(t, registry)
	for name := range expectedFamilies {
		if value := familyValue(t, families, name); value != 0 {
			t.Fatalf("metric %s = %v before first block, want 0", name, value)
		}
	}

	metrics.RecordSuccessfulEndBlock(1, 1_000)
	metrics.RecordSuccessfulEndBlock(2, 1_500)
	metrics.RecordSuccessfulEndBlock(3, 1_500)

	families = gather(t, registry)
	if value := familyValue(t, families, MetricLastSuccessfulBlockHeight); value != 3 {
		t.Fatalf("last successful block height = %v, want 3", value)
	}
	if value := familyValue(t, families, MetricCompletedBlocksTotal); value != 3 {
		t.Fatalf("completed blocks total = %v, want 3", value)
	}
	// The invariant-cycle gauge tracks the block height exactly. This is only
	// sound because crisis is the last EndBlocker in app.SetOrderEndBlockers
	// and its check period is one block: a broken invariant panics inside
	// crisis.AssertInvariants, so app.mm.EndBlock cannot return success for a
	// height whose invariant cycle failed. The app-side coupling is guarded
	// by the root-package TestInvariantMetricCoupling.
	if value := familyValue(t, families, MetricLastSuccessfulInvariantCycleHeight); value != 3 {
		t.Fatalf("last successful invariant cycle height = %v, want 3", value)
	}
	if value := familyValue(t, families, MetricPNYXSupplyBaseUnits); value != 1_500 {
		t.Fatalf("pnyx supply base units = %v, want 1500", value)
	}
	wantHeadroom := float64(token.MaxSupplyBaseUnits - 1_500)
	if value := familyValue(t, families, MetricPNYXSupplyHeadroomBaseUnits); value != wantHeadroom {
		t.Fatalf("pnyx supply headroom = %v, want %v", value, wantHeadroom)
	}
}

func TestCapHeadroomArithmetic(t *testing.T) {
	// The cap must round-trip through float64 exactly; otherwise the gauge
	// contract would silently lose precision.
	if got := int64(float64(token.MaxSupplyBaseUnits)); got != token.MaxSupplyBaseUnits {
		t.Fatalf("token.MaxSupplyBaseUnits is not exact in float64: %d != %d", got, token.MaxSupplyBaseUnits)
	}

	registry := prometheus.NewRegistry()
	metrics, err := NewAppMetrics(registry)
	if err != nil {
		t.Fatalf("NewAppMetrics: %v", err)
	}

	cases := []struct {
		supply   int64
		headroom int64
	}{
		{supply: 0, headroom: token.MaxSupplyBaseUnits},
		{supply: 1, headroom: token.MaxSupplyBaseUnits - 1},
		{supply: token.MaxSupplyBaseUnits - 1, headroom: 1},
		{supply: token.MaxSupplyBaseUnits, headroom: 0},
	}
	for _, testCase := range cases {
		metrics.RecordSuccessfulEndBlock(1, testCase.supply)
		families := gather(t, registry)
		if value := familyValue(t, families, MetricPNYXSupplyBaseUnits); value != float64(testCase.supply) {
			t.Fatalf("supply %d: gauge = %v", testCase.supply, value)
		}
		if value := familyValue(t, families, MetricPNYXSupplyHeadroomBaseUnits); value != float64(testCase.headroom) {
			t.Fatalf("supply %d: headroom gauge = %v, want %d", testCase.supply, value, testCase.headroom)
		}
	}
}

func TestNewAppMetricsDuplicateRegistrationFailsAndRollsBack(t *testing.T) {
	registry := prometheus.NewRegistry()
	if _, err := NewAppMetrics(registry); err != nil {
		t.Fatalf("first NewAppMetrics: %v", err)
	}
	// A second registration on the same registry must fail cleanly instead of
	// panicking, and must leave the first registration intact.
	if _, err := NewAppMetrics(registry); err == nil {
		t.Fatal("second NewAppMetrics on the same registry must return an error")
	} else if !strings.Contains(err.Error(), "register application metrics") {
		t.Fatalf("error %v does not identify the metrics registration failure", err)
	}
	if families := gather(t, registry); len(families) != len(expectedFamilies) {
		t.Fatalf("families after duplicate registration = %d, want %d", len(families), len(expectedFamilies))
	}
}

func TestNewAppMetricsPartialCollisionRollsBackCompletely(t *testing.T) {
	registry := prometheus.NewRegistry()
	// Occupy only the last family name so NewAppMetrics registers four
	// collectors and then fails; all four must be unregistered again.
	conflicting := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: AppMetricsNamespace,
		Subsystem: tokenMetricsSubsystem,
		Name:      "pnyx_supply_headroom_base_units",
		Help:      "conflicting collector",
	})
	if err := registry.Register(conflicting); err != nil {
		t.Fatalf("register conflicting collector: %v", err)
	}

	if _, err := NewAppMetrics(registry); err == nil {
		t.Fatal("NewAppMetrics with a colliding family must return an error")
	}
	families := gather(t, registry)
	if len(families) != 1 {
		t.Fatalf("families after rollback = %d, want only the conflicting one: %v", len(families), families)
	}
	if _, ok := families[MetricPNYXSupplyHeadroomBaseUnits]; !ok {
		t.Fatalf("conflicting family missing after rollback: %v", families)
	}
}

func TestNewAppMetricsNilRegisterer(t *testing.T) {
	metrics, err := NewAppMetrics(nil)
	if err != nil {
		t.Fatalf("NewAppMetrics(nil): %v", err)
	}
	metrics.RecordSuccessfulEndBlock(7, 42)
}

func TestUnregisteredCollectorsRecordSafely(t *testing.T) {
	// The default-singleton fallback path keeps unregistered collectors when
	// registration is impossible; recording on them must stay a safe no-op.
	metrics := newAppMetricsCollectors()
	metrics.RecordSuccessfulEndBlock(1, 0)
	metrics.RecordSuccessfulEndBlock(2, token.MaxSupplyBaseUnits)
}

func TestRecordOnNilMetricsIsSafe(t *testing.T) {
	var metrics *AppMetrics
	metrics.RecordSuccessfulEndBlock(1, 0)
}

func TestDefaultAppMetricsSingletonAndConcurrentAccess(t *testing.T) {
	first := DefaultAppMetrics()
	second := DefaultAppMetrics()
	if first == nil {
		t.Fatal("DefaultAppMetrics returned nil")
	}
	if first != second {
		t.Fatal("DefaultAppMetrics must return one process-wide singleton")
	}

	// Concurrent singleton lookups and recordings must be race-free; run with
	// -race to enforce.
	var wait sync.WaitGroup
	for worker := 0; worker < 8; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			DefaultAppMetrics().RecordSuccessfulEndBlock(1, 0)
		}()
	}
	wait.Wait()
}
