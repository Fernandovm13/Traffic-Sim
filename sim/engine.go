package sim

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

// Engine: fachada pública del motor de simulación.
type Engine struct {
	mu sync.RWMutex

	// estado mutable
	cars   []*Car
	queues map[Dir][]*Car

	// semáforos y ocupación
	semNS *Semaphore
	semEW *Semaphore

	occupiedCount     int
	intersectionOwner Dir

	// pipeline
	jobs    chan job
	results chan jobResult

	// spawn
	spawnCh chan Dir
	SpawnCh chan<- Dir

	// snapshots (UI)
	snapshotCh chan Snapshot

	// lifecycle: context + cancel, waitgroup
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// fase principal
	phaseIdx   int
	phaseTimer int

	nextID int
}

// NewEngine crea el motor y canales (sin lanzar goroutines).
func NewEngine() *Engine {
	rand.Seed(time.Now().UnixNano())
	jobs := make(chan job, 512)
	results := make(chan jobResult, 512)
	spawn := make(chan Dir, 128)
	snapCh := make(chan Snapshot, 1)

	// create context for cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// duraciones en ticks (~60ms per tick)
	semNS := &Semaphore{GreenDur: 200, YellowDur: 36, state: SemGreen, timer: 200, Name: "NS"}
	semEW := &Semaphore{GreenDur: 160, YellowDur: 30, state: SemRed, timer: 160, Name: "EW"}

	e := &Engine{
		cars:       make([]*Car, 0, 128),
		queues:     map[Dir][]*Car{North: {}, East: {}, South: {}, West: {}},
		semNS:      semNS,
		semEW:      semEW,
		jobs:       jobs,
		results:    results,
		spawnCh:    spawn,
		SpawnCh:    spawn,
		snapshotCh: snapCh,
		ctx:        ctx,
		cancel:     cancel,
		wg:         sync.WaitGroup{},
		nextID:     1,
		phaseIdx:   0,
		phaseTimer: semNS.GreenDur,
	}
	return e
}

// SnapshotChan devuelve el canal donde el engine publica snapshots (solo-lectura).
func (e *Engine) SnapshotChan() <-chan Snapshot { return e.snapshotCh }

// Start lanza workers, spawnLoop y loop principal.
func (e *Engine) Start() {
	// workers
	numWorkers := 4
	for i := 0; i < numWorkers; i++ {
		e.wg.Add(1)
		go func() {
			defer e.wg.Done()
			worker(e)
		}()
	}
	// spawn loop
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		spawnLoop(e)
	}()

	// control loop
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		loop(e)
	}()
}

// Stop solicita cierre y espera goroutines.
func (e *Engine) Stop() {
	// Cancel context (idempotent)
	e.cancel()
	// Wait all goroutines finish
	e.wg.Wait()
}
