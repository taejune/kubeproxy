package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func main() {
	remote, err := url.Parse("https://C686027562BCA0DEC40D16A621BB7353.gr7.ap-northeast-2.eks.amazonaws.com")
	if err != nil {
		panic(err)
	}

	log.Println("hello")

	handler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {

			log.Println(parseHttpRequest(r))
			p.ServeHTTP(w, r)
		}
	}

	//pattern := `[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}`
	//re, err := regexp.Compile(pattern)
	//if err != nil {
	//	fmt.Printf("There is a problem with your regexp.\n")
	//	return
	//}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(remote)
			//r.Out.RequestURI = r.In.RequestURI
			r.Out.Host = r.In.Host // if desired
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	http.HandleFunc("/", handler(proxy))
	err = http.ListenAndServeTLS(":8443", "server.crt", "server.key", nil)
	if err != nil {
		panic(err)
	}
}

func parseHttpRequest(r *http.Request) string {
	base := fmt.Sprintf("%s %s %s %s %s %d",
		r.RemoteAddr, r.Proto, r.Method, r.Host, r.RequestURI, r.ContentLength)

	var headerFields []string
	for key, values := range r.Header {
		field := fmt.Sprintf("%s: %s", key, strings.Join(values, ", "))
		headerFields = append(headerFields, field)
	}
	header := strings.Join(headerFields, "\n")

	defer r.Body.Close()
	dat, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("failed to read body")
	}
	body, err := json.MarshalIndent(string(dat), "", " ")
	if err != nil {
		log.Println("failed to marshal body")
	}

	return fmt.Sprintf("%s\n%s\n%s\n", base, header, body)
}
