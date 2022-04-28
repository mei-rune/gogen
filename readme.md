# gogen

a web api generator tool.

## 简介

这个是一个 web api 的生成工具，它可以根据你的接口生成对应的服务代码，这个项目原本是因为写 web 服务时, 需要从 request 中解析各个参数，这些都是重复代码，写得烦了，就想用工具来生成，最后就有了这个项目，


它主要功能如下
1. 根据你定义的接口生成服务端框架代码
2. 根据你定义的接口生成客户端代码
3. 根据你定义的接口生成 swagger 文档

生成代码的原则： 确保它和手写一样清晰，简洁易懂


## 使用方法

1. 定义接口
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

2. 生成服务端代码

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

3. 生成客户端代码

gogen client domains.go

4. 生成文档, 请看 [github.com/swaggo/swag](https://github.com/swaggo/swag)，


## 最后

v1 目录是老代码，当初方法的参数很简单，就是各种简单的基本类型，后来要求越来越多，类型越来越复杂，但工作太忙，没有时间重构，只能堆代码，变得很臭。

v2 目录是我有时间重构后的，我们现在在使用这个，