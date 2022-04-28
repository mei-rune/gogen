//go:build gin
// +build gin

package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()
	InitStringSvc(r.Group("/test"), &StringSvcImpl{})
	r.Run() // listen and serve on 0.0.0.0:8080
}
