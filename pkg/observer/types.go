package observer

type DDMObserver struct {
	DiskStat []DiskStat
}

const ()

type DiskStat struct {
	Name                string
	UUID                string
	Active              bool
	InactiveReason      []string
	TestThreadIsRunning bool
}
