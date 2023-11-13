package ddm

import (
	"fmt"
	"os/exec"
	"regexp"
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

const (
	mounted int = iota
	mountedButWrongPlace
	notMounted
	commandError
	pathCreated
	pathNotExists
	pathExists
)

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

func checkMountPathExistence(path string) int {
	lsPath := fmt.Sprintf(" ls %s", path)
	_, err := exec.Command("/bin/sh", "-c", lsPath).CombinedOutput()
	if err == nil {
		return pathExists
	} else {
		return pathNotExists
	}
}

func mkDirForMount(diskPath string) int {
	if checkMountPathExistence(diskPath) == pathExists {
		return pathExists
	}
	mkDir := fmt.Sprintf("sudo mkdir %s", diskPath)
	out, err := exec.Command("/bin/sh", "-c", mkDir).CombinedOutput()
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return commandError
	}
	return pathCreated
}

func checkMountStatus(uuid, path string) int {
	lsblkCmd := "sudo lsblk -o UUID,MOUNTPOINT"
	out, err := exec.Command("/bin/sh", "-c", lsblkCmd).CombinedOutput()
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return commandError
	}
	lsblkOut := strings.Split(string(out[:]), "\n")

	idx := slices.IndexFunc(lsblkOut, func(element string) bool {
		return strings.Contains(element, uuid)
	})
	if idx == -1 {
		return notMounted
	}

	uuidOut, pathOut := func() (string, string) {
		space := regexp.MustCompile(`\s+`)
		lsblkWoSpace := space.ReplaceAllString(lsblkOut[idx], " ")
		x := strings.Split(lsblkWoSpace, " ")
		return x[0], x[1]
	}()

	if uuidOut == uuid && pathOut == path {
		return mounted
	}

	return mountedButWrongPlace
}

func mountCommand(disk config.Disk) int {
	swch := checkMountStatus(disk.UUID, disk.Mount.Path)
	switch swch {
	case mounted, commandError, mountedButWrongPlace:
		return swch
	}
	// if checkAlreadyMounted(disk.UUID, disk.Mount.Path) {
	// 	return true
	// }
	if mkDirForMount(disk.Mount.Path) == commandError {
		return commandError
	}
	sudoMount := fmt.Sprintf("sudo mount UUID=%s %s", disk.UUID, disk.Mount.Path)
	out, err := exec.Command("/bin/sh", "-c", sudoMount).CombinedOutput()
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return commandError
	}
	return mounted
}

func (d *DDMObserver) preCheckBeforeMount(disks []config.Disk) {
	for _, disk := range disks {
		currentDiskStat := d.GetDiskStat(disk)
		currentDiskStat.Active = true
		if !checkDiskAvailability(disk.UUID) {
			currentDiskStat.Active = false
			currentDiskStat.InactiveReason = append(currentDiskStat.InactiveReason, "Disk UUID not found")
		}
		// if isExists := checkMountPathExistence(disk.Mount.Path); isExists {
		// 	currentDiskStat.Active = false
		// 	currentDiskStat.InactiveReason = append(currentDiskStat.InactiveReason, "Mount path already used ...")
		// }
		d.DiskStat = append(d.DiskStat, currentDiskStat)
	}
}

func (d *DDMObserver) Mount(c *config.DDMConfig) {
	for _, disk := range c.Disks {
		diskStat := d.GetDiskStat(disk)
		
		if !diskStat.Active {
			log.WithFields(log.Fields{
				"disk":   disk.Name,
				"reason": diskStat.InactiveReason,
			}).Debug("Disk skipped cuz inactive")
			continue
		}

		switch mountCommand(disk) {
		case commandError:
			diskStat.InactiveReason = append(diskStat.InactiveReason, "Command error happened")
		case notMounted:
			diskStat.InactiveReason = append(diskStat.InactiveReason, "Disk not or cannot mounted")
		case mountedButWrongPlace:
			diskStat.InactiveReason = append(diskStat.InactiveReason, "Disk already mounted somewhere else")
		}
		if len(diskStat.InactiveReason) > 0 {
			diskStat.Active = false
		}

	}
}

func (d *DDMObserver) Run(c *config.DDMConfig) {
	d.preCheckBeforeMount(c.Disks)
	d.Mount(c)

}
