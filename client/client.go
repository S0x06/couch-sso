package client

import (
	"../model"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

type Client struct {
	ProxyPort    string `json:"proxy_port"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	TokenHost    string `json:"token_host"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	GrantType    string `json:"grant_type"`
	Remote       string `json:"remote"`
}

var Token = &model.Token{}
var remote = &url.URL{}

func (c *Client) Run() {
	c.RetrieveToken()
	remote, _ = url.Parse(c.Remote)
	go c.TimerRefreshToken()
	http.HandleFunc("/", c.Proxy)
	addr := "0.0.0.0" + ":" + c.ProxyPort
	fmt.Printf("Listing on " + addr + "\n")
	log.Fatal(http.ListenAndServe(addr, nil))
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// Proxy Switch to the client execution agent.
func (c *Client) Proxy(w http.ResponseWriter, r *http.Request) {
	if ok := c.CheckAuth(r); ok {
		director := func(req *http.Request) {
			req.Header.Set("Authorization", "Bearer "+Token.AccessToken)
			req.URL.Scheme = remote.Scheme
			req.URL.Host = remote.Host
			req.Host = remote.Host
		}
		proxy := &httputil.ReverseProxy{Director: director}
		proxy.ServeHTTP(w, r)
	} else {
		fmt.Println("Authentication failed.", r)
		c.ProxyUnauthorized(w, r)
	}

}

// RetrieveToken Retrieve the token when the request occurs.
func (c *Client) RetrieveToken() (tok *model.Token, err error) {
	defer func() {
		if e := recover(); e != nil {
			log.Println(c, e)
			log.Println("Retry after 5 seconds. RetrieveToken()")
			time.Sleep(5 * time.Second)
			c.RetrieveToken()
		}
	}()

	data := url.Values{}
	data.Add("grant_type", c.GrantType)
	data.Add("client_id", c.ClientID)
	data.Add("client_secret", c.ClientSecret)
	data.Add("username", c.Username)
	data.Add("password", c.Password)

	return c.wwwFormRequest(data)
}

func (c *Client) RefreshToken() (tok *model.Token, err error) {

	if Token.RefreshToken == "" {
		return c.RetrieveToken()
	}

	data := url.Values{}
	data.Add("grant_type", "refresh_token")
	data.Add("client_id", c.ClientID)
	data.Add("client_secret", c.ClientSecret)
	data.Add("refresh_token", Token.RefreshToken)

	return c.wwwFormRequest(data)
}

func (c *Client) wwwFormRequest(data url.Values) (tok *model.Token, err error) {
	contentType := "application/x-www-form-urlencoded"
	resp, err := http.Post(c.TokenHost, contentType, bytes.NewBufferString(data.Encode()))
	checkErr(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	checkErr(err)
	err = json.Unmarshal(body, &Token)
	checkErr(err)

	return Token, err
}

// Periodic tasks
//func timer() {
//	ticker := time.Tick(500 * time.Millisecond)
//	<-ticker
//	go func() {
//		i := runtime.NumGoroutine()
//		fmt.Println("timer", i)
//		timer()
//	}()
//}
func (c *Client) TimerRefreshToken() {
	Task := Token.ExpiresIn - 180 // Refresh task time 180 seconds ahead
	ticker := time.Tick(time.Duration(Task) * time.Second)
	<-ticker
	go func() {
		Token, err := c.RefreshToken()
		if err != nil {
			fmt.Println("Try again in 10 seconds. ", err)
			time.Sleep(10 * time.Second)
		}
		fmt.Println("Token has been refreshed. ", Token)
		c.TimerRefreshToken()
	}()
}

func (c *Client) CheckAuth(r *http.Request) bool {
	reqUser, reqPass, ok := r.BasicAuth()
	if !ok || reqUser != c.Username || reqPass != c.Password {
		return false
	}
	return true
}

func (c *Client) ProxyUnauthorized(w http.ResponseWriter, r *http.Request) {
	type message struct {
		Error  string `json:"error"`
		Reason string `json:"reason"`
	}
	data := message{"unauthorized", "You are not authorized to access this db."}

	c.ResponseJson(data, w, r)
}

func (c *Client) ResponseJson(body interface{}, w http.ResponseWriter, r *http.Request) {
	data, _ := json.Marshal(body)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
