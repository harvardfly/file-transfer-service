package main

import (
	"context"
	fileConfig "file-transfer-service/file/config"
	pb "file-transfer-service/file/proto"
	"flag"
	"fmt"
	rl "github.com/juju/ratelimit"
	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-plugins/registry/etcdv3/v2"

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/errors"
	"github.com/micro/go-micro/v2/transport/grpc"
	"github.com/micro/go-plugins/wrapper/ratelimiter/ratelimit/v2"
	"io/ioutil"
	"log"
)

type FileDo struct {
}

func (f *FileDo) File(ctx context.Context, file pb.File_FileStream) error {
	// 创建临时文件
	temp, err := ioutil.TempFile("", "micro")
	if err != nil {
		return errors.InternalServerError("file.service", err.Error())
	}
	for {
		req, err := file.Recv()
		if err != nil {
			return errors.InternalServerError("file.service", err.Error())
		}
		// 文件末尾
		if req.Len == -1 {
			break
		}
		_, err = temp.Write(req.Byte)
		if err != nil {
			return errors.InternalServerError("file.service", err.Error())
		}
	}
	fmt.Println("fileName:", temp.Name())

	return file.SendMsg(&pb.FileResponse{
		FileName: temp.Name(),
	})
}

func main() {
	fileRpcFlag := cli.StringFlag{
		Name:  "f",
		Value: "./config/config_rpc.json",
		Usage: "please use xxx -f config_rpc.json",
	}
	configFile := flag.String(fileRpcFlag.Name, fileRpcFlag.Value, fileRpcFlag.Usage)
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
	// 服务端限流
	b := rl.NewBucketWithRate(
		float64(conf.Server.RateLimit),
		int64(conf.Server.RateLimit),
	)
	// 创建RPC服务
	service := micro.NewService(
		micro.Name(conf.Server.Name),
		micro.Registry(etcdRegisty),
		micro.Transport(grpc.NewTransport()),
		micro.WrapHandler(ratelimit.NewHandlerWrapper(b, false)),
		//micro.Flags(fileRpcFlag),
	)
	service.Init()
	err := pb.RegisterFileHandler(service.Server(), new(FileDo))
	if err != nil {
		return
	}

	err = service.Run()
	if err != nil {
		return
	}
}
