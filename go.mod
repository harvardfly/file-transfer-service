module file-transfer-service

go 1.14

require (
	github.com/gin-gonic/gin v1.6.3
	github.com/golang/protobuf v1.4.0
	github.com/juju/ratelimit v1.0.2-0.20191002062651-f60b32039441
	github.com/micro/cli v0.2.0
	github.com/micro/go-micro/v2 v2.5.1-0.20200417165434-16db76bee2fb
	github.com/micro/go-plugins/registry/etcdv3 v0.0.0-20200119172437-4fe21aa238fd
	github.com/micro/go-plugins/registry/etcdv3/v2 v2.5.0 // indirect
	github.com/micro/go-plugins/wrapper/breaker/hystrix v0.0.0-20200119172437-4fe21aa238fd
	github.com/micro/go-plugins/wrapper/breaker/hystrix/v2 v2.5.0 // indirect
	github.com/micro/go-plugins/wrapper/ratelimiter/ratelimit v0.0.0-20200119172437-4fe21aa238fd
	github.com/micro/go-plugins/wrapper/ratelimiter/ratelimit/v2 v2.5.0 // indirect
	github.com/micro/protoc-gen-micro/v2 v2.3.0 // indirect
	google.golang.org/protobuf v1.22.0
)
