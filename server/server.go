package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Server struct {
	Port   string `json:"port"`
	User   string `json:"user"`
	Pass   string `json:"pass"`
	Target string `json:"target"`
}

var target = &url.URL{}

func (s *Server) Run() {
	target, _ = url.Parse(s.Target)
	http.HandleFunc("/", s.Proxy)
	addr := "0.0.0.0"
	fmt.Printf("Listing on " + addr + ":" + s.Port + "\n")
	log.Fatal(http.ListenAndServe(addr+":"+s.Port, nil))
}

func (s *Server) Proxy(w http.ResponseWriter, r *http.Request) {
	if s.IsRootPath(r.RequestURI) {
		s.ProxyRootPath(w, r)
	} else {
		err := s.CheckJWT(w, r)
		if err != nil {
			s.ProxyUnauthorized(w, r)
		} else {
			s.ProxyServer(w, r)
		}
	}
}

// ProxyServer Switch to the server execution agent.
func (s *Server) ProxyServer(w http.ResponseWriter, r *http.Request) {
	director := func(req *http.Request) {
		req.SetBasicAuth(s.User, s.Pass)
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
	}
	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(w, r)
}

func (s *Server) ProxyUnauthorized(w http.ResponseWriter, r *http.Request) {
	type message struct {
		Error  string `json:"error"`
		Reason string `json:"reason"`
	}
	unAuth := message{"unauthorized", "You are not authorized to access this db. Please check the client request."}
	s.ResponseJson(unAuth, w, r)
}

func (s *Server) ProxyRootPath(w http.ResponseWriter, r *http.Request) {
	type message struct {
		Msg  string `json:"msg"`
		Name string `json:"name"`
	}
	unAuth := message{"Welcome", "The Couch-SSO service."}
	s.ResponseJson(unAuth, w, r)
}

func (s *Server) ResponseJson(body interface{}, w http.ResponseWriter, r *http.Request) {
	data, _ := json.Marshal(body)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s *Server) CheckJWT(w http.ResponseWriter, r *http.Request) error {
	token, err := s.GetJWTFormHeader(r)
	if err != nil {
		return err
	}
	err = VerifyToken(token)
	if err != nil {
		return err
	}
	_, err = DecodeJWT(token)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) GetJWTFormHeader(r *http.Request) (str string, err error) {
	auth := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return "", errors.New("Not found token in header. ")
	}
	str = strings.Replace(auth, prefix, "", -1)

	return str, nil
}

func (s *Server) IsRootPath(path string) bool {
	if path == "/" || path == "" {
		return true
	}
	return false
}
