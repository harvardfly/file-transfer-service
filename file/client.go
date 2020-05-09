package main

import (
	"context"
	fileConfig "file-transfer-service/file/config"
	pb1 "file-transfer-service/file/proto"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/transport/grpc"
	"github.com/micro/go-micro/v2/web"
	"github.com/micro/go-plugins/registry/etcdv3/v2"
	"github.com/micro/go-plugins/wrapper/breaker/hystrix/v2"
	"io"
	"log"
	"net/http"
)

var c client.Client
var fileService pb1.FileService

func UploadFile(ctx *gin.Context) {
	// 取到文件对象
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.String(
			http.StatusBadRequest,
			fmt.Sprintf("get form err: %s", err.Error()))
		return
	}
	files := form.File["files"]
	file, err := files[0].Open()
	if err != nil {
		ctx.String(
			http.StatusBadRequest,
			fmt.Sprintf("get form err: %s", err.Error()))
		return
	}
	// 建立链接
	// 因为这里是用的临时文件储存的方式,如果因为负载均衡算法导致下一次节点切换,另外一个节点是无法通过,文件名来获取到文件数据的
	// 使用这种方法来固定一个节点
	next, _ := c.Options().Selector.Select("file.rpc.server")
	node, _ := next()
	stream, err := fileService.File(context.Background(), func(options *client.CallOptions) {
		// 指定节点
		options.Address = []string{node.Address}
	})
	if err != nil {
		ctx.String(
			http.StatusBadRequest,
			fmt.Sprintf("get form err: %s", err.Error()))
		return
	}
	for {
		buff := make([]byte, 1024*1024) // 缓冲1MB,每次发送1MB的内容,注意不能超过rpc的限制(grpc默认为4MB)
		sendLen, err := file.Read(buff)
		if err != nil {
			if err == io.EOF {
				//全部读取完成,发送一个完成标识,跳出
				err = stream.Send(&pb1.FileRequest{
					Byte: nil,
					Len:  -1,
				})
				if err != nil {
					ctx.String(
						http.StatusBadRequest,
						fmt.Sprintf("get form err: %s", err.Error()))
					return
				}
				break
			}
			ctx.String(
				http.StatusBadRequest,
				fmt.Sprintf("get form err: %s", err.Error()))
			return
		}
		err = stream.Send(&pb1.FileRequest{
			Byte: buff[:sendLen],
			Len:  int64(sendLen),
		})
		if err != nil {
			ctx.String(
				http.StatusBadRequest,
				fmt.Sprintf("get form err: %s", err.Error()))
			return
		}
	}
	// 等待接收，当收到的消息之后就可以关闭了
	fileMsg := &pb1.FileResponse{}
	if err := stream.RecvMsg(fileMsg); err != nil {
		ctx.String(
			http.StatusBadRequest,
			fmt.Sprintf("get form err: %s", err.Error()))
		return
	}
	_ = stream.Close()
	ctx.JSON(http.StatusOK, gin.H{"fileName": fileMsg.FileName})
}

func main() {
	fileApiFlag := cli.StringFlag{
		Name:  "f",
		Value: "./config/config_api.json",
		Usage: "please use xxx -f config_api.json",
	}
	configFile := flag.String(fileApiFlag.Name, fileApiFlag.Value, fileApiFlag.Usage)
	flag.Parse()
	conf := new(fileConfig.ApiConfig)

	if err := config.LoadFile(*configFile); err != nil {
		log.Fatal(err)
	}
	if err := config.Scan(conf); err != nil {
		log.Fatal(err)
	}
	etcdRegisty := etcdv3.NewRegistry(
		func(options *registry.Options) {
			options.Addrs = conf.Etcd.Address
		});
	fileRpcService := micro.NewService(
		micro.Name("file.client"),
		micro.Registry(etcdRegisty),
		micro.Transport(grpc.NewTransport()),
		micro.WrapClient(hystrix.NewClientWrapper()), // 客户端熔断
	)

	fileRpcService.Init()
	c = fileRpcService.Client()
	// 创建用户服务客户端 直接可以通过它调用user rpc的服务
	fileService = pb1.NewFileService(conf.Server.Name, c)

	service := web.NewService(
		web.Name(conf.Server.Name+".web"),
		web.Version(conf.Version),
		web.Address(conf.Port),
	)

	router := gin.Default()
	fileRouterGroup := router.Group("/file")
	{
		fileRouterGroup.POST("/upload", UploadFile)
	}
	service.Handle("/", router)
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
