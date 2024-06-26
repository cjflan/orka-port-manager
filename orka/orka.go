package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
)

const API_URL = "http://10.221.188.20" 

type OrkaClient struct {
    Client http.Client
    token string
}

func main() {
    email := flag.String("user", "support@macstadium.com", "username of your orka account")
    pass := flag.String("pass" , "", "password for your orka account")

    flag.Parse()

    if *pass == "" {
        panic("password needed")
    }

    orka := &OrkaClient{
        Client: http.Client{},
    }
    orka.getToken(*email, *pass)
    orka.CreateConfig()

    for i := 0; i < 10; i++ {
        go func() {
            port := orka.getPort()
            log.Println("checked out port ", port)

            vm := orka.DeployVM(port)

            reservedPorts := orka.ReservedPorts(vm.VMID)
            log.Printf("vmid: %s   port mapping: %v\n", vm.VMID, reservedPorts[3])

            time.Sleep(time.Second)

            orka.DeleteVM(vm.VMID)
            orka.ReturnPort(port)
        }()
        sleepTime := time.Duration(1 + rand.Intn(4)) * time.Second
        time.Sleep(sleepTime)
    }

}

type TokenResponse struct {
    Message string `json:"message"`
    Errors []any  `json:"errors"`
    Token  string `json:"token"`
}

type TokenRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

func (o *OrkaClient) getToken(email string, password string) {
    client := o.Client 
    url := fmt.Sprint(API_URL, "/token")

    reqBody := TokenRequest{
        Email: email,
        Password: password,
    }
    reqBytes, _ := json.Marshal(reqBody)
    payload := bytes.NewReader(reqBytes)

    req, err := http.NewRequest(http.MethodPost, url, payload)
    if err != nil {
        fmt.Println(err)
        return
    }

    req.Header.Add("Content-Type", "application/json")
    res, err := client.Do(req)
    if err != nil {
        fmt.Println(err)
        return
    }

    defer res.Body.Close()

    var tokenRes TokenResponse
    body, err := io.ReadAll(res.Body)
    json.Unmarshal(body, &tokenRes)

    bearerToken := fmt.Sprint("Bearer ", tokenRes.Token)
    o.token = bearerToken
}

type ConfigRequest struct {
    OrkaVMName    string `json:"orka_vm_name"`
    OrkaBaseImage string `json:"orka_base_image"`
    OrkaImage     string `json:"orka_image"`
    OrkaCPUCore   int    `json:"orka_cpu_core"`
    VcpuCount     int    `json:"vcpu_count"`
}

func (o *OrkaClient) CreateConfig() {
    client := o.Client

    vmConfig := ConfigRequest{
        OrkaVMName: "port-test",
        OrkaImage: "port-test",
        OrkaBaseImage: "90GBVenturaSSH.img",
        OrkaCPUCore: 6,
        VcpuCount: 6,
    }
    configBytes, _ := json.Marshal(vmConfig)
    request := bytes.NewReader(configBytes)

    url := fmt.Sprint(API_URL, "/resources/vm/create")
    req, err := http.NewRequest(http.MethodPost, url, request)

    if err != nil {
        fmt.Println(err)
        return
    }
    req.Header.Add("Content-Type", "application/json")
    req.Header.Add("Authorization", o.token)

    res, err := client.Do(req)
    if err != nil {
        fmt.Println(err)
        return
    }
    defer res.Body.Close()

    if res.StatusCode != http.StatusCreated {
        e := fmt.Sprintf("failed to create config: %s", res.Status)
        panic(e)
    }

}

type PortResponse struct {
    Port int `json:"port"`
    Message string `json:"message"`
}

func (o *OrkaClient) getPort() int {
    client := o.Client

    res, err := client.Get("http://127.0.0.1:8080/checkout")
    if err != nil {
        e := fmt.Sprint("failed to send request: ", err)
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

type CheckinRequest struct {
    Port int
}

func (o *OrkaClient) ReturnPort(port int) {
    client := o.Client

    checkinRequest := CheckinRequest{
        Port: port,
    }

    reqBytes, _ := json.Marshal(checkinRequest)
    req, _ := http.NewRequest(
        http.MethodPost,
        "http://127.0.0.1:8080/checkin",
        bytes.NewReader(reqBytes),
        )
    req.Header.Add("Content-Type", "application/json")
    res, err := client.Do(req)
    if err != nil {
        e := fmt.Sprint("Port return failed")
        panic(e)
    }

    defer res.Body.Close()

    if res.StatusCode != http.StatusOK {
        e := fmt.Sprint("failed to return port")
        panic(e)
    }

    log.Printf("port %d returned to queue", port)

}

type VMDeployRequest struct {
    OrkaVMName string `json:"orka_vm_name"`
    OrkaNodeName string `json:"orka_node_name"` 
    ReservedPorts []string `json:"reserved_ports"`
}
type StatusResponse struct {
    Message string `json:"message"`
    Errors                  []any `json:"errors"`
    VirtualMachineResources []struct {
        VirtualMachineName string `json:"virtual_machine_name"`
        VMDeploymentStatus string `json:"vm_deployment_status"`
        Status             []struct {
            ReservedPorts         []ReservedPorts `json:"reserved_ports"`
        } `json:"status"`
    } `json:"virtual_machine_resources"`
}

type ReservedPorts struct {
    HostPort  int    `json:"host_port"`
    GuestPort int    `json:"guest_port"`
    Protocol  string `json:"protocol"`
}

type DeployResponse struct {
    VMID            string `json:"vm_id"`
    Errors []any `json:"errors"`
}

func (o *OrkaClient) DeployVM(port int) DeployResponse {
    client := o.Client
    url := fmt.Sprint(API_URL, "/resources/vm/deploy")
    var ports []string
    ports = append(ports, fmt.Sprint("8080:", port))


    vmDeploy := VMDeployRequest{
        OrkaVMName: "port-test",
        ReservedPorts: ports,
    }

    reqBytes, _ := json.Marshal(vmDeploy)
    req, err := http.NewRequest(
        http.MethodPost,
        url, 
        bytes.NewReader(reqBytes),
        )

    if err != nil {
        panic(err)
    }
    req.Header.Add("Content-Type", "application/json")
    req.Header.Add("Authorization", o.token)

    res, err := client.Do(req)
    if err != nil {
        panic(err)
    }

    defer res.Body.Close()

    var response DeployResponse
    body, err := io.ReadAll(res.Body)
    json.Unmarshal(body, &response)

    if res.StatusCode != http.StatusOK {
        e := fmt.Sprintf("vm not created: %s, %v", res.Status, response.Errors)
        panic(e)
    }

    return response

}

func (o *OrkaClient) ReservedPorts(vmid string) []ReservedPorts {
    url := fmt.Sprint(API_URL, "/resources/vm/status/", vmid)
    client := o.Client
    req, err := http.NewRequest(http.MethodGet, url, nil)

    if err != nil {
        panic(err)
    }
    req.Header.Add("Authorization", o.token)

    res, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer res.Body.Close()

    var response StatusResponse
    body, err := io.ReadAll(res.Body)
    json.Unmarshal(body, &response)

    return response.VirtualMachineResources[0].Status[0].ReservedPorts
}

type VMDeleteRequest struct {
	OrkaVMName   string `json:"orka_vm_name"`
}

func (o *OrkaClient) DeleteVM(vmid string) {
    url := fmt.Sprint(API_URL, "/resources/vm/delete")

    client := o.Client
    vmDelete := VMDeleteRequest{
        OrkaVMName: vmid,
    }
    reqBytes, _ := json.Marshal(vmDelete)
    req, _ := http.NewRequest(
        http.MethodDelete,
        url,
        bytes.NewReader(reqBytes),
        ) 

    req.Header.Add("Content-Type", "application/json")
    req.Header.Add("Authorization", o.token)

    res, err := client.Do(req)
    if err != nil {
        e := fmt.Sprint("failed to delete vm: ", err)
        panic(e)
    }
    defer res.Body.Close()
    
    if res.StatusCode != http.StatusOK {
        e := fmt.Sprint("failed to delete vm: ", res.Status)
        panic(e)
    }


}
