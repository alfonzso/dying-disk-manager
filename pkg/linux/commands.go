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
	None                 Linux = 0
	Mounted              Linux = 1 << iota
	UMounted                   // Linux = 2
	MountedButWrongPlace       // Linux = 4
	NotMounted                 // Linux = 8
	CommandError               // Linux = 16
	CommandSuccess             // Linux = 32
	PathCreated                // Linux = 64
	PathNotExists              // Linux = 128
	PathExists                 // Linux = 256
)

var DetailedLinuxType = map[Linux]string{
	CommandError:         "Command error happened",
	NotMounted:           "Disk not or cannot mounted",
	MountedButWrongPlace: "Disk already mounted somewhere else",
}

var MountOrCommandError = MountedButWrongPlace | NotMounted | CommandError
var MountWillBeSkip = Mounted | CommandError | MountedButWrongPlace

func (l Linux) IsSucceed() bool {
	return l == CommandSuccess
}

func (l Linux) IsFailed() bool {
	return l == CommandError
}

func (l Linux) IsPathExists() bool {
	return l == PathExists
}

func (err Linux) IsMountOrCommandError() bool {
	return (err & MountOrCommandError) == err
}

func (err Linux) IsMountWillBeSkip() bool {
	return (err & MountWillBeSkip) == err
}

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

func MkDir(diskPath string) Linux {
	if CheckMountPathExistence(diskPath).IsPathExists() {
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

func Mount(disk config.Disk) Linux {
	sudoMount := fmt.Sprintf("sudo mount UUID=%s %s", disk.UUID, disk.Mount.Path)
	out, err := exec.Command("/bin/sh", "-c", sudoMount).CombinedOutput()
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return CommandError
	}
	return Mounted
}

func UMount(disk config.Disk) Linux {
	sudoUmount := fmt.Sprintf("sudo umount -l %s", disk.Mount.Path)
	out, err := exec.Command("/bin/sh", "-c", sudoUmount).CombinedOutput()
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return CommandError
	}
	return UMounted
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

func WriteIntoDisk(path string) Linux {
	dateToFile := fmt.Sprintf(`date > %s/.tstfile`, path)
	out, err := exec.Command("sudo /bin/sh", "-c", dateToFile).CombinedOutput()
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return CommandError
	}
	return CommandSuccess
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
		return NotMounted // TODO create UUIDNotExists error
	}

	expectedUuidPath := []string{uuid, path}
	expectedNotMountedUuidPath := []string{uuid, ""}
	resultUuidPath := common.Split(lsblkFiltered, `\s+`)

	if common.IsEquals[string](expectedNotMountedUuidPath, resultUuidPath) {
		return NotMounted
	}

	if common.IsEquals[string](expectedUuidPath, resultUuidPath) {
		return Mounted
	}

	return MountedButWrongPlace
}

func MountCommand(disk config.Disk) Linux {
	if mountStatus := CheckMountStatus(disk.UUID, disk.Mount.Path); mountStatus.IsMountWillBeSkip() {
		log.Debugf("[%s] Mount will be skiped", mountStatus)
		return mountStatus
	}
	if MkDir(disk.Mount.Path).IsFailed() {
		return CommandError
	}
	return Mount(disk)
}
