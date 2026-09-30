package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

func main() {
	port := flag.Int("port", 8080, "port to listen on")
	flag.Parse()
	addr := fmt.Sprintf(":%d", *port)
	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("=== %s ===\n", time.Now().Format(time.RFC3339))
		fmt.Printf("%s %s %s\n", r.Method, r.URL.Path, r.Proto)
		fmt.Printf("URL: %s\n", r.URL.String())
		fmt.Println("Query:")
		query := r.URL.Query()
		if len(query) == 0 {
			fmt.Println("  <none>")
		} else {
			for name, values := range query {
				for _, v := range values {
					fmt.Printf("  %s: %s\n", name, v)
				}
			}
		}
		fmt.Println("Headers:")
		for name, values := range r.Header {
			for _, v := range values {
				fmt.Printf("  %s: %s\n", name, v)
			}
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println("Body: <error reading:", err, ">")
		} else {
			if len(body) == 0 {
				fmt.Println("Body: <empty>")
			} else {
				fmt.Printf("Body:\n%s\n", body)
			}
		}
		fmt.Println()

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}

	http.HandleFunc("/mirror/", mirrorHandler)
	http.HandleFunc("/", handler)

	log.Printf("Mock server listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// mirrorHandler echoes back the request's headers and body. The optional
// "status" query param overrides the response status code (default 200).
func mirrorHandler(w http.ResponseWriter, r *http.Request) {
	status := http.StatusOK
	if s := r.URL.Query().Get("status"); s != "" {
		parsed, err := strconv.Atoi(s)
		if err != nil || parsed < 100 || parsed > 599 {
			http.Error(w, "invalid status: must be an integer between 100 and 599", http.StatusBadRequest)
			return
		}
		status = parsed
	}

	fmt.Printf("=== %s %s ===\n", time.Now().Format(time.RFC3339), r.URL.String())

	for name, values := range r.Header {
		for _, v := range values {
			w.Header().Add(name, v)
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error reading body: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		log.Printf("mirror: failed to write response body: %v", err)
	}
}
