package linux

import (
	"fmt"
	"os/exec"
	"regexp"
	"slices"
	"strings"

	"github.com/alfonzso/dying-disk-manager/pkg/config"
	log "github.com/sirupsen/logrus"
)

const (
	Mounted int = iota
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

func CheckMountPathExistence(path string) int {
	lsPath := fmt.Sprintf(" ls %s", path)
	_, err := exec.Command("/bin/sh", "-c", lsPath).CombinedOutput()
	if err == nil {
		return PathExists
	} else {
		return PathNotExists
	}
}

func MkDirForMount(diskPath string) int {
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

func CheckMountStatus(uuid, path string) int {
	lsblkCmd := "sudo lsblk -o UUID,MOUNTPOINT"
	out, err := exec.Command("/bin/sh", "-c", lsblkCmd).CombinedOutput()
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return CommandError
	}
	lsblkOut := strings.Split(string(out[:]), "\n")

	idx := slices.IndexFunc(lsblkOut, func(element string) bool {
		return strings.Contains(element, uuid)
	})
	if idx == -1 {
		return NotMounted
	}

	uuidOut, pathOut := func() (string, string) {
		space := regexp.MustCompile(`\s+`)
		lsblkWoSpace := space.ReplaceAllString(lsblkOut[idx], " ")
		x := strings.Split(lsblkWoSpace, " ")
		return x[0], x[1]
	}()

	if uuidOut == uuid && pathOut == path {
		return Mounted
	}

	return MountedButWrongPlace
}

func MountCommand(disk config.Disk) int {
	swch := CheckMountStatus(disk.UUID, disk.Mount.Path)
	switch swch {
	case Mounted, CommandError, MountedButWrongPlace:
		return swch
	}
	// if checkAlreadyMounted(disk.UUID, disk.Mount.Path) {
	// 	return true
	// }
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
