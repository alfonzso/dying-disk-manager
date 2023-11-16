// Code generated by "stringer -type=Linux"; DO NOT EDIT.

package linux

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[None-0]
	_ = x[Mounted-2]
	_ = x[UMounted-4]
	_ = x[MountedButWrongPlace-8]
	_ = x[NotMounted-16]
	_ = x[CommandError-32]
	_ = x[CommandSuccess-64]
	_ = x[PathCreated-128]
	_ = x[PathNotExists-256]
	_ = x[PathExists-512]
}

const (
	_Linux_name_0 = "None"
	_Linux_name_1 = "Mounted"
	_Linux_name_2 = "UMounted"
	_Linux_name_3 = "MountedButWrongPlace"
	_Linux_name_4 = "NotMounted"
	_Linux_name_5 = "CommandError"
	_Linux_name_6 = "CommandSuccess"
	_Linux_name_7 = "PathCreated"
	_Linux_name_8 = "PathNotExists"
	_Linux_name_9 = "PathExists"
)

func (i Linux) String() string {
	switch {
	case i == 0:
		return _Linux_name_0
	case i == 2:
		return _Linux_name_1
	case i == 4:
		return _Linux_name_2
	case i == 8:
		return _Linux_name_3
	case i == 16:
		return _Linux_name_4
	case i == 32:
		return _Linux_name_5
	case i == 64:
		return _Linux_name_6
	case i == 128:
		return _Linux_name_7
	case i == 256:
		return _Linux_name_8
	case i == 512:
		return _Linux_name_9
	default:
		return "Linux(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
