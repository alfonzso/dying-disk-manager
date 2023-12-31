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
	CantUmounted
	CantMounted
	MountedButWrongPlace
	NotMounted
	AlreadyMounted
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
	UUIDNotExists:        "Disk not found by UUID",
	DiskAvailable:        "Disk found by ls /dev/disk/by-uuid",
	DiskUnAvailable:      "Disk not found by ls /dev/disk/by-uuid",
}

var MountOrCommandError = MountedButWrongPlace | NotMounted | UUIDNotExists | CommandError
var MountWillBeSkip = Mounted | MountedButWrongPlace | UUIDNotExists | CommandError
var DiskUnAvailableOrUUIDNotExists = DiskUnAvailable | CommandError
var ForceRemountError = CantMounted | CantUmounted

func (l LinuxCommands) IsSucceed() bool {
	return l == CommandSuccess
}

func (l LinuxCommands) IsFailed() bool {
	return l == CommandError
}

func (l LinuxCommands) IsPathExists() bool {
	return l == PathExists
}

func (l LinuxCommands) IsAlreadyMounted() bool {
	return l == AlreadyMounted
}

func (l LinuxCommands) IsNotMounted() bool {
	return l == NotMounted
}

func (err LinuxCommands) IsMountOrCommandError() bool {
	return (err & MountOrCommandError) == err
}

func (err LinuxCommands) IsMountWillBeSkip() bool {
	return (err & MountWillBeSkip) == err
}

func (err LinuxCommands) IsDiskUnAvailableOrUUIDNotExists() bool {
	return (err & DiskUnAvailableOrUUIDNotExists) == err
}

func (err LinuxCommands) IsForceRemountError() bool {
	return (err & ForceRemountError) == err
}

func (l LinuxCommands) IsAvailable() bool {
	return l == DiskAvailable
}

func (l LinuxCommands) IsUnAvailable() bool {
	return l == DiskUnAvailable
}

type ExecCommandsType struct {
	checkDiskAvailability   func(string) ([]byte, error)
	checkMountPathExistence func(string) ([]byte, error)
	mkDir                   func(string) ([]byte, error)
	mount                   func(string) ([]byte, error)
	umount                  func(string) ([]byte, error)
	lsblk                   func(string) ([]byte, error)
	writeIntoDisk           func(string) ([]byte, error)
	runDryFsck              func(string) ([]byte, error)
	runFsck                 func(string) ([]byte, error)
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
		runDryFsck:              basicCmd,
		runFsck:                 basicCmd,
	}
	return execC
}

func (e ExecCommandsType) CheckDiskAvailability(uuid string) LinuxCommands {
	out, err := e.checkDiskAvailability("ls /dev/disk/by-uuid/")
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return CommandError
	}
	output := regexp.MustCompile(`[\n\t]`).ReplaceAllString(string(out[:]), " ")
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

func parseCommandError(name string, out []byte) LinuxCommands {
	_out := string(out)
	switch {
	case strings.Contains(_out, "not mounted"):
		log.Warnf("[%s] Disk not mounted", name)
		return NotMounted
	case strings.Contains(_out, "already mounted"):
		log.Warnf("[%s] Disk already mounted", name)
		return AlreadyMounted
	case strings.Contains(_out, "exit status 4"):
		log.Debugf("[%s] DryRunFsck", name)
	case strings.Contains(_out, "exit status 1"):
		log.Debugf("[%s] RunFsck", name)
	}
	return None
}

func (e ExecCommandsType) Mount(disk config.Disk) LinuxCommands {
	out, err := e.mount(fmt.Sprintf("sudo mount UUID=%s %s", disk.UUID, disk.Mount.Path))
	if err != nil {
		if !parseCommandError(disk.Name, out).IsAlreadyMounted() {
			log.Errorf(fmt.Sprint(err) + ": " + string(out))
			return CommandError
		}
	}
	return Mounted
}

func (e ExecCommandsType) UMount(disk config.Disk) LinuxCommands {
	out, err := e.umount(fmt.Sprintf("sudo umount -l %s", disk.Mount.Path))
	if err != nil {
		if !parseCommandError(disk.Name, out).IsNotMounted() {
			log.Errorf(fmt.Sprint(err) + ": " + string(out))
			return CommandError
		}
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
	out, err := e.writeIntoDisk(fmt.Sprintf(`sudo date > %s/.ddmfile`, path))
	if err != nil {
		log.Errorf(fmt.Sprint(err) + ": " + string(out))
		return CommandError
	}
	return CommandSuccess
}

func (e ExecCommandsType) RunDryFsck(uuid string) LinuxCommands {
	out, err := e.runDryFsck(fmt.Sprintf(`sudo fsck -fn /dev/disk/by-uuid/%s`, uuid))
	if err != nil {
		errAsStr := fmt.Sprintf("RunDryFsck - %s", err)
		if parseCommandError(uuid, []byte(errAsStr[:])) != None {
			log.Errorf("%s: %s", errAsStr, string(out))
		}
		return CommandError
	}
	return CommandSuccess
}

func (e ExecCommandsType) RunFsck(uuid string) LinuxCommands {
	out, err := e.runFsck(fmt.Sprintf(`sudo fsck -fy /dev/disk/by-uuid/%s`, uuid))
	if err != nil {
		errAsStr := fmt.Sprint(err)
		if parseCommandError(uuid, []byte("RunFsck-"+errAsStr[:])) != None {
			log.Errorf("%s: %s", errAsStr, string(out))
			return CommandError
		}
	}
	return CommandSuccess
}

func (e ExecCommandsType) GetDiskByUUIDWithError(uuid string) (string, LinuxCommands) {
	lsblkOut, err := e.Lsblk()
	if err.IsFailed() {
		return "", err
	}

	lsblkFiltered := common.GrepInList(lsblkOut, uuid)
	if lsblkFiltered == "" {
		return "", UUIDNotExists
	}
	return lsblkFiltered, CommandSuccess
}

func (e ExecCommandsType) GetDiskByUUID(uuid string) string {
	lsblkFiltered, _ := e.GetDiskByUUIDWithError(uuid)
	return lsblkFiltered
}

func (e ExecCommandsType) CheckMountStatus(uuid, path string) LinuxCommands {
	lsblkFiltered, err := e.GetDiskByUUIDWithError(uuid)
	if err != CommandSuccess {
		return err
	}

	expectedUuidPath := []string{uuid, path}
	expectedNotMountedUuidPath := []string{uuid}
	// resultUuidPath := common.DeleteEmpty(common.Split(lsblkFiltered, `\s+`))
	resultUuidPath := common.Maybe(lsblkFiltered).Split(`\s+`).DeleteEmpty("").GetList()

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
