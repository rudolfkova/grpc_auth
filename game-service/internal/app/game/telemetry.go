package gameapp

import "time"

// Telemetry описывает наблюдаемость прикладного слоя и транспортных адаптеров.
// Реализации могут быть no-op или инфраструктурными (например Prometheus).
type Telemetry interface {
	ObserveTick(duration time.Duration, actionsInTick int)
	ObserveEventDropped(channel string)
	ObserveWSConnectionOpened()
	ObserveWSConnectionClosed()
	ObserveWSUpgradeFailure()
	ObserveActionRejected(reason string)
	ObserveSaveWorldResult(ok bool, code string)
	ObserveGRPCClientCall(service, method, grpcCode string, duration time.Duration)
}

type noopTelemetry struct{}

func (noopTelemetry) ObserveTick(time.Duration, int)                              {}
func (noopTelemetry) ObserveEventDropped(string)                                  {}
func (noopTelemetry) ObserveWSConnectionOpened()                                  {}
func (noopTelemetry) ObserveWSConnectionClosed()                                  {}
func (noopTelemetry) ObserveWSUpgradeFailure()                                    {}
func (noopTelemetry) ObserveActionRejected(string)                                {}
func (noopTelemetry) ObserveSaveWorldResult(bool, string)                         {}
func (noopTelemetry) ObserveGRPCClientCall(string, string, string, time.Duration) {}

func NewNoopTelemetry() Telemetry {
	return noopTelemetry{}
}
