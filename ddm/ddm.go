package ddm

import (
	"github.com/alfonzso/dying-disk-manager/pkg/config"
)

func Run(c *config.DDMConfig) int {

	d := &DDMObserver{DiskStat: []DiskStat{}}

	d.Run(c)
	// dict, parseErr := loadYAML(yaml)
	// ReadConfig()
	// c, err := config.ReadConf(yaml)
	// if err != nil {
	// 	fmt.Println("Error: ", err.Error())
	// 	return 1
	// }

	// log.Info(c)
	// log.Info(c.Common.Mount.Enabled)
	// fmt.Printf("v ==== %v \n", a)
	// log.Printf("%v\n", &c)
	// fmt.Printf("v ==== %v \n", c)

	// a := &A{a: 1, B: &B{b: 2}}

	// // using the Stringer interface
	// fmt.Printf("v ==== %v \n", a)
	// fmt.Printf("\n\n\n%+v\n-------------\n\n%#v\n\n\n\n", c, c)
	// b, err := json.Marshal(dict)
	// if err != nil {
	// 	return 1
	// }

	// log.Debug("Yaml: ", string(b))

	// if parseErr != nil {
	// 	fmt.Println("Error: ", parseErr.Error())
	// 	return 1
	// }

	return 0
}
