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
	Provider string
	Response *http.Response
	Duration time.Duration
	Error    error
}

type brasilapi struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

type viacep struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

// Função para mapear brasilapi para respostaCepPadrão
func mapFromBrasilapi(b brasilapi) entity.Cep {
	return entity.Cep{
		Cep:          b.Cep,
		State:        b.State,
		City:         b.City,
		Neighborhood: b.Neighborhood,
		Street:       b.Street,
	}
}

// Função para mapear viacep para respostaCepPadrão
func mapFromViacep(v viacep) entity.Cep {
	return entity.Cep{
		Cep:          v.Cep,
		State:        v.Uf,
		City:         v.Localidade,
		Neighborhood: v.Bairro,
		Street:       v.Logradouro,
	}
}

func makeRequest(ctx context.Context, url string, provider string, ch chan<- APIResponse) {
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
		URL:      url,
		Provider: provider,
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

	api2URL := "http://viacep.com.br/ws/" + cep + "/json/"  // response -> brasilapi
	api1URL := "https://brasilapi.com.br/api/cep/v1/" + cep // response -> viacep

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	channel := make(chan APIResponse, 2)

	go makeRequest(ctx, api2URL, "viacep", channel)
	go makeRequest(ctx, api1URL, "brasilapi", channel)

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

	if "viacep" == fasterResponse.Provider {
		var viacep viacep
		if err := json.Unmarshal(res, &viacep); err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao fazer parse da resposta: %v\n", err)
			http.Error(responseWriter, "Erro interno", http.StatusInternalServerError)
			return
		}
		cepData = mapFromViacep(viacep)
	}

	if "brasilapi" == fasterResponse.Provider {
		var brasilapi brasilapi
		if err := json.Unmarshal(res, &brasilapi); err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao fazer parse da resposta: %v\n", err)
			http.Error(responseWriter, "Erro interno", http.StatusInternalServerError)
			return
		}
		cepData = mapFromBrasilapi(brasilapi)
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
