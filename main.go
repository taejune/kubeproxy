package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

func main() {
	remote, err := url.Parse("")
	if err != nil {
		panic(err)
	}

	caCert, _ := os.ReadFile("cert.pem")
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}

	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(remote)
			r.Out.Header.Add("Authorization", "Bearer "+"")
		},
		Transport: transport,
	}

	handler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			p.ServeHTTP(w, r)
		}
	}

	http.HandleFunc("/", handler(proxy))
	go func() {
		err = http.ListenAndServe(":8443", nil)
		if err != nil {
			panic(err)
		}
	}()

	r := gin.Default()
	r.Any("/cluster/:cluster/*kubeAPI", func(c *gin.Context) {
		log.Println(parseHttpRequest(c.Request))
		//cluster := c.Param("cluster")
		c.Request.URL.Path = c.Param("kubeAPI")
		proxy.ServeHTTP(c.Writer, c.Request)
	})
	err = r.Run(":8080")
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
