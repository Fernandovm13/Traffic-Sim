package sim

type Snapshot struct {
	Cars  []Car
	Light SnapshotLight
}

type SnapshotLight struct {
	NSState SemState
	EWState SemState
}
