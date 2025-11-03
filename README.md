# Traffic-Sim 🚦

Simulación de tráfico concurrente escrita en Go + Ebiten

## 📋 Resumen

Traffic-Sim es una simulación de tráfico que muestra coches moviéndose y siendo controlados por semáforos. Está diseñada con una arquitectura limpia y concurrente: la lógica de simulación (paquete `sim`) corre en goroutines y comunica su estado a la capa de render (Ebiten, en `ui`) mediante snapshots no bloqueantes. El objetivo es enseñar y demostrar patrones de concurrencia seguros (worker-pool, productor-consumidor, fan-out/fan-in) y servir como base para experimentos (tuning de semáforos, pruebas unitarias, etc.).

## 🎯 Objetivos del Proyecto

- Simular un cruce con dos ejes (NS / EW) y dos semáforos coordinados (mutua exclusión)
- Mantener la lógica concurrente separada de la capa de render para evitar bloquear el hilo de UI
- Usar goroutines, canales y sincronización (mutex, WaitGroup, context) de forma segura
- Demostrar patrones: Worker Pool (fan-out/fan-in), Productor-Consumidor
- Proveer una base limpia, modular y fácil de extender/medir

## ✨ Características Principales

- **Dos semáforos coordinados** (NS / EW) con fases: GREEN → YELLOW → RED
- **Colas por dirección** (N, S, E, W) para que los coches no se encimen
- **Reserva de intersección** (campo `maxOccupy`) para evitar colisiones
- **Worker pool** que calcula posiciones de coches (procesamiento concurrente)
- **Snapshot inmutable** publicado por canal bufferizado (UI lee sin bloquear)
- **Cancelación con context.Context** y cierre ordenado con `sync.WaitGroup`
- **Código modular** dividido en archivos por responsabilidad (`sim/`, `ui/`, `main.go`)

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

## 🔌 API Pública (Resumen Rápido)

### Engine

- **`sim.NewEngine() *Engine`** — crea la instancia del motor (no inicia goroutines)
- **`engine.Start()`** — lanza workers, spawn loop y loop principal
- **`engine.Stop()`** — cancela el contexto y espera goroutines (cierre ordenado)
- **`engine.SpawnCh`** — canal para solicitar spawn de coche (ej.: `engine.SpawnCh <- sim.North`)
- **`engine.SnapshotChan()`** — canal de lectura desde el que la UI obtiene snapshots (no bloqueante)

## 🔄 Cómo Funciona (Flujo y Concurrencia)

### 1. Spawn (Productor)
- `spawnLoop()` corre en su propia goroutine (archivo `sim/spawn.go`)
- Produce coches periódicamente (ticker) o en respuesta a `SpawnCh`

### 2. Worker Pool + Jobs (Fan-out / Fan-in)
En cada tick (≈ 60 ms) el `loop()` (archivo `sim/control.go`):
- Hace una copia de estado `carsCopy` y fan-out jobs al canal `jobs`
- N trabajadores (`worker()` en `sim/worker.go`) procesan cada job (cálculo de la siguiente posición) y envían `jobResult` a `results`
- El `loop()` hace fan-in leyendo `results` y aplica los cambios al estado `e.cars`

### 3. Control de Semáforos y Reserva
- La fase de semáforos se gestiona centralmente en `loop()` para evitar doble verde
- Cuando el eje (NS o EW) está en GREEN, ambos carriles del eje (N y S o E y W) pueden reservar la intersección hasta `maxOccupy`

**Flujo por coche:**
- stop-line → semáforo GREEN → `crossingPoint` (dentro del cruce) → `exitTarget` (fuera) → liberación de ocupación al salir

### 4. Snapshots (Comunicación UI Segura)
- `loop()` copia `e.cars` y estado de semáforos y publica un `Snapshot` en `snapshotCh` usando `select { case snapshotCh <- snap: default: }` — no bloqueante
- `ui/game.go` en `Update()` lee `SnapshotChan()` de forma no bloqueante (`select { case s := <-...: default: }`) y almacena localmente `g.lastSnap` para `Draw()`

### 5. Cancelación y Cierre
- Engine usa `context.Context` + `cancel()` para señalizar parada a todas las goroutines
- `sync.WaitGroup` garantiza que `Stop()` espere hasta que todas las goroutines terminen

## 🎨 Patrones de Concurrencia — Explicación Detallada

A continuación se explica cada patrón usado, por qué se eligió y cómo está implementado en el código (archivos y funciones relevantes).

### Producer-Consumer (Productor-Consumidor)

**Qué es:**
Un productor añade trabajo a una cola (canal) y uno o varios consumidores procesan ese trabajo.

**Por qué lo usamos:**
El spawn de coches produce nuevas entidades (coches) y distintas goroutines deben procesarlas/actualizarlas. Separar productor y consumidor permite desacoplar ritmo de llegada de coches del ritmo de procesamiento.

**Cómo está implementado:**
- **Productor:** `spawnLoop()` (`sim/spawn.go`) que produce coches en intervalos o por `SpawnCh`
- **Consumidor:** `loop()` y `worker()` consumen datos mediante los canales `jobs` y `results`
- **Canal implicado:** `spawnCh` (para pedir creación), y `jobs`/`results` para el procesamiento por coche

### Worker Pool + Fan-out / Fan-in

**Qué es:**
Un pool de workers paraleliza procesamiento: el dispatcher fan-out envía trabajo a los workers, cada worker produce un resultado y el dispatcher hace fan-in para recogerlos y aplicar cambios.

**Por qué lo usamos:**
Calcular la posición y movimiento de muchos coches puede paralelizarse. Worker pool mantiene control sobre número de goroutines y evita crear una goroutine por coche cada tick.

**Cómo está implementado:**
- **Dispatcher:** `loop()` crea `carsCopy` y envía `job{car}` al canal `jobs` (fan-out). (`sim/control.go`)
- **Workers:** N goroutines que ejecutan `worker(e)` (`sim/worker.go`) consumen `jobs`, calculan `jobResult` y lo envían a `results`
- **Collector:** `loop()` consume `results` (fan-in) y aplica las posiciones a `e.cars`
- **Beneficio:** paralelismo controlado + menor latencia de cálculo por tick

### Snapshot publish/subscribe (última snapshot)

**Qué es:**
El motor publica su estado (snapshot) y la UI consume la última versión disponible sin bloquear al motor.

**Por qué lo usamos:**
Evita que la UI bloquee la simulación y evita bloqueos recíprocos entre render y lógica. También permite que la UI dibuje siempre una copia inmutable, sin tomar locks mientras dibuja.

**Cómo está implementado:**
- **Publisher:** `loop()` crea `snap := Snapshot{ Cars: make([]Car, len(e.cars)), Light: ... }` y lo envía a `snapshotCh` con `select ... default` (no espera si UI está ocupada). (`sim/control.go`)
- **Subscriber:** `ui/game.go` en `Update()` lee `SnapshotChan()` de forma no bloqueante y asigna `g.lastSnap`. `Draw()` dibuja `g.lastSnap` sin locks
- **Detalle:** `snapshotCh` está bufferizado con tamaño 1 para mantener la última snapshot posible

### Cancellation pattern (context + WaitGroup)

**Qué es:**
Uso de `context.Context` para señalizar cancelación a goroutines y `sync.WaitGroup` para esperar a que terminen.

**Por qué lo usamos:**
Permite cerrar la aplicación de forma ordenada, propagando la señal de parada a todas las goroutines y asegurando que ninguna quede huérfana.

**Cómo está implementado:**
- Engine contiene `ctx context.Context` y `cancel context.CancelFunc`. (`sim/engine.go`)
- Cada goroutine comprueba `case <-e.ctx.Done():` en sus `select` y sale
- `Engine.Stop()` llama `cancel()` y luego `wg.Wait()` para esperar que todas terminen

### Combinación de patrones en una pipeline segura

La arquitectura combina **Productor-Consumidor** (spawn), **Worker Pool** (jobs/results), **Snapshot publish/subscribe** (UI) y **Cancellation** (ctx) para crear una pipeline completa y segura. El motor es responsable de actualizar el estado y publicar snapshots; la UI solo consume, nunca bloquea, y la cancelación se propaga de forma centralizada.

## 🚀 Cómo Ejecutar

### Requisitos

- Go 1.20+ (o versión moderna)
- Módulo Go inicializado (`go.mod`)
- Ebiten v2 en `go.mod` (ej.: `github.com/hajimehoshi/ebiten/v2`)

### Instalar Dependencias
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

- **Espacio** — forzar spawn de un coche (cuando la ventana está enfocada)
- **Cerrar ventana** — la app finaliza; `engine.Stop()` se encarga de cancelar y esperar goroutines

## ⚙️ Parámetros y Tuning

### Duraciones de semáforos (ticks ≈ 60 ms):
- `semNS.GreenDur`, `semNS.YellowDur`
- `semEW.GreenDur`, `semEW.YellowDur`

**Conversión:** `ticks = int(seconds / 0.06)`

Ejemplo: para 8 segundos: `int(8.0 / 0.06) ≈ 133` ticks

### Otros Parámetros:

- **`maxOccupy`** (`sim/types.go`) — cuántos coches pueden reservar la intersección por eje (por defecto: 3)
- **`spawnLoop ticker`** — frecuencia de creación de coches (por defecto: 900 ms)
- **`Worker step`** — velocidad de coches (`step := 5.0` en `sim/worker.go`)
- **`queueGap`** — separación en fila entre coches


## ✅ Buenas Prácticas Ya Implementadas

- ✓ UI no bloquea la simulación (snapshot bufferizado + lectura no bloqueante)
- ✓ Protecciones contra data races (`sync.RWMutex` + snapshots por copia)
- ✓ Cancelación con context y cierre con `sync.WaitGroup`
- ✓ Separación clara de responsabilidades (`sim` vs `ui`)


---

Hecho con ❤️ usando Go + Ebiten