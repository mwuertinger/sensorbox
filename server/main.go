package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/mwuertinger/sensorbox/server/pb"
)

func main() {
	certFile := ""
	keyFile := ""

	mux := http.NewServeMux()
	mux.HandleFunc("/sensorbox", handler)

	s := &http.Server{
		Addr:           ":8080",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServeTLS(certFile, keyFile))
}

func handler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("read body: %v", err)
		w.WriteHeader(500)
		return
	}
	var request pb.Request
	if err := proto.Unmarshal(buf, &request); err != nil {
		log.Printf("unmarshal: %v", err)
		w.WriteHeader(500)
		return
	}

	// process request

	response := pb.Response{}
	data, err := proto.Marshal(&response)
	if err != nil {
		log.Printf("marshal: %v", err)
		w.WriteHeader(500)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		log.Printf("write response: %v", err)
		w.WriteHeader(500)
		return
	}
}
