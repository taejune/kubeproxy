package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
)

func main() {
	remote, err := url.Parse("https://C686027562BCA0DEC40D16A621BB7353.gr7.ap-northeast-2.eks.amazonaws.com")
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

	re := regexp.MustCompile("/cluster/([0-9a-zA-Z-_]*)/(.*)")
	proxy := &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			log.Printf("%+v", r.Out)
			r.SetURL(remote)
			r.Out.Header.Add("Authorization", "Bearer "+"eyJhbGciOiJSUzI1NiIsImtpZCI6IlA5UDUyajd4anJzQlZFNXVoU3h4M1BHajNBSzBxTWpaR2VlSjJMRTR1TEUifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJ6Y3Atc3lzdGVtIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6InpjcC1tY20tYmFja2VuZC1zZXJ2aWNlLWFkbWluIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQubmFtZSI6InpjcC1tY20tYmFja2VuZC1zZXJ2aWNlLWFkbWluIiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZXJ2aWNlLWFjY291bnQudWlkIjoiMTRjOWRiODMtNWUwOS00YjYxLWEwYjItYTY0ZmJhOTAwN2IxIiwic3ViIjoic3lzdGVtOnNlcnZpY2VhY2NvdW50OnpjcC1zeXN0ZW06emNwLW1jbS1iYWNrZW5kLXNlcnZpY2UtYWRtaW4ifQ.nKsSJAWAu-kDDxe1nBR9uk0WjpFipG8CI00MHwZFFSiuSLjhzzLNaxk2Ryy6FLqDeCkgaee6UC_MgYdIJjB1XQ1O0mUCgQkl5wEWCDEiXDx1S8p_xzm8op4dmYJ2FW83L5qicBzLG83VNI42FvXO1MiK5RtkOVOHc-6CEyqZSSWVAr-Dq32ehTCcvigs3ZRi-Sp61nSrHd9hMMqsMVNzeek3f6oDXV54RJeVxaP9Ezs_I8DFBCaDO3QRVQvOa-CaUUO9oyTAVOdBkl-EHqWNEMtqmT-rWt9UPJYIGwLQ-UkyNzOB10smBGtEKCaRaHdp1jjlfVF4temyJf8h_mxOXg")
		},
		Transport: transport,
	}

	handler := func(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			log.Printf("URL: %+v", r.URL)
			r.URL.Path = "/" + re.ReplaceAllString(r.RequestURI, "$2")
			p.ServeHTTP(w, r)
		}
	}

	http.HandleFunc("/", handler(proxy))
	err = http.ListenAndServe(":8443", nil)
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
