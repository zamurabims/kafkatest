package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"wish1/internal/metrics"
	"wish1/internal/stoplist"
	"wish1/internal/window"
)

type Handler struct {
	win      *window.Window
	stoplist *stoplist.StopList
}

func NewHandler(win *window.Window, sl *stoplist.StopList) *Handler {
	return &Handler{win: win, stoplist: sl}
}

func (h *Handler) GetTop(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		metrics.TopRequestDuration.Observe(time.Since(start).Seconds())
	}()

	limit := 10
	if s := r.URL.Query().Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			limit = n
		}
	}

	entries := h.win.Top(limit, h.stoplist.Contains)
	writeJSON(w, http.StatusOK, entries)
}

func (h *Handler) AddStopWord(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Word string `json:"word"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Word) == "" {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	h.stoplist.Add(body.Word)
	metrics.StopListSize.Inc()
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RemoveStopWord(w http.ResponseWriter, r *http.Request) {
	word := strings.TrimPrefix(r.URL.Path, "/stoplist/")
	if word == "" {
		http.Error(w, "word required", http.StatusBadRequest)
		return
	}
	h.stoplist.Remove(word)
	metrics.StopListSize.Dec()
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetStopList(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.stoplist.List())
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
