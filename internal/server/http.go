package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// 스토리지 구현 전 까지 In-memory logging을 위한 구조체
type httpServer struct {
	Log *Log
}

// In-memory logging server
func newHTTPServer() *httpServer {
	return &httpServer{Log: NewLog()}
}

// Real world에서 대세는 http+json이기 때문에 http로 먼저 구현
func NewHTTPServer(addr string) *http.Server {
	srv := newHTTPServer()
	r := mux.NewRouter()
	r.HandleFunc("/", srv.handleProduce).Methods("POST")
	r.HandleFunc("/", srv.handleConsume).Methods("GET")
	return &http.Server{Addr: addr, Handler: r}
}

type ProduceRequest struct {
	Record Record `json:"record"`
}

type ProduceResponse struct {
	Offset uint64 `json:"offset"`
}

type ConsumeRequest struct {
	Offset uint64 `json:"offset"`
}

type ConsumeResponse struct {
	Record Record `json:"record"`
}

// Produce 핸들러
// 요청을 deserialize하고, 로그에 append한 다음, offset을 serialize해서 응답
func (srv *httpServer) handleProduce(w http.ResponseWriter, r *http.Request) {
	var req ProduceRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	offset, err := srv.Log.Append(req.Record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := ProduceResponse{Offset: offset}
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Consume 핸들러
// 요청을 deserialize하고, 로그에서 읽은 다음, record를 serialize해서 응답
func (srv *httpServer) handleConsume(w http.ResponseWriter, r *http.Request) {
	var req ConsumeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	record, err := srv.Log.Read(req.Offset)
	// unsigned offset integer 범위가 넘어가는 에러 핸들링이 있었다.
	if err == ErrOffsetNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res := ConsumeResponse{Record: record}
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
