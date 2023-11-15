package linux

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"github.com/alfonzso/dying-disk-manager/pkg/common"
	"github.com/alfonzso/dying-disk-manager/pkg/config"
	log "github.com/sirupsen/logrus"
)

//go:generate stringer -type=Linux
type Linux int

const (
	Mounted Linux = iota
	MountedButWrongPlace
	NotMounted
	CommandError
	PathCreated
	PathNotExists
	PathExists
)

func CheckDiskAvailability(uuid string) bool {
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

func CheckMountPathExistence(path string) Linux {
	lsPath := fmt.Sprintf(" ls %s", path)
	_, err := exec.Command("/bin/sh", "-c", lsPath).CombinedOutput()
	if err == nil {
		return PathExists
	} else {
		return PathNotExists
	}
}

func MkDirForMount(diskPath string) Linux {
	if CheckMountPathExistence(diskPath) == PathExists {
		return PathExists
	}
	mkDir := fmt.Sprintf("sudo mkdir %s", diskPath)
	out, err := exec.Command("/bin/sh", "-c", mkDir).CombinedOutput()
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return CommandError
	}
	return PathCreated
}

func Lsblk() ([]string, Linux) {
	lsblkCmd := "sudo lsblk -o UUID,MOUNTPOINT"
	out, err := exec.Command("/bin/sh", "-c", lsblkCmd).CombinedOutput()
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return []string{}, CommandError
	}
	return strings.Split(string(out[:]), "\n"), -1
}

func GrepInList(source []string, pattern string) string {
	idx := slices.IndexFunc(source, func(row string) bool {
		return strings.Contains(row, pattern)
	})
	if idx == -1 {
		return ""
	}
	return source[idx]
}

func CheckMountStatus(uuid, path string) Linux {
	lsblkOut, err := Lsblk()
	if err >= 0 {
		return err
	}

	lsblkFiltered := GrepInList(lsblkOut, uuid)
	if lsblkFiltered == "" {
		return NotMounted
	}

	expectedUuidPath := []string{uuid, path}
	resultUuidPath := common.Split(lsblkFiltered, `\s+`)

	if common.IsEquals[string](expectedUuidPath, resultUuidPath) {
		return Mounted
	}

	return MountedButWrongPlace
}

func MountCommand(disk config.Disk) Linux {
	swch := CheckMountStatus(disk.UUID, disk.Mount.Path)
	switch swch {
	case Mounted, CommandError, MountedButWrongPlace:
		log.Debugf("Skipping mount because: %s", swch)
		return swch
	}
	if MkDirForMount(disk.Mount.Path) == CommandError {
		return CommandError
	}
	sudoMount := fmt.Sprintf("sudo mount UUID=%s %s", disk.UUID, disk.Mount.Path)
	out, err := exec.Command("/bin/sh", "-c", sudoMount).CombinedOutput()
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return CommandError
	}
	return Mounted
}
