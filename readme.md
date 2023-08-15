# gogen

a web api generator tool.

## 简介

这个是一个 web api 的生成工具，它可以根据你的接口生成对应的服务代码，这个项目原本是因为写 web 服务时, 需要从 request 中解析各个参数，这些都是重复代码，[写得烦了](intro.md)，就想用工具来生成，最后就有了这个项目，


它主要功能如下
1. 根据你定义的接口生成服务端框架代码
2. 根据你定义的接口生成客户端代码
3. 根据你定义的接口生成 swagger 文档

生成代码的原则： 确保它和手写一样清晰，简洁易懂


## 使用方法

### 1. 定义接口
````golang
type MoDomains interface {
  // @Summary get domain object by name
  // @Description get domain object by name
  // @ID MoDomains.GetByName
  // @Accept  json
  // @Produce  json
  // @Param   name      path   string     true  "domain name"
  // @Success 200 {string} string "ok"
  // @Failure 400 {object} string "We need domain name!!"
  // @Failure 404 {object} string "Can not find this domain"
  // @Router /by_name/{name} [get]
  GetByName(ctx context.Context, name string) (*MoDomain, error)
}
````

方法的标注是使用 [github.com/swaggo/swag](https://github.com/swaggo/swag) 的标注，具体文档请看 [swag](https://github.com/swaggo/swag)

### 2. 生成服务端代码

生成 github.com/labstack/echo （由-plugin=echo参数指定） 服务端代码，注意你也可以指定生成 chi, gin 等等

gogen server -plugin=echo file.go

它会生成下面代码

````golang
func InitMoDomains(mux loong.Party, svc MoDomains) {
  mux.GET("/by_name/:name", func(ctx *echo.Context) error {
    var name = ctx.Param("name")

    result, err := svc.GetByName(ctx.Request().Context, name)
    if err != nil {
      ctx.Error(fmt.Errorf("argument %q is invalid - %q", "key", s, err))
      return nil
    }
    return ctx.JSON(http.StatusOK, result)
  })
}
````

### 3. 生成客户端代码

gogen client domains.go

### 4. 生成文档, 请看 [github.com/swaggo/swag](https://github.com/swaggo/swag)，


## 文档

#### 方法中的参数名

    方法中的参数名和 [swag](https://github.com/swaggo/swag) 中参数名一般要求匹配，不匹配时会出错，
    匹配时我们将忽略大小写，这意味着方法中参数名为 ignoreCase, 在[swag](https://github.com/swaggo/swag) 中参数名为 ignorecase 是可以的
    匹配时还会尝试将参数名转为 SnakeCase 形式进行比较，这意味着方法中参数名为 ignoreCase, 在[swag](https://github.com/swaggo/swag) 中参数名为 ignore_case 是可以的

    这么做的原因是为了让你可以自定义参数在请求中的名字

   
#### 方法中的简单类型参数
    

   golang 的原生类型 int8, int16, int32, int64, int, uint8, uint16, uint32, uint64, uint, bool, float32, float64 都已经支持

   原生类型的指针都已经支持 

   原生类型的 slice 都已经支持 

   此外常见的  time.Time, net.IP,  net.HardwareAddr 也支持了


   此外常见的 database/sql 中的 sql.NullXXX 也支持了

   此外对 context.Context 做了特殊处理，将会从 Request.Context() 方法中获取。

   此外对 \*http.Request, http.ResponseWriter 也做了支持

#### 方法中的 struct 参数

  如果参法中的参数比较复杂，使用了 struct 也是支持的，规则如下
  struct 类型的参数不支持从 path 中取数，仅支持从 query 或 body 中取值, 
  
  如果从 query 取值时，参数的名称以 json 标注中的为准(没有 json 标注时，将字段名转换为  SnakeCase 形式 )


##### extensions(x-gogen-extend=inline)

   ````golang
      type QueryParam struct {
        Name string `json:"name"`
        Type string `json:"name"`
      }

      type Service interface {
        // @Param   param      query   string     true  "domain name"
        Get(param QueryParam) (XXXX, error)
      }
   ````

   Service.Get() 方法中 param 参数取值时如下

   ````golang
     var param QueryParam
     param.Name = ctx.QueryParam("param.name")
     param.Type = ctx.QueryParam("param.type")
   ````

     字段 param.Name 对应的 query 参数名为  param.name
     字段 param.Type 对应的 query 参数名为  param.type

     但有时我们不想将要这个参数名前面的 "param.",  这时我们可以加 extensions(x-gogen-extend=inline), 如下

   ````golang
      type QueryParam struct {
        Name string `json:"name"`
        Type string `json:"name"`
      }

      type Service interface {
        // @Param   param      query   string     true  "domain name" extensions(x-gogen-extend=inline)
        Get(param QueryParam) (XXXX, error)
      }
   ````

    加了 extensions(x-gogen-extend=inline) 后它们的参数名如下

     字段 param.Name 对应的 query 参数名为  name
     字段 param.Type 对应的 query 参数名为  type


#### 方法中的 body 参数

  当方法中的有且仅有一个（等于1时）参数被标注为从 body 中取值时，我们会将整个请求的 body 作为这个参数的值, 例如

   ````golang
      type Service interface {
        // @Param   record      body   Record     false  ""
        Save(record Record) error
      }
   ````

   Service.Save() 方法中 record 参数取值时如下
   ````golang
      var record Record
      ctx.ReadJSON(&record)
   ````


   当方法中的有多个（大于1时）参数被标注为从 body 中取值时，我们会将将所有参数转换为一个 struct 中， 并以参数名为字段名, 例如

   ````golang
      type Service interface {
        // @Param   name      body   string     false  ""
        // @Param   type      body   string     false  ""
        Save(name, description string) error
      }
   ````

   Service.Save() 方法中 record 参数取值时如下
   ````golang
      var bindArgs struct {
        Name string `json:"name"`
        Description string `json:"description"`
      }
      ctx.ReadJSON(&bindArgs)
   ````

##### extensions(x-gogen-entire-body=false)

    有时我们只传一个 struct 对象时，我们仍然想将这参数放在一个结构中，我们可以加上这个，如

   ````golang
      type Record struct {
        Name string `json:"name"`
        Description string `json:"description"`
      }

      type Service interface {
        // @Param   record      body   Record     false  "record"  extensions(x-gogen-entire-body=false)
        Save(record Record) error
      }
   ````
    这时 Service.Save() 方法生成的服务端代码就变成了
   ````golang
      var bindArgs struct {
        Record Record `json:"record"`
      }
      ctx.ReadJSON(&bindArgs)
   ````

    Service.Save() 方法传参数时就要这么传

     ````json
     {
      "record": {
        "name": "xxx",
        "description": "xxx"
      }
     }
     ````


#### 方法中的返回参数

方法中的返回参数中必须有一个 error 参数，并且它必须是最后一个参数。

##### 只有两个返回参数（且其中一个为 error）

   生成的代码如下

   ````golang
    result, err := svc.XXXX(key)
    if err != nil {
      return ctx.JSON(httpCodeWith(err), err)
    }
    return ctx.JSON(http.StatusOK, result)
   ````

##### 两个以上返回参数

   两个以上返回参数时我们求每个返回参数都要有名称, http 请求返回时我们会将所有返回参数放到一个对象中，并以每个返回参数作为对象的字段，参返回参数的名称作为字段名。

   ````golang
       Get()  (result1 int, result2 int, result3 int, err error)
   ````

  生成的代码如下

   ````golang
    result1, result2, result3, err := svc.XXXX(key)
    if err != nil {
      return ctx.JSON(httpCodeWith(err), err)
    }
    return ctx.JSON(http.StatusOK, map[string]interface{}{
      "result1": result1,
      "result2": result2,
      "result3": result3,
    })
   ```` 


##### 只有一个 error 返回参数

   生成的代码如下

   ````golang
    err := svc.XXXX(key)
    if err != nil {
      return ctx.JSON(httpCodeWith(err), err)
    }
    return ctx.JSON(http.StatusOK, "OK")
   ````

######  @x-gogen-noreturn

   有时侯我们有一些特列要求需要直接处理 http 的返回， 如下

   ````golang
      type Service interface {
        Test(w http.ResponseWriter, xxx string) error
      }
   ````

    在 Service.Test() 方法中我们已经处理了响应(http.ResponseWriter)，不然希望生成代码时对响应再次处理, 这时我们再生成如下代码就不对了

    ````golang
    err := svc.Test(ctx.ResponseWriter(), xxx)
    if err != nil {
      return ctx.JSON(httpCodeWith(err), err)
    }
    return ctx.JSON(http.StatusOK, "OK")
    ````

   最后的 return ctx.JSON(http.StatusOK, "OK") 是多余的

   这时我们加上   @x-gogen-noreturn 就能正确处理了， 生成代码如下

  ````golang
    err := svc.Test(ctx.ResponseWriter(), xxx)
    if err != nil {
      return ctx.JSON(httpCodeWith(err), err)
    }
    return nil
   ````



## 最后

v1 目录是老代码，当初方法的参数很简单，就是各种简单的基本类型，后来要求越来越多，类型越来越复杂，但工作太忙，没有时间重构，只能堆代码，变得很臭。

v2 目录是我有时间重构后的，我们现在在使用这个，
