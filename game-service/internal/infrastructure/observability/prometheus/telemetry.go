package prometheusobs

import (
	"strings"
	"time"

	gameapp "game/internal/app/game"

	"github.com/prometheus/client_golang/prometheus"
)

type Telemetry struct {
	ticksTotal           prometheus.Counter
	tickDuration         prometheus.Histogram
	wsConnectionsOpen    prometheus.Gauge
	wsConnectionsTotal   *prometheus.CounterVec
	wsUpgradeFailures    prometheus.Counter
	actionsRejectedTotal *prometheus.CounterVec
	eventsDroppedTotal   *prometheus.CounterVec
	saveWorldTotal       *prometheus.CounterVec
	grpcCallsTotal       *prometheus.CounterVec
	grpcDuration         *prometheus.HistogramVec
}

var _ gameapp.Telemetry = (*Telemetry)(nil)

func New(reg prometheus.Registerer) *Telemetry {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	t := &Telemetry{
		ticksTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "game_ticks_total",
			Help: "Total processed simulation ticks.",
		}),
		tickDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "game_tick_duration_seconds",
			Help:    "Tick loop duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}),
		wsConnectionsOpen: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "game_ws_connections_open",
			Help: "Currently opened websocket connections.",
		}),
		wsConnectionsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "game_ws_connections_total",
			Help: "Total websocket connection lifecycle events.",
		}, []string{"event"}),
		wsUpgradeFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "game_ws_upgrade_failures_total",
			Help: "Total failed websocket upgrades.",
		}),
		actionsRejectedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "game_actions_rejected_total",
			Help: "Total rejected incoming actions.",
		}, []string{"reason"}),
		eventsDroppedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "game_events_dropped_total",
			Help: "Total dropped outbound events due to backpressure.",
		}, []string{"channel"}),
		saveWorldTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "game_save_world_total",
			Help: "Total save_world requests.",
		}, []string{"outcome", "code"}),
		grpcCallsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "game_grpc_client_requests_total",
			Help: "Total outgoing gRPC calls from game-service.",
		}, []string{"service", "method", "grpc_code"}),
		grpcDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "game_grpc_client_duration_seconds",
			Help:    "Outgoing gRPC client call latency.",
			Buckets: prometheus.DefBuckets,
		}, []string{"service", "method"}),
	}
	reg.MustRegister(
		t.ticksTotal,
		t.tickDuration,
		t.wsConnectionsOpen,
		t.wsConnectionsTotal,
		t.wsUpgradeFailures,
		t.actionsRejectedTotal,
		t.eventsDroppedTotal,
		t.saveWorldTotal,
		t.grpcCallsTotal,
		t.grpcDuration,
	)
	return t
}

func (t *Telemetry) ObserveTick(duration time.Duration, _ int) {
	t.ticksTotal.Inc()
	t.tickDuration.Observe(duration.Seconds())
}

func (t *Telemetry) ObserveEventDropped(channel string) {
	t.eventsDroppedTotal.WithLabelValues(sanitize(channel)).Inc()
}

func (t *Telemetry) ObserveWSConnectionOpened() {
	t.wsConnectionsOpen.Inc()
	t.wsConnectionsTotal.WithLabelValues("opened").Inc()
}

func (t *Telemetry) ObserveWSConnectionClosed() {
	t.wsConnectionsOpen.Dec()
	t.wsConnectionsTotal.WithLabelValues("closed").Inc()
}

func (t *Telemetry) ObserveWSUpgradeFailure() {
	t.wsUpgradeFailures.Inc()
}

func (t *Telemetry) ObserveActionRejected(reason string) {
	t.actionsRejectedTotal.WithLabelValues(sanitize(reason)).Inc()
}

func (t *Telemetry) ObserveSaveWorldResult(ok bool, code string) {
	outcome := "error"
	if ok {
		outcome = "ok"
	}
	t.saveWorldTotal.WithLabelValues(outcome, sanitize(code)).Inc()
}

func (t *Telemetry) ObserveGRPCClientCall(service, method, grpcCode string, duration time.Duration) {
	svc := sanitize(service)
	mtd := sanitize(method)
	t.grpcCallsTotal.WithLabelValues(svc, mtd, sanitize(grpcCode)).Inc()
	t.grpcDuration.WithLabelValues(svc, mtd).Observe(duration.Seconds())
}

func sanitize(v string) string {
	if strings.TrimSpace(v) == "" {
		return "unknown"
	}
	return strings.TrimSpace(v)
}
