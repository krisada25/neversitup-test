package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type LogReq struct {
	Timestamp     time.Time `json:"timestamp"`
	CustomerID    string    `json:"customerID"`
	TransactionID string    `json:"transactionID"`
	Service       string    `json:"service"`
	Endpoint      string    `json:"endpoint"`
	HttpCode      string    `json:"httpCode"`
	Request       string    `json:"request"`
	Response      string    `json:"response"`
}

func Log(customerID string, transactionID string, service string, endpoint string, httpCode string, request string, response string) error {

	url := os.Getenv("URL_LOG_SERVICE")
	method := "POST"

	reqPayload := &LogReq{
		Timestamp:     time.Now().UTC(),
		CustomerID:    customerID,
		TransactionID: transactionID,
		Service:       service,
		Endpoint:      endpoint,
		HttpCode:      httpCode,
		Request:       request,
		Response:      response,
	}
	payload, err := json.Marshal(reqPayload)
	if err != nil {
		fmt.Println(err)
		return err
	}
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewReader(payload))

	if err != nil {
		fmt.Println(err)
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()

	// body, err := io.ReadAll(res.Body)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }
	// fmt.Println(string(body))
	return nil
}
