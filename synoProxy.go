package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

type SynoSession struct {
	Data struct {
		IsPortalPort bool   `json:"is_portal_port"`
		Sid          string `json:"sid"`
		Synotoken    string `json:"synotoken"`
	} `json:"data"`
	Success bool `json:"success"`
}

func main() {

	diskStation, ok := os.LookupEnv("DS")

	if !ok {
		log.Panicln("DS is not defined. Please add it as an environmental variable.")
	}

	remote, _ := url.Parse(diskStation)

	proxy := httputil.NewSingleHostReverseProxy(remote)
	http.HandleFunc("/", handler(proxy, diskStation))

	host, ok := os.LookupEnv("HOST")
	if !ok {
		host = ":6834"
	}

	log.Println("Starting Syno-Proxy on:", host)
	http.ListenAndServe(host, nil)
}

func handler(p *httputil.ReverseProxy, host string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL)
		apiRequest := strings.HasPrefix(r.URL.Path, "/webapi")

		if apiRequest {
			login, _ := url.Parse(host)
			query, _ := url.ParseQuery(login.RawQuery)

			user, userok := os.LookupEnv("USER")
			pass, passok := os.LookupEnv("PASS")

			if !userok {
				log.Panicln("USER is not defined. Please add it as an environmental variable.")
			}

			if !passok {
				log.Panicln("PASS is not defined. Please add it as an environmental variable.")
			}

			query.Add("api", "SYNO.API.Auth")
			query.Add("method", "Login")
			query.Add("version", "6")
			query.Add("account", user)
			query.Add("passwd", pass)
			query.Add("enable_syno_token", "yes")
			query.Add("format", "sid")
			query.Add("session", "SurveillanceStation")

			login.RawQuery = query.Encode()
			login.Path = "/webapi/auth.cgi"

			resp, err := http.Get(login.String())
			if err != nil {
				log.Panicln(err)
			}

			defer resp.Body.Close()

			var session SynoSession

			sessionDecoder := json.NewDecoder(resp.Body)
			err = sessionDecoder.Decode(&session)

			if err != nil {
				log.Panicln("Session decoding failed.", err)
			}

			query, _ = url.ParseQuery(r.URL.RawQuery)
			query.Add("_sid", session.Data.Sid)
			query.Add("SynoToken", session.Data.Synotoken)

			r.URL.RawQuery = query.Encode()

			p.ServeHTTP(w, r)
		}
	}
}
