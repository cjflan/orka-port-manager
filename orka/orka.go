package orka

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"
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
    Help    struct {
    } `json:"help"`
    Errors                  []any `json:"errors"`
    VirtualMachineResources []struct {
        VirtualMachineName string `json:"virtual_machine_name"`
        VMDeploymentStatus string `json:"vm_deployment_status"`
        Status             []struct {
            Owner                 string `json:"owner"`
            VirtualMachineName    string `json:"virtual_machine_name"`
            VirtualMachineID      string `json:"virtual_machine_id"`
            NodeLocation          string `json:"node_location"`
            NodeStatus            string `json:"node_status"`
            VirtualMachineIP      string `json:"virtual_machine_ip"`
            VncPort               string `json:"vnc_port"`
            ScreenSharingPort     string `json:"screen_sharing_port"`
            SSHPort               string `json:"ssh_port"`
            CPU                   int    `json:"cpu"`
            Vcpu                  int    `json:"vcpu"`
            Gpu                   string `json:"gpu"`
            RAM                   string `json:"RAM"`
            BaseImage             string `json:"base_image"`
            Image                 string `json:"image"`
            ConfigurationTemplate string `json:"configuration_template"`
            VMStatus              string `json:"vm_status"`
            IoBoost               bool   `json:"io_boost"`
            NetBoost              bool   `json:"net_boost"`
            UseSavedState         bool   `json:"use_saved_state"`
            ReservedPorts         []ReservedPorts `json:"reserved_ports"`
            CreationTimestamp time.Time `json:"creationTimestamp"`
            Tag               string    `json:"tag"`
            TagRequired       bool      `json:"tag_required"`
        } `json:"status"`
    } `json:"virtual_machine_resources"`
}

type ReservedPorts struct {
    HostPort  int    `json:"host_port"`
    GuestPort int    `json:"guest_port"`
    Protocol  string `json:"protocol"`
}

type DeployResponse struct {
    Message string `json:"message"`
    Help    struct {
        StartVirtualMachine            string `json:"start_virtual_machine"`
        StopVirtualMachine             string `json:"stop_virtual_machine"`
        ResumeVirtualMachine           string `json:"resume_virtual_machine"`
        SuspendVirtualMachine          string `json:"suspend_virtual_machine"`
        DataForVirtualMachineExecTasks struct {
            OrkaVMName string `json:"orka_vm_name"`
        } `json:"data_for_virtual_machine_exec_tasks"`
        VirtualMachineVnc string `json:"virtual_machine_vnc"`
    } `json:"help"`
    Errors          []any  `json:"errors"`
    RAM             string `json:"ram"`
    Vcpu            string `json:"vcpu"`
    HostCPU         string `json:"host_cpu"`
    IP              string `json:"ip"`
    SSHPort         string `json:"ssh_port"`
    ScreenSharePort string `json:"screen_share_port"`
    VMID            string `json:"vm_id"`
    PortWarnings    []any  `json:"port_warnings"`
    IoBoost         bool   `json:"io_boost"`
    NetBoost        bool   `json:"net_boost"`
    UseSavedState   bool   `json:"use_saved_state"`
    GpuPassthrough  bool   `json:"gpu_passthrough"`
    VncPort         string `json:"vnc_port"`
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
    url := fmt.Sprint(API_URL, "/resources/vm/status/myorkavm")
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
