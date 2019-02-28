// +build beego

package main

import "github.com/astaxie/beego"

func main() {
	ns := beego.NewNamespace("/v1")
	InitStringSvc(ns, &StringSvcImpl{})

	beego.AddNamespace(ns)
	beego.Run()
}
