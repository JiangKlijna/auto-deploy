package server

import (
	"archive/zip"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"auto-deploy/config"
	"auto-deploy/lib"
	"auto-deploy/shell"

	"github.com/gorilla/websocket"
)

type Action string

const (
	Login    Action = "login"
	Shutdown Action = "shutdown"
)

const CookieName = "KEY"

// MainHandler Login verification
func (s *AutoDeployServer) MainHandler(w http.ResponseWriter, r *http.Request) {
	isGet := r.Method == http.MethodGet
	isPost := r.Method == http.MethodPost
	if !isGet && !isPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// get and static
	if isGet && strings.HasPrefix(r.RequestURI, "/static/") {
		s.StaticFileHandler(w, r)
		return
	}
	clientIP := r.RemoteAddr[0:strings.LastIndex(r.RemoteAddr, ":")]
	secret, token, cookie := lib.GenerateAll(s.Server.Username, s.Server.Password, clientIP, r.UserAgent())
	r.Header.Set("secret", secret)
	r.Header.Set("token", token)
	r.Header.Set("value", cookie)

	c, _ := r.Cookie(CookieName)
	isLogin := c != nil && c.Value == cookie
	if isPost {
		// post and actions
		s.PostHandler(w, r, isLogin)
	} else {
		// get and [template].html & websocket
		s.GetHandler(w, r, isLogin)
	}
}

// StaticFileHandler get and output static file
func (s *AutoDeployServer) StaticFileHandler(w http.ResponseWriter, r *http.Request) {
	txt := mime.TypeByExtension(path.Ext(r.RequestURI))
	w.Header().Set("Content-Type", txt)

	f, err := s.Fs.Open(r.RequestURI)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer f.Close()
	_, err = io.Copy(w, f)
	if err != nil {
		w.WriteHeader(http.StatusGone)
	}
}

// PostHandler post and handler actions
func (s *AutoDeployServer) PostHandler(w http.ResponseWriter, r *http.Request, isLogin bool) {
	w.Header().Set("Content-Type", "text/json; charset=utf-8")
	secret := r.Header.Get("secret")
	action := r.Header.Get("action")
	cookie := r.Header.Get("value")
	if lib.HashCheck(string(Login), secret, action) {
		if r.Header.Get("token") == r.FormValue("token") { // login success & set cookie
			http.SetCookie(w, &http.Cookie{Name: CookieName, Value: cookie, HttpOnly: true, Path: "/"})
			w.Write([]byte(`{"code":0,"msg":"login success!"}`))
		} else {
			http.SetCookie(w, &http.Cookie{Name: CookieName, HttpOnly: true, Path: "/", MaxAge: -1})
			w.Write([]byte(`{"code":1,"msg":"login failure!"}`))
		}
		return
	}
	if lib.HashCheck(string(Shutdown), secret, action) {
		if isLogin {
			os.Exit(0)
		}
	}
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"code":1,"msg":"action failure!"}`))
}

// GetHandler index.html or project.html or websocket or upload
func (s *AutoDeployServer) GetHandler(w http.ResponseWriter, r *http.Request, isLogin bool) {
	if !isLogin {
		s.Html(w, "login.html", map[string]interface{}{
			"server": s.Server,
			"secret": r.Header.Get("secret"),
		})
		return
	}
	if r.RequestURI == "/" {
		// /
		s.Html(w, "index.html", map[string]interface{}{
			"projects": s.Projects,
			"server":   s.Server,
			"secret":   r.Header.Get("secret"),
		})
		return
	}
	arr := strings.Split(r.RequestURI, "/")
	if len(arr) > 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	project := s.Projects.GetProjectBySecretPath(r.Header.Get("secret"), arr[1])
	if len(arr) == 2 { // /[project]
		s.Html(w, "project.html", map[string]interface{}{
			"isProject": project != nil,
			"project":   project,
			"server":    s.Server,
			"secret":    r.Header.Get("secret"),
		})
	} else { // /[project]/[action]
		action := project.GetActionBySecretPath(r.Header.Get("secret"), arr[2])
		if action == nil {
			w.WriteHeader(http.StatusNotImplemented)
			return
		}
		// websocket upgrade
		conn, err := s.WebsocketUpgrade.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer conn.Close()
		WebsocketHandler(conn, &s.Server, project, action)
	}
}

// WebsocketHandler call action
func WebsocketHandler(conn *websocket.Conn, server *config.ServerConfig, project *config.Project, action *config.Action) {
	if action.Name == "upload" {
		var r io.Reader
		_, r, err := conn.NextReader()
		if err != nil {
			println(err.Error())
			return
		}
		uploadFile := path.Join(server.UploadPath, project.GetFileName("upload", "zip"))
		err = lib.OutputFile(r, uploadFile)
		if err != nil {
			println(err.Error())
			return
		}
		zipReader, err := zip.OpenReader(uploadFile)
		if err != nil {
			println(err.Error())
			return
		}
		defer zipReader.Close()
		err = lib.DeleteDirElseSelf(*project.Dir)
		if err != nil {
			conn.WriteMessage(websocket.BinaryMessage, []byte("delete project dir error "+err.Error()))
			return
		}
		conn.WriteMessage(websocket.BinaryMessage, []byte("delete project dir success dir is "+*project.Dir+"\n"))

		err = lib.Unzip(zipReader, *project.Dir)
		if err != nil {
			conn.WriteMessage(websocket.BinaryMessage, []byte("unzip upload file error "+err.Error()))
			return
		}
		conn.WriteMessage(websocket.BinaryMessage, []byte("unzip upload file success dir is "+*project.Dir+"\n"))
		return
	} else if action.Name == "backup" {
		lib.Zip(*project.Dir, path.Join(server.BackupPath, project.GetFileName("backup", "zip")), func(s string) {
			conn.WriteMessage(websocket.BinaryMessage, []byte(s))
		})
		return
	}

	cmd := shell.ExecShell(action.Shell)
	stdOut, err := cmd.StdoutPipe()
	if err != nil {
		conn.WriteMessage(websocket.BinaryMessage, []byte(err.Error()))
		return
	}
	err = cmd.Start()
	if err != nil {
		conn.WriteMessage(websocket.BinaryMessage, []byte(err.Error()))
		return
	}

	buf := make([]byte, 4096)
	for {
		n, err := stdOut.Read(buf)
		//time.Sleep(time.Second)
		if err != nil {
			conn.WriteMessage(websocket.BinaryMessage, []byte("Error ReadPtyAndWriteSkt cmd read failed: "+err.Error()))
			return
		}
		conn.WriteMessage(websocket.BinaryMessage, buf[:n])
	}
}

// ContentPathHandler content path prefix
func ContentPathHandler(contentpath string, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, contentpath)
		r.URL.Path = p
		r.RequestURI = p
		next.ServeHTTP(w, r)
	})
}

// LoggingHandler Log print
func LoggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		str := fmt.Sprintf(
			"%s [%s] <%s> in (%v) from {%s}",
			start.Format("2006/01/02 15:04:05"),
			r.Method,
			r.URL.Path,
			time.Since(start),
			r.RemoteAddr)
		println(str)
	})
}
