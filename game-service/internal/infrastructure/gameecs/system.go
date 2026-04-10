package gameecs

// System — одна ECS-система; порядок вызовов задаёт SystemRegistry.
type System interface {
	Update(ctx *TickContext)
}
