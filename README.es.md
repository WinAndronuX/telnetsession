# telnetsession

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![Version](https://img.shields.io/badge/version-v0.2.0-green.svg)](https://github.com/WinAndronuX/telnetsession/releases/tag/v0.2.0)

---

El paquete `telnetsession` proporciona una interfaz robusta para automatizar sesiones Telnet en dispositivos remotos como OLTs (Huawei, VSOL, Nokia, TP-Link), switches y routers.

Utiliza un **Patrón Builder Fluido** para definir una secuencia de acciones antes de la ejecución, ofreciendo control preciso y características modernas de Go.

## Características Principales

- **Arquitectura FSM**: Máquina de Estados Finita confiable para la gestión del ciclo de vida de la conexión.
- **Potencia de Regex**: Uso de expresiones regulares para prompts, disparadores de login y respuestas esperadas.
- **Paginación Automática**: Manejo de prompts estilo `--More--` en tiempo real.
- **Detección de Errores Fail-Fast**: Aborta la sesión inmediatamente al detectar errores del dispositivo.
- **Go Moderno**: Soporte completo para `context.Context` (cancelación/timeouts) y errores tipificados.
- **Protocolo Robusto**: Filtrado automático de Telnet IAC y limpieza de códigos ANSI (colores).
- **Plantillas**: Generación dinámica de comandos usando `text/template`.
- **Acciones Pre-Login**: Soporte para dispositivos excéntricos (como Nokia TL1) que requieren interacción antes del login.

## Instalación

```bash
go get github.com/WinAndronuX/telnetsession
```

## Uso Básico

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
        SetPrompt(`[>#]`). // Regex para prompts comunes
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

## Características Avanzadas

### Modo Privilegiado (Cisco)
```go
session, _ := builder.
    SetPrompt(">").
    Enable("privileged_password").
    SetPrompt("#").
    Send("show running-config").
    Build()
```

### Pre-Login (Estilo Nokia TL1)
```go
session, _ := builder.
    SendInitial("T").      // Enviar 'T' para seleccionar modo
    ExpectInitial("sure?").
    SendInitial("y").      // Confirmar
    SetLoginExpr("User:", "Pass:").
    Build()
```

### Callbacks y Plantillas
```go
data := map[string]any{"vlan": 100}
session, _ := builder.
    SendTempl("display vlan {{.vlan}}", data).
    SendAndDo("display current-configuration", func(output string) error {
        // Procesar salida aquí
        return nil
    }).
    Build()
```

## API Reference

### SessionBuilder
- `WithTimeout / WithReadTimeout / WithWriteTimeout`: Control de tiempos.
- `SetPrompt(regex)`: Establece el prompt esperado.
- `SetLoginExpr(user, pass)`: Configura expresiones de login.
- `WithPagination(pattern, response)`: Manejo de paginación.
- `WithErrors(patterns...)`: Abortar ante patrones de error.
- `WithDebug()`: Habilitar logs detallados de E/S.
- `Send / SendAndDo / SendTempl`: Enviar comandos.
- `Expect / ExpectAndDo`: Esperar salida específica.
- `Enable(password)`: Ayudante para modo privilegiado Cisco.
- `SendInitial / ExpectInitial`: Interacción pre-login.

### TelnetSession
- `Run(host, port, user, pass)`: Ejecución estándar.
- `RunWithContext(ctx, ...)`: Ejecución con soporte de cancelación.
- `GetOutput()`: Salida limpia (sin ANSI, sin IAC, líneas colapsadas).

## Licencia
Licencia MIT.
