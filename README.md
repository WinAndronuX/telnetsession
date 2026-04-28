# telnetsession

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![Version](https://img.shields.io/badge/version-v0.2.0-green.svg)](https://github.com/WinAndronuX/telnetsession/releases/tag/v0.2.0)

Versión en español: [README.es.md](README.es.md)

---

The `telnetsession` package provides a robust interface for automating Telnet sessions on remote devices like OLTs (Huawei, VSOL, Nokia, TP-Link), switches, and routers.

It uses a **Fluent Builder Pattern** to define a sequence of actions before execution, offering precise control and modern Go features.

## Key Features

- **FSM Architecture**: Reliable Finite State Machine for connection lifecycle management.
- **Regex Powered**: Use regular expressions for prompts, login triggers, and expected responses.
- **Auto-Pagination**: Automatically handle `--More--` prompts in real-time.
- **Fail-Fast Error Detection**: Abort sessions immediately when device errors are detected.
- **Modern Go**: Full support for `context.Context` (cancellation/timeouts) and typed errors.
- **Protocol Robustness**: Automatic Telnet IAC (Interpret As Command) filtering and ANSI code stripping.
- **Templates**: Dynamic command generation using `text/template`.
- **Pre-Login Actions**: Support for eccentric devices (like Nokia TL1) requiring interaction before login.

## Installation

```bash
go get github.com/WinAndronuX/telnetsession
```

## Basic Usage

```go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/WinAndronuX/telnetsession"
)

func main() {
    session, _ := telnetsession.NewBuilder().
        WithTimeout(5 * time.Second).
        SetLoginExpr("Login:", "Password:").
        SetPrompt(`[>#]`). // Regex for common prompts
        WithPagination(`--More--`, " ").
        WithErrors(`Invalid command`, `Permission denied`).
        Send("show version").
        Build()
    
    device := telnetsession.New(session)
    if err := device.Run("192.168.1.1", 23, "admin", "password"); err != nil {
        log.Fatal(err)
    }
    
    fmt.Println(device.GetOutput())
}
```

## Advanced Features

### Cisco-style Enable Mode
```go
session, _ := builder.
    SetPrompt(">").
    Enable("privileged_password").
    SetPrompt("#").
    Send("show running-config").
    Build()
```

### Pre-Login (Nokia TL1 Style)
```go
session, _ := builder.
    SendInitial("T").      // Send 'T' to select mode
    ExpectInitial("sure?").
    SendInitial("y").      // Confirm
    SetLoginExpr("User:", "Pass:").
    Build()
```

### Success Callbacks & Templates
```go
data := map[string]any{"vlan": 100}
session, _ := builder.
    SendTempl("display vlan {{.vlan}}", data).
    SendAndDo("display current-configuration", func(output string) error {
        // Process output here
        return nil
    }).
    Build()
```

## API Reference

### SessionBuilder
- `WithTimeout / WithReadTimeout / WithWriteTimeout`: Time control.
- `SetPrompt(regex)`: Set the expected command prompt.
- `SetLoginExpr(userRegex, passRegex)`: Configure login triggers.
- `WithPagination(pattern, response)`: Handle auto-paging.
- `WithErrors(patterns...)`: Abort on specific output patterns.
- `WithDebug()`: Enable verbose logging of all I/O.
- `Send / SendAndDo / SendTempl`: Dispatch commands.
- `Expect / ExpectAndDo`: Wait for specific output.
- `Enable(password)`: Cisco privileged mode helper.
- `SendInitial / ExpectInitial`: Pre-login interaction.

### TelnetSession
- `Run(host, port, user, pass)`: Standard execution.
- `RunWithContext(ctx, ...)`: Execution with cancellation support.
- `GetOutput()`: Cleaned output (no ANSI, no IAC, collapsed lines).

### Pool
- `NewPool()`: Creates a new pool with automatic idle cleanup.
- `Do(ctx, session, host, port, user, pass)`: Executes a session with exclusive access per device.
- `GetActiveSession(host, port)`: Retrieve an ongoing session instance.
- `CountActiveSessions()`: Get total number of concurrent sessions.

## License
MIT License.
