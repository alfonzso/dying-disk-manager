package linux

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/alfonzso/dying-disk-manager/pkg/config"
)

func TestLinux_IsMountWillBeSkip(t *testing.T) {
	tests := []struct {
		name string
		err  LinuxCommands
		want bool
	}{
		{"OK_Mounted", Mounted, true},
		{"OK_CommandError", CommandError, true},
		{"OK_MountedButWrongPlace", MountedButWrongPlace, true},
		{"NOK_None", None, true},
		{"NOK_UMounted", UMounted, false},
		{"NOK_NotMounted", NotMounted, false},
		{"NOK_CommandSuccess", CommandSuccess, false},
		{"NOK_PathCreated", PathCreated, false},
		{"NOK_PathNotExists", PathNotExists, false},
		{"NOK_PathExists", PathExists, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.IsMountWillBeSkip(); got != tt.want {
				t.Errorf("Linux.IsMountWillBeSkip() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLinux_IsMountOrCommandError(t *testing.T) {
	tests := []struct {
		name string
		err  LinuxCommands
		want bool
	}{
		{"NOK_NotMounted", NotMounted, true},
		{"OK_MountedButWrongPlace", MountedButWrongPlace, true},
		{"OK_CommandError", CommandError, true},
		{"NOK_None", None, true},
		{"OK_Mounted", Mounted, false},
		{"NOK_UMounted", UMounted, false},
		{"NOK_CommandSuccess", CommandSuccess, false},
		{"NOK_PathCreated", PathCreated, false},
		{"NOK_PathNotExists", PathNotExists, false},
		{"NOK_PathExists", PathExists, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.IsMountOrCommandError(); got != tt.want {
				t.Errorf("Linux.IsMountOrCommandError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecCommandsType_CheckDiskAvailability(t *testing.T) {
	lslaByUUID := `UUID                                   MOUNTPOINT
	44fceed1-3277-4d53-8b1e-d953b6234a77   /mnt/disks001
	`
	type fields struct {
		checkDiskAvailability func(string) ([]byte, error)
	}
	type args struct {
		uuid string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"OK", fields{checkDiskAvailability: func(s string) ([]byte, error) { return []byte(lslaByUUID), nil }}, args{"no"}, false},
		{"OK", fields{checkDiskAvailability: func(s string) ([]byte, error) { return []byte(lslaByUUID), nil }}, args{"de67ce89-8a24-4068-a7c0-8c5d67eb1fac"}, true},
		{"OK", fields{checkDiskAvailability: func(s string) ([]byte, error) { return []byte(lslaByUUID), errors.New("eeee") }}, args{"..."}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ExecCommandsType{
				checkDiskAvailability: tt.fields.checkDiskAvailability,
			}
			if got := e.CheckDiskAvailability(tt.args.uuid); got != tt.want {
				t.Errorf("ExecCommandsType.CheckDiskAvailability() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecCommandsType_CheckMountPathExistence(t *testing.T) {
	type fields struct {
		checkMountPathExistence func(string) ([]byte, error)
	}
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   LinuxCommands
	}{
		{"PathExists", fields{checkMountPathExistence: func(lsPath string) ([]byte, error) { return []byte{123}, nil }}, args{path: "kek"}, PathExists},
		{"PathNotExists", fields{checkMountPathExistence: func(lsPath string) ([]byte, error) { return []byte{123}, errors.New("Fuck") }}, args{path: "kek"}, PathNotExists},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ExecCommandsType{
				checkMountPathExistence: tt.fields.checkMountPathExistence,
			}
			if got := e.CheckMountPathExistence(tt.args.path); got != tt.want {
				t.Errorf("ExecCommandsType.CheckMountPathExistence() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecCommandsType_MkDir(t *testing.T) {
	type fields struct {
		// checkDiskAvailability   func(string) ([]byte, error)
		checkMountPathExistence func(string) ([]byte, error)
		mkDir                   func(string) ([]byte, error)
		// mount                   func(string) ([]byte, error)
		// umount                  func(string) ([]byte, error)
		// lsblk                   func(string) ([]byte, error)
		// writeIntoDisk           func(string) ([]byte, error)
	}

	checkPathAndExists := func(lsPath string) ([]byte, error) { return []byte{123}, nil }
	checkPathAndNotExists := func(lsPath string) ([]byte, error) { return []byte{123}, errors.New("NotExists") }
	mkDirOk := func(lsPath string) ([]byte, error) { return []byte{104, 101, 101, 101, 101, 101}, nil }
	mkDirErr := func(lsPath string) ([]byte, error) {
		return []byte{104, 101, 101, 101, 101, 101}, errors.New("Mkdir failed lel")
	}

	type args struct {
		diskPath string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   LinuxCommands
	}{
		{"PathExists", fields{checkMountPathExistence: checkPathAndExists, mkDir: mkDirOk}, args{diskPath: "kek"}, PathExists},
		{"PathCreated", fields{checkMountPathExistence: checkPathAndNotExists, mkDir: mkDirOk}, args{diskPath: "kek"}, PathCreated},
		{"CommandError", fields{checkMountPathExistence: checkPathAndNotExists, mkDir: mkDirErr}, args{diskPath: "kek"}, CommandError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ExecCommandsType{
				// checkDiskAvailability:   tt.fields.checkDiskAvailability,
				checkMountPathExistence: tt.fields.checkMountPathExistence,
				mkDir:                   tt.fields.mkDir,
				// mount:                   tt.fields.mount,
				// umount:                  tt.fields.umount,
				// lsblk:                   tt.fields.lsblk,
				// writeIntoDisk:           tt.fields.writeIntoDisk,
			}
			if got := e.MkDir(tt.args.diskPath); got != tt.want {
				t.Errorf("ExecCommandsType.MkDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecCommandsType_Mount(t *testing.T) {
	type fields struct {
		// checkDiskAvailability   func(string) ([]byte, error)
		// checkMountPathExistence func(string) ([]byte, error)
		// mkDir                   func(string) ([]byte, error)
		mount func(string) ([]byte, error)
		// umount                  func(string) ([]byte, error)
		// lsblk                   func(string) ([]byte, error)
		// writeIntoDisk           func(string) ([]byte, error)
	}
	mountOk := func(lsPath string) ([]byte, error) { return []byte{104, 101, 101, 101, 101, 101}, nil }
	mountErr := func(lsPath string) ([]byte, error) {
		return []byte{104, 101, 101, 101, 101, 101}, errors.New("Mount errrr")
	}

	type args struct {
		disk config.Disk
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   LinuxCommands
	}{
		{"Mounted", fields{mount: mountOk}, args{disk: config.Disk{UUID: "keke", Mount: config.ExtendedMount{Path: "/mnt/fafa"}}}, Mounted},
		{"CommandError", fields{mount: mountErr}, args{disk: config.Disk{UUID: "keke", Mount: config.ExtendedMount{Path: "/mnt/fafa"}}}, CommandError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ExecCommandsType{
				// checkDiskAvailability:   tt.fields.checkDiskAvailability,
				// checkMountPathExistence: tt.fields.checkMountPathExistence,
				// mkDir:                   tt.fields.mkDir,
				mount: tt.fields.mount,
				// umount:                  tt.fields.umount,
				// lsblk:                   tt.fields.lsblk,
				// writeIntoDisk:           tt.fields.writeIntoDisk,
			}
			if got := e.Mount(tt.args.disk); got != tt.want {
				t.Errorf("ExecCommandsType.Mount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecCommandsType_UMount(t *testing.T) {
	type fields struct {
		// checkDiskAvailability   func(string) ([]byte, error)
		// checkMountPathExistence func(string) ([]byte, error)
		// mkDir                   func(string) ([]byte, error)
		// mount func(string) ([]byte, error)
		umount func(string) ([]byte, error)
		// lsblk                   func(string) ([]byte, error)
		// writeIntoDisk           func(string) ([]byte, error)
	}
	uMountOk := func(lsPath string) ([]byte, error) { return []byte{104, 101, 101, 101, 101, 101}, nil }
	uMountErr := func(lsPath string) ([]byte, error) {
		return []byte{104, 101, 101, 101, 101, 101}, errors.New("UMount errrr")
	}
	type args struct {
		disk config.Disk
	}
	defArgs := args{disk: config.Disk{UUID: "keke", Mount: config.ExtendedMount{Path: "/mnt/fafa"}}}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   LinuxCommands
	}{
		{"UMounted", fields{umount: uMountOk}, defArgs, UMounted},
		{"CommandError", fields{umount: uMountErr}, defArgs, CommandError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ExecCommandsType{
				// checkDiskAvailability:   tt.fields.checkDiskAvailability,
				// checkMountPathExistence: tt.fields.checkMountPathExistence,
				// mkDir:                   tt.fields.mkDir,
				// mount: tt.fields.mount,
				umount: tt.fields.umount,
				// lsblk:                   tt.fields.lsblk,
				// writeIntoDisk:           tt.fields.writeIntoDisk,
			}
			if got := e.UMount(tt.args.disk); got != tt.want {
				t.Errorf("ExecCommandsType.UMount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecCommandsType_Lsblk(t *testing.T) {
	type fields struct {
		lsblk func(string) ([]byte, error)
	}
	lslaByUUID := `UUID                                   MOUNTPOINT
44fceed1-3277-4d53-8b1e-d953b6234a77   /mnt/disks001
302d0ce6-30a1-43c8-b0c9-825da670f443   /`
	lsblkOk := func(lsPath string) ([]byte, error) { return []byte(lslaByUUID), nil }
	lsblkErr := func(lsPath string) ([]byte, error) {
		return []byte{104, 101, 101, 101, 101, 101}, errors.New("Lsblk errrr")
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
		want1  LinuxCommands
	}{
		{"lsblkOk", fields{lsblk: lsblkOk}, strings.Split(lslaByUUID, "\n"), CommandSuccess},
		{"lsblkErr", fields{lsblk: lsblkErr}, []string{}, CommandError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ExecCommandsType{
				lsblk: tt.fields.lsblk,
			}
			got, got1 := e.Lsblk()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExecCommandsType.Lsblk() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ExecCommandsType.Lsblk() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestExecCommandsType_WriteIntoDisk(t *testing.T) {
	type fields struct {
		// checkDiskAvailability   func(string) ([]byte, error)
		// checkMountPathExistence func(string) ([]byte, error)
		// mkDir                   func(string) ([]byte, error)
		// mount                   func(string) ([]byte, error)
		// umount                  func(string) ([]byte, error)
		// lsblk                   func(string) ([]byte, error)
		writeIntoDisk func(string) ([]byte, error)
	}
	type args struct {
		path string
	}
	writeIntoDiskOk := func(lsPath string) ([]byte, error) { return []byte{1, 2, 3, 4, 5}, nil }
	writeIntoDiskErr := func(lsPath string) ([]byte, error) {
		return []byte{104, 101, 101, 101, 101, 101}, errors.New("Lsblk errrr")
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   LinuxCommands
	}{
		{"lsblkOk", fields{writeIntoDisk: writeIntoDiskOk}, args{"fff"}, CommandSuccess},
		{"lsblkErr", fields{writeIntoDisk: writeIntoDiskErr}, args{"fff"}, CommandError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ExecCommandsType{
				// checkDiskAvailability:   tt.fields.checkDiskAvailability,
				// checkMountPathExistence: tt.fields.checkMountPathExistence,
				// mkDir:                   tt.fields.mkDir,
				// mount:                   tt.fields.mount,
				// umount:                  tt.fields.umount,
				// lsblk:                   tt.fields.lsblk,
				writeIntoDisk: tt.fields.writeIntoDisk,
			}
			if got := e.WriteIntoDisk(tt.args.path); got != tt.want {
				t.Errorf("ExecCommandsType.WriteIntoDisk() = %v, want %v", got, tt.want)
			}
		})
	}
}

// func TestGrepInList(t *testing.T) {
// 	type args struct {
// 		source  []string
// 		pattern string
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want string
// 	}{
// 		{"OK", args{source: []string{"kekek", "lelel", "fefe", "sasa aaaa feeee"}, pattern: "sasa"}, "sasa aaaa feeee"},
// 		{"NOK", args{source: []string{"kekek", "lelel", "fefe", "sasa aaaa feeee"}, pattern: "rrrrr"}, ""},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := common.GrepInList(tt.args.source, tt.args.pattern); got != tt.want {
// 				t.Errorf("GrepInList() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

func TestExecCommandsType_CheckMountStatus(t *testing.T) {
	lslaByUUID := `UUID                                   MOUNTPOINT
	44fceed1-3277-4d53-8b1e-d953b6234a77   /mnt/disks001
	`
	lslaByUUIDNotMounted := `UUID                                   MOUNTPOINT
	44fceed1-3277-4d53-8b1e-d953b6234a77
	`
	lsblkOk := func(lsPath string) ([]byte, error) { return []byte(lslaByUUID), nil }
	lsblkNotMounted := func(lsPath string) ([]byte, error) { return []byte(lslaByUUIDNotMounted), nil }

	type fields struct {
		lsblk func(string) ([]byte, error)
	}
	type args struct {
		uuid string
		path string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   LinuxCommands
	}{
		{"CheckMountStatusOk", fields{lsblk: lsblkOk}, args{uuid: "44fceed1-3277-4d53-8b1e-d953b6234a77", path: "/mnt/disks001"}, Mounted},
		{"CheckMountStatusMountedSomewhere", fields{lsblk: lsblkOk}, args{uuid: "44fceed1-3277-4d53-8b1e-d953b6234a77", path: "/mnt/disks/001"}, MountedButWrongPlace},
		{"CheckMountStatusMountedSomewhere", fields{lsblk: lsblkOk}, args{uuid: "44fceed1-327e-d953b6234a77", path: "/mnt/disks/001"}, UUIDNotExists},
		{"CheckMountStatusMountedSomewhere", fields{lsblk: lsblkNotMounted}, args{uuid: "44fceed1-3277-4d53-8b1e-d953b6234a77", path: "/mnt/disks/001"}, NotMounted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ExecCommandsType{
				lsblk: tt.fields.lsblk,
			}
			if got := e.CheckMountStatus(tt.args.uuid, tt.args.path); got != tt.want {
				t.Errorf("ExecCommandsType.CheckMountStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecCommandsType_MountCommand(t *testing.T) {
	lslaByUUID := `UUID                                   MOUNTPOINT
	44fceed1-3277-4d53-8b1e-d953b6234a77   /mnt/disks001
	`
	lslaByUUIDNotMounted := `UUID                                   MOUNTPOINT
	44fceed1-3277-4d53-8b1e-d953b6234a77
	`

	type fields struct {
		checkDiskAvailability   func(string) ([]byte, error)
		checkMountPathExistence func(string) ([]byte, error)
		mkDir                   func(string) ([]byte, error)
		mount                   func(string) ([]byte, error)
		lsblk                   func(string) ([]byte, error)
		// umount                  func(string) ([]byte, error)
		// writeIntoDisk           func(string) ([]byte, error)
	}
	type args struct {
		disk config.Disk
	}
	argsOK := args{disk: config.Disk{UUID: "44fceed1-3277-4d53-8b1e-d953b6234a77", Mount: config.ExtendedMount{Path: "/mnt/disks001"}}}
	// argsNotMounted := args{disk: config.Disk{UUID: "44fceed1-3277-4d53-8b1e-d953b6234a77", Mount: config.ExtendedMount{Path: "/mnt/disks001"}}}
	fieldsFunc := func(lsla string, mkdirErr error) fields {
		return fields{
			checkDiskAvailability:   func(s string) ([]byte, error) { return []byte(lsla), nil },
			checkMountPathExistence: func(lsPath string) ([]byte, error) { return []byte{123}, errors.New("PathNotExists") },
			mkDir:                   func(lsPath string) ([]byte, error) { return []byte{104, 101, 101, 101, 101, 101}, mkdirErr },
			mount:                   func(lsPath string) ([]byte, error) { return []byte{104, 101, 101, 101, 101, 101}, nil },
			lsblk:                   func(lsPath string) ([]byte, error) { return []byte(lsla), nil },
		}
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   LinuxCommands
	}{
		{"AlreadyMounted", fieldsFunc(lslaByUUID, nil), argsOK, Mounted},
		{"WillBeMounted", fieldsFunc(lslaByUUIDNotMounted, nil), argsOK, Mounted},
		{"MkdirError", fieldsFunc(lslaByUUIDNotMounted, errors.New("Mkdir fail")), argsOK, CommandError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := ExecCommandsType{
				checkDiskAvailability:   tt.fields.checkDiskAvailability,
				checkMountPathExistence: tt.fields.checkMountPathExistence,
				mkDir:                   tt.fields.mkDir,
				mount:                   tt.fields.mount,
				// umount:                  tt.fields.umount,
				lsblk: tt.fields.lsblk,
				// writeIntoDisk:           tt.fields.writeIntoDisk,
			}
			if got := e.MountCommand(tt.args.disk); got != tt.want {
				t.Errorf("ExecCommandsType.MountCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
