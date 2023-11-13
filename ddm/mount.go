package ddm

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"github.com/alfonzso/dying-disk-manager/pkg/config"
	log "github.com/sirupsen/logrus"
)

type DDMObserver struct {
	DiskStat []DiskStat
}

type DiskStat struct {
	Name           string
	UUID           string
	Active         bool
	InactiveReason []string
}

func (d *DDMObserver) GetDiskStat(disk config.Disk) DiskStat {
	idx := slices.IndexFunc(d.DiskStat, func(c DiskStat) bool { return c.Name == disk.Name })
	if idx == -1 {
		return DiskStat{Name: disk.Name, UUID: disk.UUID}
	}
	return d.DiskStat[idx]
}

func checkDiskAvailability(uuid string) bool {
	out, err := exec.Command("/bin/sh", "-c", "ls /dev/disk/by-uuid/").CombinedOutput()
	// out, err := exec.Command("sudo echo faf").Output()
	// out, err := exec.Command("/bin/sh", "-c", "sudo echo faf").Output()
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return false
	}
	disks := strings.Split(string(out[:]), "\n")
	return slices.Contains(disks, uuid)
}

func checkMountPathExistence(path string) bool {
	lsPath := fmt.Sprintf(" ls %s", path)
	_, err := exec.Command("/bin/sh", "-c", lsPath).CombinedOutput()
	return err != nil
}

func mkDirForMount(diskPath string) bool {
	mkDir := fmt.Sprintf("sudo mkdir %s", diskPath)
	out, err := exec.Command("/bin/sh", "-c", mkDir).CombinedOutput()
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return true
	}
	return false
}

func mountCommand(disk config.Disk) bool {
	if mkDirErr := mkDirForMount(disk.Mount.Path); mkDirErr {
		return false
	}
	sudoMount := fmt.Sprintf("sudo mount %s %s", disk.UUID, disk.Mount.Path)
	out, err := exec.Command("/bin/sh", "-c", sudoMount).CombinedOutput()
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return false
	}
	return true
}

func (d *DDMObserver) preCheckBeforeMount(disks []config.Disk) {
	for _, disk := range disks {
		currentDiskStat := d.GetDiskStat(disk)
		currentDiskStat.Active = true
		if availability := checkDiskAvailability(disk.UUID); !availability {
			currentDiskStat.Active = false
			currentDiskStat.InactiveReason = append(currentDiskStat.InactiveReason, "Disk UUID not found")
		}
		if isExists := checkMountPathExistence(disk.Mount.Path); isExists {
			currentDiskStat.Active = false
			currentDiskStat.InactiveReason = append(currentDiskStat.InactiveReason, "Mount path already used ...")
		}
		d.DiskStat = append(d.DiskStat, currentDiskStat)
	}
}

func (d *DDMObserver) Mount(c *config.DDMConfig) {
	for _, disk := range c.Disks {
		diskStat := d.GetDiskStat(disk)
		if !diskStat.Active {
			log.Debugf("Disk skipped cuz inactive: %s", disk.Name)
			continue
		}
		if succeeded := mountCommand(disk); !succeeded {
			diskStat.Active = false
			diskStat.InactiveReason = append(diskStat.InactiveReason, "Mount failed")
		}

	}
}

func (d *DDMObserver) Run(c *config.DDMConfig) {
	d.preCheckBeforeMount(c.Disks)
	d.Mount(c)

}
