

[gogen](https://github.com/runner-mei/gogen) 是一个服务端 web api 代码生成工具（ https://github.com/runner-mei/gogen ）， 


## 背景

我很久以前用 java, 后来改成使用 golang， 用 golang 写过很多 restful 的接口，用过 gin, echo 等各种架构，但这些框架大同小异，都是各种 Get, Post, Delete, Put 方法，方法上都是只有一个 Context 参数，使用方法大至如下：


 ````golang
    mux.Get("/xxxx", func(ctx XContxt) {
    	// 各种从 ctx 中读参数, 如
    	idStr := ctx.PathParam("id")
    	id, err := strconv.ParseInt(idStr, 10, 64)
    	if err != nil {
    		ctx.TEXT(http.StatusBadRequest, err.Error())
    		return
    	}
    	a := ctx.QueryParam("a")


    	// 业务处理，如从数据库中读，等等
        ......

    	// 返回, 无非是 JSON, TEXT 之类的
    	ctx.JSON(http.StatusOK, result)
    })
 ````

大致流程都是先从 ctx 读参数， 再业务处理， 再返回， 当参数少时还行，参数多时读数据，再转换真的好烦，我就想为什么不能像 java 的框架一样呢

 ````java
	    @Path("/show-on-screen")
		public class JerseyHelloWorldService
		{
		    @GET
		    @Path("/{message}")
		    public string getMsg(@PathParam("message") String msg)
		    {
		        return "Message requested : " + msg;
		    }
		}
 ````

可能原来因为 golang 不支持在方法上加 [tag literal](https://github.com/golang/go/issues/18702) 吧, golang 短时间内不会加这个啦， 但我写的 restful 太多了， 所以就想用生成代码的方式来解决这个问题。


正好有 [github.com/swaggo/swag](https://github.com/swaggo/swag) 这个项目，它的标注功能正好是我想要的， 而且它可以生成文档，也是我想要的，所有我在它的基础上开了一个代码生成工具， 它可以自动帮你生成好框架代码，不用再关注参数的读取和返回对象的序列化等问题。


## 使用方法

### 1. 定义接口
````golang
type MoDomains interface {
  // @Summary get domain object by id
  // @Description get domain object by id
  // @ID MoDomains.GetByID
  // @Accept  json
  // @Produce  json
  // @Param   id      path   int     true  "domain id"
  // @Success 200 {string} string "ok"
  // @Failure 400 {object} string "We need domain id!!"
  // @Failure 404 {object} string "Can not find this domain"
  // @Router /domains/{id} [get]
  GetByID(id int64) (*MoDomain, error)
}
````

方法的标注是使用 [github.com/swaggo/swag](https://github.com/swaggo/swag) 的标注，具体文档请看 [swag](https://github.com/swaggo/swag)

### 2. 生成服务端代码

生成 github.com/gin-gonic/gin （由-plugin=gin参数指定） 服务端代码，注意你也可以指定生成 chi, echo, gin 等等

gogen server -plugin=chi file.go

它会生成下面代码

````golang
func InitMoDomains(mux gin.IRouter, svc MoDomains) {
  mux.GET("/domains/:id", func(ctx *gin.Context) {
    id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
    if err != nil {
			ctx.JSON(http.StatusBadRequest, NewBadArgument(err, "CaseSvc.TestCase3_1", "id"))
			return
    }
    result, err := svc.GetByID(id)
    if err != nil {
		ctx.JSON(httpCodeWith(err), err)
		return
    }
	 ctx.JSON(http.StatusOK, result)
	 return
  })
}
````

我们只要专注于实现 MoDomains 接口中就好了， 然后在 main 函数中注册路由就好了


````golang
func main() {
	r := gin.Default()

	var svc = NewMoDomains(xxx)
	InitMoDomains(r.Group("/test"), svc)
	r.Run()
}
````

### 3. 生成客户端代码

gogen client domains.go

### 4. 生成文档, 请看 [github.com/swaggo/swag](https://github.com/swaggo/swag)，
