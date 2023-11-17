package linux

import (
	"fmt"
	"os/exec"
	"regexp"
	"slices"
	"strings"

	"github.com/alfonzso/dying-disk-manager/pkg/common"
	"github.com/alfonzso/dying-disk-manager/pkg/config"
	log "github.com/sirupsen/logrus"
)

type Linux struct {
	Exec *ExecCommandsType
}

//go:generate stringer -type=LinuxCommands
type LinuxCommands int

const (
	None    LinuxCommands = 0
	Mounted LinuxCommands = 1 << iota
	UMounted
	MountedButWrongPlace
	NotMounted
	UUIDNotExists
	CommandError
	CommandSuccess
	DiskAvailable
	DiskUnAvailable
	PathCreated
	PathNotExists
	PathExists
)

var DetailedLinuxType = map[LinuxCommands]string{
	CommandError:         "Command error happened",
	NotMounted:           "Disk not or cannot mounted",
	MountedButWrongPlace: "Disk already mounted somewhere else",
}

var MountOrCommandError = MountedButWrongPlace | NotMounted | UUIDNotExists | CommandError
var MountWillBeSkip = Mounted | MountedButWrongPlace | UUIDNotExists | CommandError

func (l LinuxCommands) IsSucceed() bool {
	return l == CommandSuccess
}

func (l LinuxCommands) IsFailed() bool {
	return l == CommandError
}

func (l LinuxCommands) IsPathExists() bool {
	return l == PathExists
}

func (err LinuxCommands) IsMountOrCommandError() bool {
	return (err & MountOrCommandError) == err
}

func (err LinuxCommands) IsMountWillBeSkip() bool {
	return (err & MountWillBeSkip) == err
}

func (l LinuxCommands) IsAvailable() bool {
	return l == DiskAvailable
}

func (l LinuxCommands) IsUnAvailable() bool {
	return l == DiskUnAvailable
}

type ExecCommandsType struct {
	// GrepInList              func([]string, string) string
	// LsblkCMD                func() ([]string, LinuxCommands)
	checkDiskAvailability   func(string) ([]byte, error)
	checkMountPathExistence func(string) ([]byte, error)
	mkDir                   func(string) ([]byte, error)
	mount                   func(string) ([]byte, error)
	umount                  func(string) ([]byte, error)
	lsblk                   func(string) ([]byte, error)
	writeIntoDisk           func(string) ([]byte, error)
}

func NewExecCommand() *ExecCommandsType {
	basicCmd := func(command string) ([]byte, error) { return exec.Command("/bin/sh", "-c", command).CombinedOutput() }
	execC := &ExecCommandsType{
		checkDiskAvailability:   basicCmd,
		checkMountPathExistence: basicCmd,
		mkDir:                   basicCmd,
		mount:                   basicCmd,
		umount:                  basicCmd,
		lsblk:                   basicCmd,
		writeIntoDisk:           basicCmd,
	}
	// execC.LsblkCMD = execC.Lsblk
	// execC.GrepInList = common.GrepInList
	return execC
}

func (e ExecCommandsType) CheckDiskAvailability(uuid string) LinuxCommands {
	out, err := e.checkDiskAvailability("ls /dev/disk/by-uuid/")
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return CommandError
	}
	output := regexp.MustCompile(`[\n\t]`).ReplaceAllString(string(out[:]), "")
	if slices.Contains(strings.Split(output, " "), uuid) {
		return DiskAvailable
	}
	return DiskUnAvailable
}

func (e ExecCommandsType) CheckMountPathExistence(path string) LinuxCommands {
	_, err := e.checkMountPathExistence(fmt.Sprintf("ls %s", path))
	if err != nil {
		return PathNotExists
	}
	return PathExists
}

func (e ExecCommandsType) MkDir(diskPath string) LinuxCommands {
	if e.CheckMountPathExistence(diskPath).IsPathExists() {
		return PathExists
	}
	out, err := e.mkDir(fmt.Sprintf("sudo mkdir %s", diskPath))
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return CommandError
	}
	return PathCreated
}

func (e ExecCommandsType) Mount(disk config.Disk) LinuxCommands {
	out, err := e.mount(fmt.Sprintf("sudo mount UUID=%s %s", disk.UUID, disk.Mount.Path))
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return CommandError
	}
	return Mounted
}

func (e ExecCommandsType) UMount(disk config.Disk) LinuxCommands {
	out, err := e.umount(fmt.Sprintf("sudo umount -l %s", disk.Mount.Path))
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return CommandError
	}
	return UMounted
}

func (e ExecCommandsType) Lsblk() ([]string, LinuxCommands) {
	out, err := e.lsblk("sudo lsblk -o UUID,MOUNTPOINT")
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return []string{}, CommandError
	}
	return strings.Split(string(out[:]), "\n"), CommandSuccess
}

func (e ExecCommandsType) WriteIntoDisk(path string) LinuxCommands {
	out, err := e.writeIntoDisk(fmt.Sprintf(`sudo date > %s/.tstfile`, path))
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return CommandError
	}
	return CommandSuccess
}

func (e ExecCommandsType) CheckMountStatus(uuid, path string) LinuxCommands {
	lsblkOut, err := e.Lsblk()
	if err.IsFailed() {
		return err
	}

	lsblkFiltered := common.GrepInList(lsblkOut, uuid)
	if lsblkFiltered == "" {
		return UUIDNotExists
	}

	expectedUuidPath := []string{uuid, path}
	expectedNotMountedUuidPath := []string{uuid}
	resultUuidPath := common.DeleteEmpty(common.Split(lsblkFiltered, `\s+`))

	if common.IsEquals[string](expectedNotMountedUuidPath, resultUuidPath) {
		return NotMounted
	}

	if common.IsEquals[string](expectedUuidPath, resultUuidPath) {
		return Mounted
	}

	return MountedButWrongPlace
}

func (e ExecCommandsType) MountCommand(disk config.Disk) LinuxCommands {
	if mountStatus := e.CheckMountStatus(disk.UUID, disk.Mount.Path); mountStatus.IsMountWillBeSkip() {
		log.Debugf("[%s] Mount will be skiped", mountStatus)
		return mountStatus
	}
	if e.MkDir(disk.Mount.Path).IsFailed() {
		return CommandError
	}
	return e.Mount(disk)
}
