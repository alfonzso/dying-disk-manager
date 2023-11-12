package ddm

import (
	"fmt"

	"github.com/alfonzso/dying-disk-manager/pkg/config"
	log "github.com/sirupsen/logrus"
)

// func buildConfigDetails(source map[string]interface{}, env map[string]string) composeType.ConfigDetails {
// 	workingDir, err := os.Getwd()
// 	if err != nil {
// 		panic(err)
// 	}

// 	return composeType.ConfigDetails{
// 		WorkingDir: workingDir,
// 		ConfigFiles: []composeType.ConfigFile{
// 			{Filename: "filename.yml", Config: source},
// 		},
// 		Environment: env,
// 	}
// }

// func loadYAML(yaml string) (*composeType.Config, error) {
// 	return loadYAMLWithEnv(yaml, nil)
// }

// func loadYAMLWithEnv(yaml string, env map[string]string) (*composeType.Config, error) {
// 	dict, err := loader.ParseYAML([]byte(yaml))
// 	if err != nil {
// 		return nil, err
// 	}

// 	return loader.Load(buildConfigDetails(dict, env))
// }


type A struct {
	a int32
	B *B
}

type B struct{ b int32 }

func Run(c *config.DDMConfig) int {
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
	log.Printf("%v\n", &c)
	fmt.Printf("v ==== %v \n", c)

	a := &A{a: 1, B: &B{b: 2}}

	// using the Stringer interface
	fmt.Printf("v ==== %v \n", a)
	fmt.Printf("\n\n\n%+v\n-------------\n\n%#v\n\n\n\n", c, c)
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
