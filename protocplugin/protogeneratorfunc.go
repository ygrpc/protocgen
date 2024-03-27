package protocplugin

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
	"io"
	"log"
	"strconv"
	"strings"
)

type ProtocPluginHandler func(request *pluginpb.CodeGeneratorRequest) (genFiles []*pluginpb.CodeGeneratorResponse_File)

func ExecProtocPluginHandler(request *pluginpb.CodeGeneratorRequest, genFunc ProtocPluginHandler) (response *pluginpb.CodeGeneratorResponse, err error) {

	response = &pluginpb.CodeGeneratorResponse{}

	genFiles := genFunc(request)

	response.File = append(response.File, genFiles...)

	return response, nil

}

// ProtoGeneratorMain protobuf plugin from main
func ProtoGeneratorMain(genFunc ProtocPluginHandler, in io.Reader, out io.Writer, logPrefix string) (request *pluginpb.CodeGeneratorRequest, response *pluginpb.CodeGeneratorResponse, err error) {
	if len(logPrefix) > 0 {
		log.SetPrefix(logPrefix)
	}

	data, err := io.ReadAll(in)
	if err != nil {
		return nil, nil, fmt.Errorf("error: reading input: %v", err)
	}

	request = &pluginpb.CodeGeneratorRequest{}
	if err := proto.Unmarshal(data, request); err != nil {
		return nil, nil, fmt.Errorf("error: failed to unmarshal input proto: %v", err)

	}

	reqParam := request.GetParameter()
	log.Println("request para:", reqParam)
	if strings.Contains(reqParam, "--protocgen-port") {

		response, err = ExecProtocPluginRpc(request, genFunc)
	} else {

		response, err = ExecProtocPluginHandler(request, genFunc)
	}

	if err != nil {
		return request, nil, fmt.Errorf("error: failed to execute protoc plugin handler: %v", err)

	}

	data, err = proto.Marshal(response)
	if err != nil {
		return request, nil, fmt.Errorf("error: failed to marshal output proto: %v", err)
	}
	if _, err := out.Write(data); err != nil {
		return request, response, fmt.Errorf("error: failed to write output proto: %v", err)
	}

	return request, response, nil
}

func ExecProtocPluginRpc(request *pluginpb.CodeGeneratorRequest, genFunc ProtocPluginHandler) (response *pluginpb.CodeGeneratorResponse, err error) {
	reqParam := request.GetParameter()

	protocgenPortPos := strings.Index(reqParam, "--protocgen-port")
	if protocgenPortPos == -1 {
		return nil, fmt.Errorf("error: failed to get protocgen-port from parameter: %v", reqParam)
	}

	protocgenPort := DefaultProtogenPort

	protocgenPortStr := reqParam[protocgenPortPos+16+1:]
	spacePos := strings.Index(protocgenPortStr, " ")
	if spacePos != -1 {
		protocgenPortStr = protocgenPortStr[:spacePos]
		port, err := strconv.Atoi(protocgenPortStr)
		if err != nil {
			return nil, fmt.Errorf("error: failed to get protocgen-port from parameter: %v", reqParam)
		}
		protocgenPort = port
	} else {
		if len(protocgenPortStr) > 0 {
			port, err := strconv.Atoi(protocgenPortStr)
			if err != nil {
				return nil, fmt.Errorf("error: failed to get protocgen-port from parameter: %v", reqParam)
			}
			protocgenPort = port
		}
	}

	rpcObj := &TProtocgenRpc{
		Port: protocgenPort,
	}

	response = &pluginpb.CodeGeneratorResponse{}
	response, err = rpcObj.CallReqRes(request)

	return response, err

}
