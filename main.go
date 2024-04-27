package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type apiConfig struct {
	fileserverHits int
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	apiCfg := apiConfig{
		fileserverHits: 0,
	}
	mux := http.NewServeMux()
	mux.Handle("/app/*", apiCfg.middlewareMetricInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", healthHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("GET /api/reset", apiCfg.resetHandler)
	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)

	corsMux := middlewareCors(mux)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type chirpBody struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	chirp := chirpBody{}
	err := decoder.Decode(&chirp)
	if err != nil {
		log.Printf("Error decoding chirp body: %s", err)
		w.WriteHeader(500)
		return
	}

	type invalidChirp struct {
		Error string `json:"error"`
	}

	if len(chirp.Body) >= 140 {
		resp := invalidChirp{
			Error: "chirp is too long",
		}
		dat, err := json.Marshal(resp)
		if err != nil {
			log.Printf("Error marshalling JSON: %s", err)
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(400)
		w.Write(dat)
		return
	}

	type validChirp struct {
		Valid bool `json:"valid"`
	}

	respBody := validChirp{
		Valid: true,
	}

	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)
}
