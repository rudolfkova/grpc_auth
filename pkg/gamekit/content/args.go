package content

import "encoding/json"

// RawJSONToArgsMap разбирает JSON-объект в map; не объект или ошибка → nil.
func RawJSONToArgsMap(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	return m
}

// MergeInteractBase: interact.args из каталога, затем поверх instance_args с тайла (тайл перекрывает каталог).
// В Runner.Run к результату ещё мержится args шага сценария (шаг перекрывает всё).
func MergeInteractBase(catalogArgs map[string]any, tileInstanceArgs json.RawMessage) map[string]any {
	tm := RawJSONToArgsMap(tileInstanceArgs)
	return mergeShallow(catalogArgs, tm)
}
