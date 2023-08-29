package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/phelipperibeiro/desafio-multithreading/internal/entity"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
)

type APIResponse struct {
	URL      string
	Response *http.Response
	Duration time.Duration
	Error    error
}

func makeRequest(ctx context.Context, url string, ch chan<- APIResponse) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		ch <- APIResponse{Error: err}
		return
	}

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		ch <- APIResponse{Error: err}
		return
	}

	bodyBytes, _ := ioutil.ReadAll(response.Body)

	defer response.Body.Close()

	elapsed := time.Since(start)
	ch <- APIResponse{
		URL: url,
		Response: &http.Response{ // Criar uma nova resposta com o corpo copiado
			Status:     response.Status,
			StatusCode: response.StatusCode,
			Header:     response.Header,
			Body:       ioutil.NopCloser(bytes.NewReader(bodyBytes)),
		},
		Duration: elapsed,
		Error:    err,
	}
}

func CepHandler(responseWriter http.ResponseWriter, request *http.Request) {
	cep := chi.URLParam(request, "cep")
	fmt.Println("cep informado:", cep)

	api2URL := "http://viacep.com.br/ws/" + cep + "/json/"
	api1URL := "https://cdn.apicep.com/file/apicep/" + cep + ".json"

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channel := make(chan APIResponse, 2)

	go makeRequest(ctx, api2URL, channel)
	go makeRequest(ctx, api1URL, channel)

	var fasterResponse APIResponse
	select {
	case fasterResponse = <-channel:
		if fasterResponse.Response.StatusCode != 200 {
			fasterResponse = <-channel
		}
		fmt.Println("status code:", fasterResponse.Response.StatusCode)
	case <-ctx.Done():
		fmt.Println("Timeout occurred")
		return
	}

	res, err := io.ReadAll(fasterResponse.Response.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao ler resposta: %v\n", err)
	}

	//fmt.Println("API mais rápida (response):", string(res))
	fmt.Println("API mais rápida (provider):", fasterResponse.URL)
	fmt.Println("API mais rápida (time):", fasterResponse.Duration)

	var cepData entity.Cep
	err = json.Unmarshal(res, &cepData)
	cepData.Validate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao fazer parse da resposta: %v\n", err)
		http.Error(responseWriter, "Erro interno", http.StatusInternalServerError)
		return
	}

	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(http.StatusOK)
	json.NewEncoder(responseWriter).Encode(cepData)
}

func main() {
	fmt.Println("Iniciando servidor")
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/cep/{cep}", CepHandler)
	http.ListenAndServe(":8081", router)
}
