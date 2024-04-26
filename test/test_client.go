package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type TestClient struct {
    client *http.Client
}

func main() {
    tc := TestClient{
        client: &http.Client{},
    }

    for i := 0; i < 50; i++ {
        port := tc.getPort()
        fmt.Println(port)
        tc.returnPort(port)
    }
}

type PortCheckin struct {
    Port int `json:"port"`
}

type CheckinResponse struct {
    Message string `json:"message"`
}

func (tc *TestClient) returnPort(p int) {
    client := tc.client    

    port := PortCheckin{
        Port: p,
    }
    portBytes, _ := json.Marshal(port)
    
    req, _ := http.NewRequest(
        http.MethodPut,
        "http://127.0.0.1:8080/checkin",
        bytes.NewReader(portBytes),
        )

    res, err := client.Do(req)
    if err != nil {
        e := fmt.Sprintf("request failed: %s", err)
        panic(e)
    }

    defer res.Body.Close()

    var response CheckinResponse
    body, _ := io.ReadAll(res.Body)
    json.Unmarshal(body, &response)

    fmt.Println(response.Message)
}

type PortResponse struct {
    Port    int    `json:"port"`
    Message string `json:"message"`
}

func (tc *TestClient) getPort() int {
    client := tc.client
    res, err := client.Get("http://127.0.0.1:8080/checkout")
    if err != nil {
        e := fmt.Sprintf("failed to send message: %s", err)
        panic(e)
    }

    defer res.Body.Close()

    var response PortResponse
    body, _ := io.ReadAll(res.Body)
    err = json.Unmarshal(body, &response)
    if err != nil {
        panic(err)
    }
    return response.Port
}
