package main

import (
	"github.com/ygrpc/protocgen/protocplugin"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	logPrefix := "protoc-gen-ygrpc-protocsave: "
	log.SetPrefix(logPrefix)

	var protocgenPort int

	//search protocgen-port from os.Args
	for i := 0; i < len(os.Args); i++ {
		log.Println("os.Args[", i, "]=", os.Args[i])
		if strings.HasPrefix(os.Args[i], "--protocgen-port") {
			eqPos := strings.Index(os.Args[i], "=")
			if eqPos == -1 {
				log.Println("protocgen-port is not parse ok, use default port:", protocplugin.DefaultProtogenPort)
				protocgenPort = protocplugin.DefaultProtogenPort
				break
			}
			protocgenPort, _ = strconv.Atoi(os.Args[i][eqPos+1:])
			if protocgenPort <= 0 {
				log.Println("protocgen-port is not parse ok, use default port:", protocplugin.DefaultProtogenPort)
				protocgenPort = protocplugin.DefaultProtogenPort
			}
			break
		}
	}

	if protocgenPort != 0 {
		log.Println("protocgen-port:", protocgenPort)

		server := &protocplugin.TProtocgenRpc{
			Port:    protocgenPort,
			Handler: protocSaveHandler,
		}

		server.Run()
	} else {
		_, _, err := protocplugin.ProtoGeneratorMain(protocSaveHandler, os.Stdin, os.Stdout, logPrefix)
		if err != nil {
			log.Fatalf("error: failed to execute protoc plugin handler: %v", err)
		}
	}

}

func protocSaveHandler(request *pluginpb.CodeGeneratorRequest) (genFiles []*pluginpb.CodeGeneratorResponse_File) {

	//request bytes
	requestBytes, _ := proto.Marshal(request)

	//get protoc plugin option

	filename := "ygrpc-protocsave-request.out"

	reqParam := request.GetParameter()
	log.Println("request para:", reqParam)

	//get filename from protoc plugin option like filename=xxx
	if filenamePos := strings.Index(reqParam, "filename="); filenamePos != -1 {
		filename = reqParam[filenamePos+9:]
	}

	//save requestBytes to the file with filename
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("error: failed to create file %s: %v", filename, err)
	}
	defer file.Close()

	if _, err := file.Write(requestBytes); err != nil {
		log.Fatalf("error: failed to write file %s: %v", filename, err)
	}

	return nil
}
