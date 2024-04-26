package orka

import (
    "bytes"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "net/http"
    "strings"
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
    orka.createConfig()

    port := orka.getPort()

    vm := orka.DeployVM(port)
    reservedPorts := orka.ReservedPorts(vm.VMID)

    fmt.Printf("reservedPorts: %v\n", reservedPorts)
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

func (o *OrkaClient) createConfig() {
    client := o.Client
    payload := strings.NewReader(`{
        "orka_vm_name": "port-test",
        "orka_base_image": "90gbsonomassh.img",
        "orka_image": "port-test",
        "orka_cpu_core": 3,
        "vcpu_count": 3,
        }`)

    url := fmt.Sprint(API_URL, "/resources/vm/create")
    req, err := http.NewRequest(http.MethodPost, url, payload)

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
        fmt.Println("failed to create config: ", res.Status)
    }

}

type PortResponse struct {
    Port int `json:"port"`
    Message string `json:"message"`
}

func (o *OrkaClient) getPort() int {
    client := o.Client

    res, err := client.Get("127.0.0.1:8080/checkout")
    if err != nil {
        panic("failed to send request") 
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

type VMDeployRequest struct {
    OrkaVMName string `json:"orka_vm_name"`
    OrkaNodeName string `json:"orka_node_name"` 
    ReservedPorts string `json:"reserved_ports"`
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
}

func (o *OrkaClient) DeployVM(port int) DeployResponse {
    client := o.Client
    url := fmt.Sprint(API_URL, "/resources/vm/deploy")

    ports := fmt.Sprint("8080:", port)

    vmDeploy := VMDeployRequest{
        OrkaVMName: "port-test",
        ReservedPorts: ports,
    }

    reqBytes, _ := json.Marshal(vmDeploy)
    req, err := http.NewRequest(
        http.MethodPost,
        url, bytes.NewReader(reqBytes),
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
