package main

import (
	"auto-deploy/config"
	"auto-deploy/server"
)

func main() {
	//lib.DownloadMdui()
	//lib.MakeGen()
	s := new(server.AutoDeployServer)
	s.Init(config.Server, config.Projects)
	s.Run(config.Server.Port)
}
