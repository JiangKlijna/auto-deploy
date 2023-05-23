package config

import (
	"flag"
	"os"

	"gopkg.in/yaml.v2"
)

const Version = "1.0"
const ServerName = "ad-" + Version

var (
	Server   ServerConfig
	Projects ProjectArray
)

func init() {
	var (
		help, version, test bool
		filename            string
	)
	flag.BoolVar(&help, "h", false, "this help")
	flag.BoolVar(&version, "v", false, "show version and exit")
	flag.BoolVar(&test, "t", false, "test configuration and exit")
	flag.StringVar(&filename, "c", "config.yml", "set configuration file")

	flag.Parse()
	if help {
		printUsage()
		flag.PrintDefaults()
		os.Exit(1)
	} else if version {
		printVersion()
		os.Exit(1)
	}

	configFile, err := os.ReadFile(filename)
	if err != nil {
		println("read the configuration file", filename, "error:", err.Error())
		os.Exit(1)
	}
	var Config = new(config)
	err = yaml.Unmarshal(configFile, &Config)
	if err != nil {
		println("configuration file", filename, "syntax is bad. error:", err.Error())
		os.Exit(1)
	}
	if test {
		println("configuration file", filename, "is successful")
		os.Exit(1)
	}
	Server = Config.Server
	Projects = make([]Project, 0)
	for name, project := range Config.Projects {
		project.Name = name
		if project.Actions == nil {
			m := make(map[string]string)
			project.Actions = &m
		}
		if project.Dir != nil {
			(*project.Actions)["backup"] = "backup"
			(*project.Actions)["upload"] = "upload"
		}
		Projects = append(Projects, project)
	}
}

func printUsage() {
	println(`Usage: auto-deploy [-c .yml]
Example: auto-deploy -c config.yml

Options:`)
}

func printVersion() {
	println("auto-deploy version:", Version)
}
