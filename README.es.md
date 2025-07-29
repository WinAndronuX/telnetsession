# telnetsession

---

El paquete `telnetsession` proporciona una interfaz para crear y ejecutar sesiones telnet en dispositivos remotos.

### ¿Qué es una `session`?

A diferencia de otros paquetes telnet, con `telnetsession` debes definir una `session` utilizando 
```telnetsession.NewBuilder()```. En esta configuración puedes especificar parámetros como timeout, 
carácter de Enter, prompt, entre otros. Además, debes declarar las acciones a ejecutar usando `Send()`, 
`Expect()` y sus variantes.

Una vez que hayas declarado todas las acciones en una `session`, puedes ejecutarla. Este enfoque te 
proporciona un control más preciso al enviar comandos y manejar las respuestas, además de ofrecer 
un mejor manejo de errores.

Una característica destacada es la capacidad de enviar plantillas (https://pkg.go.dev/text/template) 
y utilizar datos dinámicos para generar comandos de forma programática.

## Características

- **Sistema de acciones fluido** para enviar comandos y esperar respuestas
- **Soporte para plantillas** que permite generar comandos dinámicamente
- **Callbacks de éxito** para procesar respuestas automáticamente
- **Manejo de login** integrado con username y password
- **Timeouts configurables** para conexión, lectura y escritura
- **Interfaz builder** para una configuración fluida y legible

## Instalación

```bash
go get github.com/WinAndronuX/telnetsession
```

## Uso Básico

### Ejemplo Simple

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/WinAndronuX/telnetsession"
)

func main() {
    // Configurar la sesión
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
    
    // Crear y ejecutar la sesión
    device := telnetsession.New(session)
    
    if err := device.Run("192.168.1.1", 23, "admin", "password"); err != nil {
        log.Fatal(err)
    }
    
    // Obtener la salida
    fmt.Println(device.GetOutput())
}
```

### Ejemplo con Callbacks

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
        fmt.Println("Versión del dispositivo:", output)
        return nil
    }).
    Build()
```

### Ejemplo con Plantillas

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

El `SessionBuilder` proporciona una interfaz fluida para configurar sesiones telnet.

#### Métodos de Configuración

- `WithTimeout(duration)` - Establece el timeout de conexión
- `WithReadTimeout(duration)` - Establece el timeout de lectura
- `WithWriteTimeout(duration)` - Establece el timeout de escritura
- `SetEnter(string)` - Establece el carácter de fin de línea
- `SetPrompt(string)` - Establece el carácter de prompt
- `SetLoginExpr(username, password)` - Establece las expresiones de login

#### Métodos de Acciones

- `Send(text)` - Envía texto al dispositivo
- `SendAndDo(text, callback)` - Envía texto y ejecuta un callback para procesar la respuesta
- `SendTempl(template, data)` - Envía una plantilla
- `SendTemplAndDo(template, data, callback)` - Envía una plantilla y ejecuta un callback para procesar la respuesta
- `Expect(text)` - Espera que aparezca un texto específico
- `ExpectAndDo(text, callback)` - Espera un texto específico y ejecuta un callback para procesar la respuesta

### TelnetSession

La sesión telnet principal que maneja la conexión y ejecuta las acciones.

#### Métodos Principales

- `Run(host, port, user, pass)` - Ejecuta la sesión completa
- `GetOutput()` - Obtiene la salida acumulada de la sesión

## Buenas prácticas

- **Gestión automática de prompts**: No es necesario esperar manualmente a que aparezca el prompt (`#`, `>`, `$`) usando `Expect("#")`, ya que la biblioteca maneja esto automáticamente. Ignorar esta recomendación puede causar errores inesperados.

- **Configuración de SetEnter**: El método `SetEnter` debe configurarse únicamente una vez y siempre antes de declarar cualquier acción. Si no se especifica, el carácter por defecto es `\n`.

## Configuración de Timeouts

```go
session, err := telnetsession.NewBuilder().
    WithTimeout(10 * time.Second).        // Timeout de conexión
    WithReadTimeout(5 * time.Second).     // Timeout de lectura
    WithWriteTimeout(3 * time.Second).    // Timeout de escritura
    // ... resto de configuración
    Build()
```

## Manejo de Prompts

Los prompts se utilizan para esperar respuestas del dispositivo después de enviar comandos:

```go
session, err := telnetsession.NewBuilder().
    SetPrompt(">").           // Prompt normal
    Send("enable").
    SetPrompt("#").           // Prompt privilegiado
    Send("configure terminal").
    SetPrompt("(config)#").   // Prompt de configuración
    Build()
```

## Callbacks de Éxito

Los callbacks permiten procesar las respuestas automáticamente:

```go
session, err := telnetsession.NewBuilder().
    SendAndDo("show interfaces", func(output string) error {
        // Procesar la salida del comando
        if strings.Contains(output, "up") {
            fmt.Println("Interfaz activa encontrada")
        }
        return nil
    }).
    Build()
```

## Plantillas

Las plantillas permiten generar comandos dinámicamente usando la sintaxis de Go templates:

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

## Ejemplos

[Haga click aquí](/examples)

## Manejo de Errores

La biblioteca proporciona errores descriptivos para facilitar el debugging:

```go
if err := device.Run(host, port, user, pass); err != nil {
    log.Printf("Error en la sesión: %v", err)
    return
}
```

## Requisitos

- Go 1.23 o superior

## Licencia

Este proyecto está bajo la licencia MIT.

## Contribuciones

Las contribuciones son bienvenidas. Por favor, abre un issue o pull request para sugerencias o mejoras. 