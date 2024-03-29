package protocplugin

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/pluginpb"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
)

type TProtocgenRpc struct {
	Handler ProtocPluginHandler
	Port    int
}

type TrpcReply struct {
	ResponseByte []byte
}

const DefaultProtogenPort = 20000

func (this *TProtocgenRpc) CallReqRes(request *pluginpb.CodeGeneratorRequest) (response *pluginpb.CodeGeneratorResponse, err error) {

	reqBytes, err := proto.Marshal(request)
	if err != nil {
		return nil, err

	}

	rpcReply := &TrpcReply{
		ResponseByte: make([]byte, 0),
	}

	err = this.CallBytes(reqBytes, rpcReply)
	if err != nil {
		return nil, err
	}

	resp := &pluginpb.CodeGeneratorResponse{}
	err = proto.Unmarshal(rpcReply.ResponseByte, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil

}
func (this *TProtocgenRpc) CallBytes(request []byte, reply *TrpcReply) error {
	client, err := rpc.DialHTTP("tcp", "localhost:"+strconv.Itoa(this.Port))
	if err != nil {
		return err
	}

	err = client.Call("TProtocgenRpc.OnCallBytes", request, reply)
	if err != nil {
		return err

	}

	return nil
}
func (this *TProtocgenRpc) OnCallReqRes(request *pluginpb.CodeGeneratorRequest) (response *pluginpb.CodeGeneratorResponse) {
	response = &pluginpb.CodeGeneratorResponse{}
	genFiles := this.Handler(request)
	response.File = append(response.File, genFiles...)

	return response

}
func (this *TProtocgenRpc) OnCallBytes(requestBytes []byte, rpcreply *TrpcReply) error {

	request := &pluginpb.CodeGeneratorRequest{}
	err := proto.Unmarshal(requestBytes, request)
	if err != nil {
		return err
	}

	response := this.OnCallReqRes(request)

	respBytes, err := proto.Marshal(response)
	if err != nil {
		return err
	}

	rpcreply.ResponseByte = append(rpcreply.ResponseByte, respBytes...)

	return nil
}

func (this *TProtocgenRpc) Run() {
	err := rpc.Register(this)
	if err != nil {
		log.Fatal("rpc.Register error:", err)
		return
	}
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", ":"+strconv.Itoa(this.Port))
	if err != nil {
		log.Fatal("listen error:", err)
	}
	http.Serve(l, nil)

}
