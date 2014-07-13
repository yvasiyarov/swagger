package main

import (
	"flag"
	//	"fmt"
	"log"
	"net/http"
	"strings"
)

var host = flag.String("host", "127.0.0.1", "Host")
var port = flag.String("port", "8080", "Port")

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	isJsonRequest := false

	if acceptHeaders, ok := r.Header["Accept"]; ok {
		for _, acceptHeader := range acceptHeaders {
			if strings.Contains(acceptHeader, "json") {
				isJsonRequest = true
				break
			}
		}
	}

	if isJsonRequest {
		w.Write([]byte(resourceListingJson))
	} else {
		http.Redirect(w, r, "/swagger-ui/", http.StatusFound)
	}
}

func main() {
	flag.Parse()

	// To serve a directory on disk (/tmp) under an alternate URL
	// path (/tmpfiles/), use StripPrefix to modify the request
	// URL's path before the FileServer sees it:
	http.HandleFunc("/", IndexHandler)
	http.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.Dir("./swagger-ui"))))

	for apiKey, apiJson := range apiDescriptionsJson {
		http.HandleFunc("/"+apiKey+"/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(apiJson))
		})
	}

	listenTo := *host + ":" + *port
	log.Printf("Star listen to %s", listenTo)

	http.ListenAndServe(listenTo, http.DefaultServeMux)
	//http.ListenAndServe(":8080", http.StripPrefix("/swagger-ui/", http.FileServer(http.Dir("./swagger-ui")) )
}
