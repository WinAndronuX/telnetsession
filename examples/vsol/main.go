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

	session, errSession := telnetsession.NewBuilder().
		WithTimeout(5*time.Second).
		WithReadTimeout(10*time.Second).
		WithWriteTimeout(5*time.Second).
		SetEnter("\n").
		SetLoginExpr("Login", "Password").
		SetPrompt(">").
		Send("enable").
		SetPrompt("#").
		Expect("Password:").
		Send(pass).
		Send("terminal length 0").
		Send("configure terminal").
		SendAndDo("show version", func(value string) error {
			fmt.Println(value)
			return nil
		}).
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
