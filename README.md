# file-transfer-service
go micro 批量文件处理service

## 技术架构
```$xslt
1.微服务框架：go micro 基于go-plugins可插拔模式
2.服务发现：etcd
3.服务端限流：令牌桶(ratelimit)
4.熔断机制：hystrix
5.web框架gin：充当api网关的作用
```

## 注意事项
rpc服务使用micro v2实现的，v2版本在v1的基础上做出了不小的改动，v2版本推荐使用etcd

## 系统环境要求
```$xslt
golang >= 1.14
go micro >= 2.5
```

## 文件传输服务
### 生成pb
```$xslt
cd file/protos
protoc --proto_path=. --micro_out=. --go_out=. file.proto
```
### 启动rpc服务
```$xslt
cd /file
go run server.go
```

### 客户端以web api方式调用rpc服务
```$xslt
go run client.go
POST localhost:8088/file/upload
```