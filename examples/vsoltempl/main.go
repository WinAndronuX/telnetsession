package main

import (
	"fmt"
	"log"
	"time"

	"github.com/WinAndronuX/telnetsession"
)

func main() {

	var (
		host = "192.168.203.254"
		port = 23
		user = "admin"
		pass = "Xpon@Olt9417#"
	)

	var data = map[string]any{
		"gponInterfaces": []string{"gpon 0/1", "gpon 0/2", "gpon 0/3", "gpon 0/4"},
	}

	session, errSession := telnetsession.NewBuilder().
		WithTimeout(5*time.Second).
		SetEnter("\n").
		SetLoginExpr("Login", "Password").
		SetPrompt(">").
		Send("enable").
		SetPrompt("#").
		Expect("Password:").
		Send(pass).
		Send("terminal length 0").
		Send("configure terminal").
		SendTempl(`
{{ range .gponInterfaces }}
interface {{ . }}
show interface
exit
{{ end }}`, data).
		Build()

	if errSession != nil {
		log.Fatal(errSession)
	}

	device := telnetsession.New(session)

	if errRun := device.Run(host, port, user, pass); errRun != nil {
		log.Fatal(errRun)
	}

	fmt.Println("============ Console Output ============")
	fmt.Println(device.GetOutput())
	fmt.Println("========================================")
}
