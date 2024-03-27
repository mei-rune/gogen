module github.com/runner-mei/gogen

require (
	emperror.dev/emperror v0.33.0 // indirect
	emperror.dev/errors v0.8.1 // indirect
	github.com/aryann/difflib v0.0.0-20170710044230-e206f873d14a
	github.com/astaxie/beego v1.12.3
	github.com/axw/gocov v1.1.0
	github.com/gin-gonic/gin v1.9.1
	github.com/go-chi/chi v1.5.4
	github.com/go-openapi/spec v0.20.9
	github.com/grsmv/inflect v0.0.0-20140723132642-a28d3de3b3ad
	github.com/k0kubun/colorstring v0.0.0-20150214042306-9440f1994b88 // indirect
	github.com/kataras/iris/v12 v12.2.0-beta1
	github.com/klauspost/compress v1.15.2 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/mattn/goveralls v0.0.12
	github.com/mitchellh/mapstructure v1.4.1
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/runner-mei/GoBatis v1.5.11-0.20240323160609-04e46ff83945
	github.com/runner-mei/log v1.0.10
	github.com/runner-mei/loong v1.1.22
	github.com/swaggo/echo-swagger v1.4.0
	github.com/swaggo/files v0.0.0-20210815190702-a29dd2bc99b2 // indirect
	github.com/swaggo/swag v1.16.1
	github.com/tdewolff/minify/v2 v2.11.2 // indirect
	github.com/ugorji/go v1.1.7 // indirect
	golang.org/x/tools v0.19.0 // indirect
)

replace github.com/swaggo/swag => github.com/runner-mei/swag v1.8.2-0.20231226075722-f02eee2df576

exclude github.com/dgrijalva/jwt-go v3.2.0+incompatible

exclude github.com/hjson/hjson-go v3.1.0+incompatible

go 1.13
