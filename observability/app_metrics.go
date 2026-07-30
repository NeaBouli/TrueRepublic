package observability

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"truerepublic/token"
)

// Application metric contract (GH-80).
//
// Every collector lives in the "truerepublic" namespace and carries no labels.
// Label values derived from users, addresses, transactions, peers, or requests
// would create unbounded cardinality; these five series expose only public
// on-chain aggregate values, so the family count and cardinality stay fixed
// for the lifetime of the process.
//
// The metrics are updated exclusively after app.mm.EndBlock returns success.
// Recording never writes state, never alters the ABCI response, and a
// registration or recording problem must never become a consensus failure:
// collectors that failed registration still accept Set/Inc safely.
const (
	// AppMetricsNamespace is the Prometheus namespace for all TrueRepublic
	// application metrics.
	AppMetricsNamespace = "truerepublic"

	appMetricsSubsystem   = "app"
	tokenMetricsSubsystem = "token"

	// MetricLastSuccessfulBlockHeight is the height of the most recent block
	// whose application EndBlock completed without error.
	MetricLastSuccessfulBlockHeight = AppMetricsNamespace + "_" + appMetricsSubsystem + "_last_successful_block_height"
	// MetricCompletedBlocksTotal counts every application block whose EndBlock
	// completed without error.
	MetricCompletedBlocksTotal = AppMetricsNamespace + "_" + appMetricsSubsystem + "_completed_blocks_total"
	// MetricLastSuccessfulInvariantCycleHeight is the height of the most recent
	// block whose crisis invariant cycle passed. It equals the last successful
	// EndBlock height only because crisis is the last EndBlocker in
	// SetOrderEndBlockers and its invariant check period is one block: any
	// broken invariant panics inside crisis.AssertInvariants, so mm.EndBlock
	// cannot return success for a height whose invariant cycle failed.
	MetricLastSuccessfulInvariantCycleHeight = AppMetricsNamespace + "_" + appMetricsSubsystem + "_last_successful_invariant_cycle_height"
	// MetricPNYXSupplyBaseUnits is the canonical PNYX supply in base units
	// (upnyx), read from the bank module after a successful EndBlock.
	MetricPNYXSupplyBaseUnits = AppMetricsNamespace + "_" + tokenMetricsSubsystem + "_pnyx_supply_base_units"
	// MetricPNYXSupplyHeadroomBaseUnits is the remaining fixed-cap headroom in
	// base units: token.MaxSupplyBaseUnits minus the canonical supply.
	MetricPNYXSupplyHeadroomBaseUnits = AppMetricsNamespace + "_" + tokenMetricsSubsystem + "_pnyx_supply_headroom_base_units"
)

// AppMetrics bundles the bounded TrueRepublic application collectors.
type AppMetrics struct {
	lastSuccessfulBlockHeight          prometheus.Gauge
	completedBlocksTotal               prometheus.Counter
	lastSuccessfulInvariantCycleHeight prometheus.Gauge
	pnyxSupplyBaseUnits                prometheus.Gauge
	pnyxSupplyHeadroomBaseUnits        prometheus.Gauge
}

// newAppMetricsCollectors builds the collectors without registering them.
// Unregistered collectors still accept Set/Inc, which is the safe fallback
// when registration is impossible: recording becomes a local no-op instead of
// a consensus failure.
func newAppMetricsCollectors() *AppMetrics {
	return &AppMetrics{
		lastSuccessfulBlockHeight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: AppMetricsNamespace,
			Subsystem: appMetricsSubsystem,
			Name:      "last_successful_block_height",
			Help:      "Height of the last application block whose EndBlock completed successfully.",
		}),
		completedBlocksTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: AppMetricsNamespace,
			Subsystem: appMetricsSubsystem,
			Name:      "completed_blocks_total",
			Help:      "Total number of application blocks whose EndBlock completed successfully.",
		}),
		lastSuccessfulInvariantCycleHeight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: AppMetricsNamespace,
			Subsystem: appMetricsSubsystem,
			Name:      "last_successful_invariant_cycle_height",
			Help: "Height of the last block whose crisis invariant cycle passed. " +
				"Valid because crisis is the last EndBlocker and its check period is one block.",
		}),
		pnyxSupplyBaseUnits: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: AppMetricsNamespace,
			Subsystem: tokenMetricsSubsystem,
			Name:      "pnyx_supply_base_units",
			Help:      "Canonical PNYX supply in base units (upnyx) after the last successful EndBlock.",
		}),
		pnyxSupplyHeadroomBaseUnits: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: AppMetricsNamespace,
			Subsystem: tokenMetricsSubsystem,
			Name:      "pnyx_supply_headroom_base_units",
			Help:      "Remaining fixed-cap headroom in base units (token.MaxSupplyBaseUnits minus canonical supply).",
		}),
	}
}

// collectors returns every collector in a fixed order for registration.
func (m *AppMetrics) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		m.lastSuccessfulBlockHeight,
		m.completedBlocksTotal,
		m.lastSuccessfulInvariantCycleHeight,
		m.pnyxSupplyBaseUnits,
		m.pnyxSupplyHeadroomBaseUnits,
	}
}

// NewAppMetrics builds the application metrics and registers them with
// registerer. A nil registerer uses a fresh private registry, which is useful
// for isolated tests. Registration is all-or-nothing: on the first failure
// every already-registered collector is unregistered again and an error is
// returned, so callers never see a partially registered set.
func NewAppMetrics(registerer prometheus.Registerer) (*AppMetrics, error) {
	if registerer == nil {
		registerer = prometheus.NewRegistry()
	}
	metrics := newAppMetricsCollectors()
	registered := make([]prometheus.Collector, 0, len(metrics.collectors()))
	for _, collector := range metrics.collectors() {
		if err := registerer.Register(collector); err != nil {
			for _, rollback := range registered {
				registerer.Unregister(rollback)
			}
			return nil, fmt.Errorf("register application metrics: %w", err)
		}
		registered = append(registered, collector)
	}
	return metrics, nil
}

// RecordSuccessfulEndBlock updates every application metric after a fully
// successful app.mm.EndBlock for height. It must never be called on the error
// path: recording success for a failed block would corrupt both the block and
// the invariant-cycle signals.
//
// The invariant-cycle gauge is set to the same height because crisis runs
// last in SetOrderEndBlockers with check period one, so a successful
// mm.EndBlock return proves the invariant cycle at this height passed.
//
// supplyBaseUnits is the canonical bank supply of token.BaseDenom after the
// block. Headroom is derived as token.MaxSupplyBaseUnits - supplyBaseUnits;
// both fit in float64 exactly (MaxSupplyBaseUnits = 2.1e13 < 2^53). A negative
// headroom is exposed as-is: if the cap were ever exceeded, the metric must
// show the breach rather than hide it.
func (m *AppMetrics) RecordSuccessfulEndBlock(height int64, supplyBaseUnits int64) {
	if m == nil {
		return
	}
	m.lastSuccessfulBlockHeight.Set(float64(height))
	m.completedBlocksTotal.Inc()
	m.lastSuccessfulInvariantCycleHeight.Set(float64(height))
	m.pnyxSupplyBaseUnits.Set(float64(supplyBaseUnits))
	m.pnyxSupplyHeadroomBaseUnits.Set(float64(token.MaxSupplyBaseUnits - supplyBaseUnits))
}

var (
	defaultAppMetricsMu sync.Mutex
	defaultAppMetrics   *AppMetrics
)

// DefaultAppMetrics returns the process-wide application metrics registered on
// prometheus.DefaultRegisterer. The singleton makes repeated TrueRepublicApp
// constructions in the test suite safe: the collectors are created and
// registered exactly once, so no construction can panic or double-register.
//
// If registration on the default registry ever fails (for example a foreign
// collector already owns one of the names), the returned collectors stay
// unregistered. Recording on them remains safe, keeping a metrics problem from
// ever turning into a consensus failure.
func DefaultAppMetrics() *AppMetrics {
	defaultAppMetricsMu.Lock()
	defer defaultAppMetricsMu.Unlock()
	if defaultAppMetrics == nil {
		metrics, err := NewAppMetrics(prometheus.DefaultRegisterer)
		if err != nil {
			metrics = newAppMetricsCollectors()
		}
		defaultAppMetrics = metrics
	}
	return defaultAppMetrics
}
