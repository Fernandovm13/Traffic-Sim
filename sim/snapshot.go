package sim

// Snapshot: copia inmutable enviada a la UI
type Snapshot struct {
	Cars  []Car
	Light SnapshotLight
}

// SnapshotLight expone los estados de semáforo a UI
type SnapshotLight struct {
	NSState SemState
	EWState SemState
}
