package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type MetricReqResp struct {
	Value int64 `json:"value"`
}

func (s *MetricsService) recordMetricHandler(w http.ResponseWriter, r *http.Request) {
	var req MetricReqResp
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	key := mux.Vars(r)["key"]
	if key == "" {
		http.Error(w, "{:key} parameter cannot be empty", http.StatusBadRequest)
		return
	}

	s.mCh <- Metric{
		key:     key,
		value:   req.Value,
		eventAt: time.Now(),
	}

	JSONResponse(w, http.StatusOK, struct{}{})
}

func (s *MetricsService) getMetricSumHandler(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	JSONResponse(
		w,
		http.StatusOK,
		MetricReqResp{
			Value: s.GetMetricSum(key),
		},
	)
}

func JSONResponse(w http.ResponseWriter, code int, output interface{}) {
	response, _ := json.Marshal(output)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
