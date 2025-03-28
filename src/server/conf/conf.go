package conf

import (
	"os"
	"path/filepath"
	"search-indexer/running"
	fsutils "search-indexer/utils/fs"

	"gopkg.in/yaml.v3"
)

type Exclude struct {
	UseGitIgnore bool     `yaml:"use_git_ignore"`
	Customized   []string `yaml:"customized"` // Won't be used if enable_git_ignore is true
}

type Filters struct {
	Exclude Exclude  `yaml:"exclude"`
	Include []string `yaml:"include"`
}

type Conf struct {
	ForTest struct {
		Path string `yaml:"path"`
	} `yaml:"for_test"`

	Filters Filters `yaml:"filters"`
	Port    int     `yaml:"port"`
}

var conf *Conf

func checkMode() {
	if !running.IsServerMode() {
		panic("server conf is not accessible in client mode!")
	}
}

func Get() *Conf {
	checkMode()
	return conf
}

var serverConf *string

func Load() error {
	checkMode()

	search := []string{
		"./server.local.yaml",
		"./server.yaml",
		filepath.Join(running.RootPath(), "server.yaml"),
	}

	for _, path := range search {
		serverConf = &path
		if _, err := os.Stat(path); err == nil {
			break
		}
	}

	conf = &Conf{
		Port: running.DefaultListenPort(),
	}

	confBytes := fsutils.ReadFileWithDefault(*serverConf, []byte(``))
	if err := yaml.Unmarshal(confBytes, conf); err != nil {
		return err
	}

	return nil
}
