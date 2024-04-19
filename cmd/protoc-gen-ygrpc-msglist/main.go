package main

import (
	"bytes"
	"github.com/ygrpc/protocgen/protocplugin"
	"github.com/ygrpc/protodb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"text/template"
	"time"
)

func main() {
	logPrefix := "protoc-gen-ygrpc-msglist: "

	log.SetPrefix(logPrefix)

	var protocgenPort int

	//search protocgen-port from os.Args
	for i := 0; i < len(os.Args); i++ {
		//log.Println("os.Args[", i, "]=", os.Args[i])
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
			Handler: protocMsgListHandler,
		}

		server.Run()
	} else {
		time.Sleep(20 * time.Second)
		_, _, err := protocplugin.ProtoGeneratorMain(protocMsgListHandler, os.Stdin, os.Stdout, logPrefix)
		if err != nil {
			log.Fatalf("error: failed to execute protoc plugin handler: %v", err)
		}
	}

}

func protocMsgListHandler(request *pluginpb.CodeGeneratorRequest) (genFiles []*pluginpb.CodeGeneratorResponse_File) {
	msgListProtoFileHead := `//generated by ygrpc-msglist. DO NOT EDIT.
//source: {{.original_file}}

syntax = "proto3";
package {{.package}};

{{ if .optimize_for }}option optimize_for = {{.optimize_for}};{{- end }}
{{ if .go_package }}option go_package = "{{.go_package}}";{{- end }}
{{ if .java_package }}option java_package = "{{.java_package}}";{{- end }}

import "{{.original_file}}"; 
`

	var fd *descriptorpb.FileDescriptorProto

	oneMsgList := `
// {{.Msg}} list
message   {{.Msg}}List {
	// batch no
	uint32 No = 1;

	// rows offset, start from 0
	uint32 Offset = 2;

	// data rows
	repeated {{.Msg}} Rows = 3;
}
`

	for _, fd = range request.GetProtoFile() {
		if !slices.Contains(request.FileToGenerate, fd.GetName()) {
			//not generate this file
			continue
		}
		needWriteThisFile := false

		original_file := fd.GetName()
		originalFilenameOnly := protocplugin.ExtractFilename(original_file)
		rpcFilename := originalFilenameOnly + ".msglist.proto"

		tHead := template.Must(template.New("proto_head").Parse(msgListProtoFileHead))
		sqlHeadData := map[string]interface{}{
			"original_file":     original_file,
			"original_filename": originalFilenameOnly,
			"package":           fd.GetPackage(),
			"optimize_for":      "",
			"go_package":        fd.GetOptions().GetGoPackage(),
			"java_package":      fd.GetOptions().GetJavaPackage(),
		}
		if fd.GetOptions().OptimizeFor != nil {
			sqlHeadData["optimize_for"] = fd.GetOptions().GetOptimizeFor().String()
		}

		var tplBytes bytes.Buffer
		if err := tHead.Execute(&tplBytes, sqlHeadData); err != nil {
			log.Fatal(err)
		}

		protoFileContent := tplBytes.String()

		tMsg := template.Must(template.New("rpc_msg").Parse(oneMsgList))

		//process every msg
		for _, msg := range fd.MessageType {
			msgName := msg.GetName()

			needMsgList := false

			if strings.HasPrefix(msgName, "Db") ||
				strings.HasPrefix(msgName, "DB") {
				needMsgList = true
			}

			if proto.HasExtension(msg.Options, protodb.E_Pdbm) {
				//get extension protodb.Pdbm
				pdbmInf := proto.GetExtension(msg.Options, protodb.E_Pdbm)
				pdbm := pdbmInf.(*protodb.PDBMsg)
				if pdbm != nil {
					if pdbm.MsgList == 4 {
						//no need msg list
						needMsgList = false
						continue
					}
				}
			}

			if !needMsgList {
				continue
			}

			tMsgData := map[string]interface{}{
				"Msg": msgName,
			}

			needWriteThisFile = true

			var msgBytes bytes.Buffer
			if err := tMsg.Execute(&msgBytes, tMsgData); err != nil {
				log.Fatal(err)
			}

			protoFileContent = protoFileContent + msgBytes.String()
		}

		if !needWriteThisFile {
			continue
		}

		genFile := &pluginpb.CodeGeneratorResponse_File{
			Name:    proto.String(rpcFilename),
			Content: proto.String(protoFileContent),
		}

		genFiles = append(genFiles, genFile)

	}
	return
}
