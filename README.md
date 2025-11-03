# Traffic-Sim 🚦

Simulación de tráfico concurrente escrita en Go + Ebiten

## 📋 Resumen

Traffic-Sim es una simulación de tráfico que muestra coches moviéndose y siendo controlados por semáforos. Está diseñada con una arquitectura limpia y concurrente: la lógica de simulación (paquete `sim`) corre en goroutines y comunica su estado a la capa de render (Ebiten, en `ui`) mediante snapshots no bloqueantes.

El objetivo es enseñar y demostrar patrones de concurrencia seguros (worker pool, productor-consumidor, fan-out/fan-in), y servir como base para experimentos (tuning de semáforos, lógica de intersecciones, tests).

## 🎯 Objetivos del Proyecto

- Simular un cruce con dos ejes (NS / EW) y dos semáforos coordinados (mutua exclusión)
- Mantener la lógica concurrente separada de la capa de render para evitar bloquear el hilo de UI
- Usar goroutines, canales y sincronización (mutex / waitgroup / context) de forma segura
- Demostrar patrones: Worker Pool (fan-out/fan-in), Productor-Consumidor
- Proveer una base limpia, modular y fácil de modificar/medir

## ✨ Características Principales

- **Dos semáforos coordinados** (NS / EW) con fases: GREEN → YELLOW → RED
- **Colas por dirección** (N, S, E, W) para que los coches no se encimen
- **Reserva de intersección** para evitar colisiones (capacidad `maxOccupy` configurable)
- **Worker pool** que calcula posiciones de coches (procesamiento concurrente)
- **Snapshot inmutable** publicado por canal bufferizado (UI lee sin bloquear)
- **Cancelación con context.Context** y cierre ordenado con `sync.WaitGroup`
- **Código dividido** en archivos por responsabilidad

## 📁 Estructura de Archivos
```
.
├── main.go              # arranque: crea engine, inicia, lanza Ebiten UI
├── go.mod
├── sim/
│   ├── types.go         # constantes, tipos (Dir, SemState, Semaphore)
│   ├── car.go           # struct Car
│   ├── aux.go           # funciones auxiliares (posiciones)
│   ├── snapshot.go      # Snapshot struct (Cars + Light)
│   ├── engine.go        # API pública: NewEngine, Start, Stop, SnapshotChan, SpawnCh
│   ├── spawn.go         # spawnLoop / createCar
│   ├── worker.go        # worker pool (job/result)
│   └── control.go       # loop principal: fases, fan-out/fan-in, control de tráfico
└── ui/
    ├── game.go          # Ebiten Game consumir SnapshotChan no bloqueante
    └── draw.go          # funciones de render (carretera, coches, semáforos)
```

## 🔌 API Pública

### Engine

- **`sim.NewEngine() *Engine`** — crea la instancia del motor (aún no lanza goroutines)
- **`engine.Start()`** — lanza workers, spawn loop y loop principal
- **`engine.Stop()`** — cancela el contexto y espera goroutines (cierre ordenado)
- **`engine.SpawnCh`** — canal para solicitar spawn de coche (ej.: `engine.SpawnCh <- sim.North`)
- **`engine.SnapshotChan()`** — canal de solo lectura desde el que la UI obtiene snapshots (no bloqueante)

### Flujo Principal

En `main.go` se crea el engine, se llama `Start()`, se crea `ui.NewGame(engine)` y se arranca Ebiten. Al cerrar la ventana de Ebiten, `RunGame` retorna y `engine.Stop()` se llama para cerrar todo limpio.

## 🔄 Cómo Funciona (Flujo y Concurrencia)

### 1. Spawn (Productor)
`spawnLoop` corre en su propia goroutine. Genera coches periódicamente o en respuesta a `SpawnCh` (productor).

### 2. Worker Pool + Jobs (Fan-out / Fan-in)
En cada tick (≈60 ms) el loop del engine:
- Crea una copia de estado (`carsCopy`) y fan-out jobs al canal `jobs`
- N trabajadores (goroutines `worker`) procesan cada job (cálculo de la nueva posición) y envían `jobResult` a `results`
- El loop hace fan-in leyendo `results` y actualiza el estado `e.cars`

Esto permite paralelizar cálculo por coche.

### 3. Control de Semáforos y Reserva
El loop mantiene la fase de semáforos (NS/EW) y, si un eje está en GREEN, permite que ambos carriles del eje (N y S o E y W) reserven la intersección hasta `maxOccupy`.

**Lógica:**
- Car at stop-line → semáforo GREEN → se asigna `crossingPoint` (punto dentro del cruce)
- Cuando llega → se le asigna `exitTarget` (fuera de pantalla)
- Al salir se libera la ocupación

### 4. Snapshots (Comunicación UI Segura)
En cada tick el engine hace una copia de `e.cars` y del estado de semáforos y la publica en `snapshotCh` sin bloquear (`select` con `default` para no esperar si la UI está ocupada).

La UI consume `SnapshotChan()` en `Update()` de Ebiten con `select { case s := <-...: default: }` por lo que nunca bloquea y `Draw()` usa la última snapshot (`g.lastSnap`) sin locks.

### 5. Cancelación y Cierre
Engine usa `context.Context` (`ctx`) y `cancel()` para señalizar el cierre a todas las goroutines; `wg.Wait()` asegura que todas terminen.

## 🎨 Patrones de Concurrencia Usados

- **Producer-Consumer**: `spawnLoop` produce coches; loop / workers consumen y procesan
- **Worker Pool (fan-out / fan-in)**: jobs → multiple workers → results → collector
- **Snapshot publish/subscribe**: engine produce snapshots; UI consume última snapshot
- **Cancellation pattern**: context + WaitGroup para cierre seguro

## 🚀 Cómo Ejecutar

### Requisitos

- Go 1.20+ (o versión moderna)
- Módulo Go inicializado (deberías tener `go.mod`)
- Ebiten v2 como dependencia

### Instalación de Dependencias
```bash
go mod tidy
```

### Ejecutar
```bash
go run .
```

### Compilar
```bash
go build -o traffic-sim
./traffic-sim
```

## 🎮 Controles

- **Espacio** (si está FOCUSED) — forzar spawn de un coche (usa `SpawnCh` en UI)
- **Cerrar ventana** — detener la simulación; el engine hará `Stop()` en `main.go`

## ⚙️ Parámetros de Configuración (Tuning)

### En `sim/engine.go` (o en `NewEngine()`):

**Duraciones de semáforos** (en ticks, 1 tick ≈ 60 ms):
- `semNS.GreenDur`, `semNS.YellowDur`
- `semEW.GreenDur`, `semEW.YellowDur`

**Conversión:** `ticks = int(seconds / 0.06)`

Ejemplo: para 8 segundos: `int(8.0 / 0.06) ≈ 133` ticks

### Otros Parámetros:

- **`maxOccupy`** (en `sim/types.go`) — cuántos coches pueden reservar la intersección a la vez por eje (por defecto: 3)
- **`spawnLoop ticker`** (en `sim/spawn.go`) — frecuencia de creación de coches (por defecto: 900 ms)
- **`worker step`** (velocidad) — `step := 5.0` en `sim/worker.go`
- **`queueGap`** — separación entre coches en fila

## ✅ Buenas Prácticas Implementadas

- ✓ UI nunca bloquea esperando la simulación (snapshot bufferizado + lectura no bloqueante)
- ✓ Evitamos data races con `sync.RWMutex` y snapshots por copia
- ✓ Cancelación con context evita fugas de goroutines al cerrar
- ✓ `sync.WaitGroup` garantiza cierre ordenado


---

Hecho con ❤️ usando Go y Ebiten