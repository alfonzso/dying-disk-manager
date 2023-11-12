package ddm

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/docker/cli/cli/compose/loader"
	composeType "github.com/docker/cli/cli/compose/types"
	log "github.com/sirupsen/logrus"
)

func buildConfigDetails(source map[string]interface{}, env map[string]string) composeType.ConfigDetails {
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return composeType.ConfigDetails{
		WorkingDir: workingDir,
		ConfigFiles: []composeType.ConfigFile{
			{Filename: "filename.yml", Config: source},
		},
		Environment: env,
	}
}

func loadYAML(yaml string) (*composeType.Config, error) {
	return loadYAMLWithEnv(yaml, nil)
}

func loadYAMLWithEnv(yaml string, env map[string]string) (*composeType.Config, error) {
	dict, err := loader.ParseYAML([]byte(yaml))
	if err != nil {
		return nil, err
	}

	return loader.Load(buildConfigDetails(dict, env))
}

func Run(yaml string) int {
	dict, parseErr := loadYAML(yaml)

	b, err := json.Marshal(dict)
	if err != nil {
		return 1
	}

	log.Debug("Yaml: ", string(b))

	if parseErr != nil {
		fmt.Println("Error: ", parseErr.Error())
		return 1
	}

	return 0
}
