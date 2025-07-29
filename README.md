# telnetsession

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![Version](https://img.shields.io/badge/version-v0.1.0-green.svg)](https://github.com/WinAndronuX/telnetsession/releases/tag/v0.1.0)

Spanish version: [README.es.md](README.es.md)

---

The `telnetsession` package provides an interface for creating and executing telnet sessions on remote devices.

### What is a `session`?

Unlike other telnet packages, with `telnetsession` you must define a `session` using 
```telnetsession.NewBuilder()```. In this configuration you can specify parameters such as timeout, 
Enter character, prompt, among others. Additionally, you must declare the actions to execute using `Send()`, 
`Expect()` and their variants.

Once you have declared all the actions in a `session`, you can execute it. This approach provides 
you with more precise control when sending commands and handling responses, plus it offers 
better error handling.

A standout feature is the ability to send templates (https://pkg.go.dev/text/template) 
and use dynamic data to generate commands programmatically.

## Features

- **Fluid action system** for sending commands and waiting for responses
- **Template support** that allows generating commands dynamically
- **Success callbacks** for automatically processing responses
- **Integrated login handling** with username and password
- **Configurable timeouts** for connection, reading and writing
- **Builder interface** for fluid and readable configuration

## Installation

```bash
go get github.com/WinAndronuX/telnetsession
```

## Basic Usage

### Simple Example

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/WinAndronuX/telnetsession"
)

func main() {
    // Configure the session
    session, err := telnetsession.NewBuilder().
        WithTimeout(5 * time.Second).
        SetEnter("\n").
        SetLoginExpr("Login:", "Password:").
        SetPrompt(">").
        Send("show version").
        Build()
    
    if err != nil {
        log.Fatal(err)
    }
    
    // Create and execute the session
    device := telnetsession.New(session)
    
    if err := device.Run("192.168.1.1", 23, "admin", "password"); err != nil {
        log.Fatal(err)
    }
    
    // Get the output
    fmt.Println(device.GetOutput())
}
```

### Example with Callbacks

```go
session, err := telnetsession.NewBuilder().
    WithTimeout(5 * time.Second).
    SetEnter("\n").
    SetLoginExpr("Login:", "Password:").
    SetPrompt(">").
    Send("enable").
    SetPrompt("#").
    Expect("Password:").
    Send("enable_password").
    SendAndDo("show version", func(output string) error {
        fmt.Println("Device version:", output)
        return nil
    }).
    Build()
```

### Example with Templates

```go
data := map[string]any{
    "interfaces": []string{"eth0", "eth1", "eth2"},
}

session, err := telnetsession.NewBuilder().
    WithTimeout(5 * time.Second).
    SetEnter("\n").
    SetLoginExpr("Login:", "Password:").
    SetPrompt(">").
    SendTempl(`
{{ range .interfaces }}
interface {{ . }}
show interface
exit
{{ end }}`, data).
    Build()
```

## API Reference

### SessionBuilder

The `SessionBuilder` provides a fluid interface for configuring telnet sessions.

#### Configuration Methods

- `WithTimeout(duration)` - Sets the connection timeout
- `WithReadTimeout(duration)` - Sets the read timeout
- `WithWriteTimeout(duration)` - Sets the write timeout
- `SetEnter(string)` - Sets the end-of-line character
- `SetPrompt(string)` - Sets the prompt character
- `SetLoginExpr(username, password)` - Sets the login expressions

#### Action Methods

- `Send(text)` - Sends text to the device
- `SendAndDo(text, callback)` - Sends text and executes a callback to process the response
- `SendTempl(template, data)` - Sends a template
- `SendTemplAndDo(template, data, callback)` - Sends a template and executes a callback to process the response
- `Expect(text)` - Waits for a specific text to appear
- `ExpectAndDo(text, callback)` - Waits for specific text and executes a callback to process the response

### TelnetSession

The main telnet session that handles the connection and executes the actions.

#### Main Methods

- `Run(host, port, user, pass)` - Executes the complete session
- `GetOutput()` - Gets the accumulated output from the session

## Best Practices

- **Automatic prompt management**: It's not necessary to manually wait for the prompt (`#`, `>`, `$`) using `Expect("#")`, as the library handles this automatically. Ignoring this recommendation can cause unexpected errors.

- **SetEnter configuration**: The `SetEnter` method should be configured only once and always before declaring any actions. If not specified, the default character is `\n`.

## Timeout Configuration

```go
session, err := telnetsession.NewBuilder().
    WithTimeout(10 * time.Second).        // Connection timeout
    WithReadTimeout(5 * time.Second).     // Read timeout
    WithWriteTimeout(3 * time.Second).    // Write timeout
    // ... rest of configuration
    Build()
```

## Prompt Handling

Prompts are used to wait for device responses after sending commands:

```go
session, err := telnetsession.NewBuilder().
    SetPrompt(">").           // Normal prompt
    Send("enable").
    SetPrompt("#").           // Privileged prompt
    Send("configure terminal").
    SetPrompt("(config)#").   // Configuration prompt
    Build()
```

## Success Callbacks

Callbacks allow automatically processing responses:

```go
session, err := telnetsession.NewBuilder().
    SendAndDo("show interfaces", func(output string) error {
        // Process command output
        if strings.Contains(output, "up") {
            fmt.Println("Active interface found")
        }
        return nil
    }).
    Build()
```

## Templates

Templates allow generating commands dynamically using Go template syntax:

```go
data := map[string]any{
    "vlans": []int{10, 20, 30},
    "name": "VLAN_",
}

session, err := telnetsession.NewBuilder().
    SendTempl(`
{{ range .vlans }}
vlan {{ . }}
name {{ $.name }}{{ . }}
exit
{{ end }}`, data).
    Build()
```

## Examples

[Click here](/examples)

## Error Handling

The library provides descriptive errors to facilitate debugging:

```go
if err := device.Run(host, port, user, pass); err != nil {
    log.Printf("Session error: %v", err)
    return
}
```

## Requirements

- Go 1.23 or higher

## License

This project is under the MIT license.

## Contributions

Contributions are welcome. Please open an issue or pull request for suggestions or improvements. 