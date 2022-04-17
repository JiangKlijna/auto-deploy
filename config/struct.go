package config

import (
	"time"

	"auto-deploy/lib"
)

type config struct {
	Server   ServerConfig
	Projects map[string]Project
}

type ServerConfig struct {
	Name        string
	Port        string
	Username    string
	Password    string
	Encoding    string
	ContentPath string `yaml:"content-path"`
	BackupPath  string `yaml:"backup-path"`
	UploadPath  string `yaml:"upload-path"`
}

type Project struct {
	Name        string
	Description string
	Dir         *string
	Actions     *map[string]string
}

func (p *Project) GetActionBySecretPath(secret, path string) *Action {
	for name, shell := range *p.Actions {
		if lib.HashCheck(name, secret, path) {
			return &Action{name, shell}
		}
	}
	return nil
}

func (p *Project) GetFileName(action, ext string) string {
	return ServerName + "-" + p.Name + "-" + time.Now().Format("2006-01-02-15-04-05") + "." + action + "." + ext
}

type ProjectArray []Project

func (ps ProjectArray) GetProjectBySecretPath(secret, path string) *Project {
	for _, project := range ps {
		if lib.HashCheck(project.Name, secret, path) {
			return &project
		}
	}
	return nil
}

type Action struct {
	Name  string
	Shell string
}
