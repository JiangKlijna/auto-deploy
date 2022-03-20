package server

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"auto-deploy/config"
	"auto-deploy/res"

	"github.com/gorilla/websocket"
)

// AutoDeployServer Main Server
type AutoDeployServer struct {
	http.ServeMux
	Fs       http.FileSystem
	Server   config.ServerConfig
	Projects config.ProjectArray

	cacheTemp map[string]*template.Template

	WebsocketUpgrade *websocket.Upgrader
}

// Init AutoDeploy. register handlers
func (s *AutoDeployServer) Init(server config.ServerConfig, projects config.ProjectArray) {
	s.Server = server
	s.Projects = projects
	s.WebsocketUpgrade = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	s.Fs = res.FileSystem
	if res.IsCacheTemplate() {
		s.cacheTemp = make(map[string]*template.Template)
	}
	s.Handle(server.ContentPath+"/", LoggingHandler(ContentPathHandler(server.ContentPath, s.MainHandler)))
}

// Html output html by Fs
func (s *AutoDeployServer) Html(w http.ResponseWriter, name string, params map[string]interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	s.GetTemplate(name).Execute(w, params)
}

// GetTemplate 暂不考虑并发
func (s *AutoDeployServer) GetTemplate(name string) *template.Template {
	parse := func(name string) *template.Template {
		f, err := s.Fs.Open(name)
		if err != nil {
			panic(err)
		}
		readAll, err := ioutil.ReadAll(f)
		if err != nil {
			panic(err)
		}
		t, err := template.New(name).Parse(string(readAll))
		if err != nil {
			panic(err)
		}
		return t
	}
	if res.IsCacheTemplate() {
		if t, ok := s.cacheTemp[name]; ok {
			return t
		}
		t := parse("/" + name)
		s.cacheTemp[name] = t
		return t
	}
	return parse(name)
}

// Run AutoDeploy server
func (s *AutoDeployServer) Run(port string) {
	server := &http.Server{Addr: ":" + port, Handler: s}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalln(err.Error())
	}
}
