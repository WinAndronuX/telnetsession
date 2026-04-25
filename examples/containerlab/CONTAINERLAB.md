# Testing with Containerlab

This guide explains how to test the `telnetsession` library using a virtual network environment provided by [Containerlab](https://containerlab.dev/).

## Prerequisites

- **Docker** installed and running.
- **Containerlab** installed (`curl -sL https://containerlab.dev/setup | sudo bash -s "all"`).

## Lab Topology

The provided `clab.yml` creates two Alpine Linux nodes acting as network switches with Telnet servers enabled.

### Configuration (`clab.yml`)
```yaml
name: telnet-lab

mgmt:
  network: clab_mgmt
  ipv4-subnet: 172.30.30.0/24

topology:
  nodes:
    switch1:
      kind: linux
      image: alpine:latest
      exec:
        - apk add --no-cache busybox-extras
        - "telnetd -p 23 -l /bin/sh" # Direct shell for testing
    switch2:
      kind: linux
      image: alpine:latest
      exec:
        - apk add --no-cache busybox-extras
        - "telnetd -p 23 -l /bin/sh"
```

## Running the Lab

1. **Deploy the topology**:
   ```bash
   sudo containerlab deploy -t clab.yml
   ```

2. **Verify the nodes**:
   Containerlab will output a table with the IP addresses (typically `172.30.30.2` and `172.30.30.3`).

## Running Tests

You can run the specialized Containerlab tests included in the repository:

```bash
go test -v clab_test.go telnet.go action.go builder.go session.go state.go errors.go
```

### What these tests verify:
- **Regex Prompts**: Matches the `/ #` shell prompt dynamically.
- **I/O Flow**: Executes `ls`, `uname`, and `uptime` commands.
- **ANSI Stripping**: Validates that terminal colors don't mess with the output.
- **Error Detection**: Uses `WithErrors` to catch "command not found" immediately.

## Cleaning Up

To stop and remove the virtual switches:
```bash
sudo containerlab destroy -t clab.yml
```

---

# Pruebas con Containerlab (Español)

Esta guía explica cómo probar la librería `telnetsession` usando un entorno de red virtual proporcionado por [Containerlab](https://containerlab.dev/).

## Requisitos

- **Docker** instalado y corriendo.
- **Containerlab** instalado.

## Despliegue

1. **Levantar el laboratorio**:
   ```bash
   sudo containerlab deploy -t clab.yml
   ```

2. **Ejecutar tests**:
   ```bash
   go test -v clab_test.go telnet.go action.go builder.go session.go state.go errors.go
   ```

3. **Destruir el laboratorio**:
   ```bash
   sudo containerlab destroy -t clab.yml
   ```
