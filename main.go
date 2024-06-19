package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type BrasilAPI struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

type ViaCEP struct {
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

const BRASIL_API = "https://brasilapi.com.br/api/cep/v1/%s"
const VIA_CEP = "https://viacep.com.br/ws/%s/json"

func main() {
	cep := "01153000"

	brasilAPI := make(chan string)
	viaCEP := make(chan string)

	go func() {
		brasilAPI <- BodyToJson(BuscaCEP(fmt.Sprintf(BRASIL_API, cep)), BrasilAPI{})
	}()

	go func() {
		viaCEP <- BodyToJson(BuscaCEP(fmt.Sprintf(VIA_CEP, cep)), ViaCEP{})
	}()

	select {
	case retorno := <-brasilAPI:
		fmt.Println("Retorno da BrasilAPI: ")
		fmt.Println(retorno)
	case retorno := <-viaCEP:
		fmt.Println("Retorno da ViaCEP: ")
		fmt.Println(retorno)
	case <-time.After(time.Second):
		fmt.Println("timeout")
	}
}

func BuscaCEP(url string) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	origem := "ViaCEP"
	if url == BRASIL_API {
		origem = "BrasilAPI"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		fmt.Printf("Erro ao criar requisiçao %s: %v\n", origem, err)
		return nil
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Printf("Tempo excedido pra %s: %v\n", origem, err)
		} else {
			fmt.Printf("Erro ao enviar requisição %s: %v\n", origem, err)
		}
		return nil
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Printf("Erro na resposta ViaCEP: %s: %v\n", origem, err)
		return nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Erro ao ler corpo da resposta %s: %v\n", origem, err)
		return nil
	}

	return body
}

func BodyToJson(body []byte, v any) string {
	err := json.Unmarshal(body, &v)
	if err != nil {
		fmt.Println("Erro ao decodificar o JSON: ", err)
		return ""
	}

	jsonData, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		fmt.Println("Erro ao converter para JSON: ", err)
		return ""
	}

	return string(jsonData)
}
