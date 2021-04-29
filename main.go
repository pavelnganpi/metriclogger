package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	s := NewMetricsService()
	defer s.Shutdown()

	r := mux.NewRouter()
	r.HandleFunc("/metric/{key}", s.recordMetricHandler).Methods("POST")
	r.HandleFunc("/metric/{key}/sum", s.getMetricSumHandler).Methods("GET")

	log.Fatal(http.ListenAndServe("localhost:9093", r))
}