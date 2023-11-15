package observer

import (
	"fmt"
	"regexp"
)

type DDMObserver struct {
	DiskStat []DiskStat
}

const ()

type DiskStat struct {
	Name                string
	UUID                string
	Active              bool
	InactiveReason      []string
	TestThreadIsRunning bool
}

func (d DiskStat) String() string {
	msg := fmt.Sprintf(`
		|Name: %s
		|UUID: %s
		|Active: %t`,
		d.Name, d.UUID, d.Active,
	)
	regex, _ := regexp.Compile(`\t+[|]`)
	return regex.ReplaceAllString(msg, " ")

	// return msg
	// return strings.Replace(msg, "^\+s|", "", -1)
}
