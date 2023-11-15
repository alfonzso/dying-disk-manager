package observer

import (
	"slices"

	"github.com/alfonzso/dying-disk-manager/pkg/config"
)

func (d *DDMObserver) GetDiskStat(disk config.Disk) DiskStat {
	idx := slices.IndexFunc(d.DiskStat, func(c DiskStat) bool { return c.Name == disk.Name })
	if idx == -1 {
		return DiskStat{Name: disk.Name, UUID: disk.UUID}
	}
	return d.DiskStat[idx]
}

func New() *DDMObserver {
	d := &DDMObserver{DiskStat: []DiskStat{}}
	return d
}
