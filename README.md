# Orka Port Manager

This is a simple service to manage a pool of ports used for port forwarding in an orka enviornment

## Usage 

To build the server run `go build`.
To run the server, run the resulting executable with the following flags:
* `--ports`: Set this to the maximum number of vms that you expect to be using.
* (Optional) `--start`: Set this to the start of the port range you would like to use (Default: 9000)

## Endpoints

To check out a port use the `/checkout` endpoint.

To check a port back in use the `/checkin` endpoint. 
