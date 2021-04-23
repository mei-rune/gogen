# gogen

这个项目原本是因为写 web 服务时, 需要从 request 中解析各个参数，这些都是重复代码，写得烦了，就想用工具来生成，最后就有了这个项目，当初想法简单，就是从 interface 中各个方法生成代码，方法的参数也简单，就是各种简单的基本类型，后来要求越来越多，类型越来越复杂，但工作太忙，没有时间重构，只能堆代码，变得很臭。

## 重构的想法
1. 接口用 protocolbuffers 来定义接口, 注意不要像 github.com/twitchtv/twirp, 要支持 google.api.http
     例如：
     ````
     service Messaging {
       rpc GetMessage(GetMessageRequest) returns (Message) {
         option (google.api.http) = {
             get: "/v1/users/{user_id}/messages/{message_id}"
         };
       }
     }
     message GetMessageRequest {
       string message_id = 1;
       string user_id = 2;
    }
    ````
    生成生后，interface 应该如下
     ````golang
      type Messaging interface {
            GetMessage(userID, messageID string) (Message, error)
      }
     ````
2. 或仍然用 interface 定义，但引用 github.com/swaggo/swag 的 annotations 

3. 或采用 github.com/tal-tech/go-zero 的语法

4. 增加生成 openapi 的文档

## 当前状态
生成器的代码很乱，但生成的代码很漂亮, 和手写区别不大(这是当初的目标，也是以后的目标)，生成后的代码不依赖本项目(这是当初的目标，也是以后的目标)。
