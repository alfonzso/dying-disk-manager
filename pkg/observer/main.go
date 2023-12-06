package observer

func New() *DDMObserver {
	return &DDMObserver{DiskStat: []*DiskStat{}}
}
