# Game Backend Plan (MVP)

## Общая идея

* 2D онлайн-игра для 1–4 игроков
* Цель: интерактивное погружение игроков в DnD персонажей
* Сервер полностью authoritative (вся логика на сервере)
* Клиент = только ввод + рендер

---

## Архитектура

### Сервисы

* auth-service (JWT, сессии)
* chat-service (реалтайм чат)
* gateway (WS + HTTP фасад)
* game-service (новый, игровая логика)

### Поток

Client → WS → Gateway → Game Service

---

## Транспорт

* Используем WebSocket (не UDP)
* Один WS endpoint (multiplex):

```json
{
  "service": "game",
  "type": "move",
  "payload": {}
}
```

---

## Авторизация

* JWT валидируется в gateway
* user_id прокидывается в контексте
* game-service НЕ доверяет данным из клиента

---

## Game Loop

* Tick rate: 20 TPS (50ms)

```go
ticker := time.NewTicker(50 * time.Millisecond)
for {
  <-ticker.C
  Update()
}
```

---

## Обработка действий

### Клиент → сервер

Только intent (не позиция):

```json
{
  "type": "move",
  "dx": 1,
  "dy": 0
}
```

### Pipeline

WS → Gateway → GameService → actions channel → Update()

---

## Очередь действий

```go
type Action struct {
  PlayerID int64
  Type string
  DX int
  DY int
}

actions chan Action
```

---

## Game State (MVP)

```go
map[playerID]Position
```

---

## Update() цикл

* собрать все actions из канала
* применить к состоянию
* обновить ECS (или простую логику)
* собрать snapshot

---

## Snapshot

Сервер → клиент:

```json
{
  "type": "state",
  "players": [
    { "id": 1, "x": 10, "y": 12 }
  ]
}
```

---

## Отправка событий

GameService → events channel → Gateway → WS clients

---

## Session Model (позже)

* session_id
* players[]
* state (CREATED / RUNNING / FINISHED)

---

## ECS (план)

### Components

* Position
* MoveIntent
* (позже) Behavior

### Systems

* MovementSystem
* (позже) BehaviorSystem

---

## Важные решения

* ❌ НЕ использовать UDP
* ✅ Использовать WebSocket
* ✅ Один gateway (не разделять сейчас)
* ✅ Server authoritative
* ✅ Клиент отправляет intent, не state
* ✅ Game loop через ticker
* ✅ Очередь действий (channel)
* ✅ Snapshot-based синхронизация

---

## MVP Scope

1. Подключение по WS
2. Отправка move (dx/dy)
3. Сервер обновляет позицию
4. Сервер шлёт snapshot
5. Клиент рисует

---

## Не делать сейчас

* сложный netcode
* WebRTC
* распределённость
* сложный ECS
* античит

---

## Следующие шаги

1. Реализовать game loop
2. Сделать actions channel
3. Прокинуть WS → game-service
4. Сделать минимальный state
5. Реализовать snapshot

---

## Расширение (потом)

* мультиплеер сессии
* чат интеграция
* логирование поведения
* replay системы
