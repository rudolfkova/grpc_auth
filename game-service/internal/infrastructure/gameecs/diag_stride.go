package gameecs

// diagStrideState чинит «слишком быструю» диагональ на сетке: интент (dx≠0 и dy≠0)
// выполняется как лесенка — по одной оси за применение движения, фаза меняется только после успешного шага.
type diagStrideState struct {
	// yNext[uid]: false — следующий подшаг по X; true — по Y.
	yNext map[int64]bool
}

func (d *diagStrideState) pick(uid int64, dx, dy int) (adx, ady int, split bool) {
	if dx == 0 || dy == 0 {
		if d.yNext != nil {
			delete(d.yNext, uid)
		}
		return dx, dy, false
	}
	if d.yNext == nil {
		d.yNext = make(map[int64]bool)
	}
	sx, sy := signInt(dx), signInt(dy)
	if !d.yNext[uid] {
		return sx, 0, true
	}
	return 0, sy, true
}

func (d *diagStrideState) onApplied(uid int64, split, moved bool) {
	if !split || !moved {
		return
	}
	d.yNext[uid] = !d.yNext[uid]
}

func signInt(v int) int {
	if v < 0 {
		return -1
	}
	if v > 0 {
		return 1
	}
	return 0
}
