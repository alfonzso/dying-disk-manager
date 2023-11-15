package observer

type DDMObserver struct {
	DiskStat []DiskStat
}

type DiskStat struct {
	Name           string
	UUID           string
	Active         bool
	InactiveReason []string
}
