// Code generated by "stringer -type=Linux"; DO NOT EDIT.

package linux

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Mounted-0]
	_ = x[MountedButWrongPlace-1]
	_ = x[NotMounted-2]
	_ = x[CommandError-3]
	_ = x[PathCreated-4]
	_ = x[PathNotExists-5]
	_ = x[PathExists-6]
}

const _Linux_name = "MountedMountedButWrongPlaceNotMountedCommandErrorPathCreatedPathNotExistsPathExists"

var _Linux_index = [...]uint8{0, 7, 27, 37, 49, 60, 73, 83}

func (i Linux) String() string {
	if i < 0 || i >= Linux(len(_Linux_index)-1) {
		return "Linux(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Linux_name[_Linux_index[i]:_Linux_index[i+1]]
}
