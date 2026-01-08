# protocgen

the repo for protoc generators

ygrpc is a low code system base on grpc(protobuf) idl, we adopt protobuf as the base for all info and  communication, and generate all the code for you, so you can focus on the business logic.

ygrpc是一个基于grpc(protobuf) idl的低代码系统，我们采用protobuf作为所有数据和通信的基础，并为您生成所有的代码，这样您就可以专注于业务逻辑。
虽然ygrpc会为您生成大部分代码，但是你还是得需要理解整个流程，我们并不是一个全自动的系统，我们只是帮助您减少重复的工作，提高开发效率，底层的流程还是很清晰的，你可以很容易的理解。

ygrpc的整个流程如下：

- 编写proto文件，定义好数据库表结构
  - 根据proto文件生成sql初始化脚本
  - 根据proto文件生成一些辅助的proto message, 如 message_list, crud_rpc
  - 根据proto文件生成crud操作代码，包括后端go代码，前端ts代码, 和rpc代码

cmd下面的工具说明：

- protoc-gen-ygrpc-sql: 生成sql初始化脚本
- protoc-gen-ygrpc-msglist: 生成protobuf message list
- protoc-gen-ygrpc-cgo: 生成可通过 C ABI 调用的 Go 导出函数（CGO / buildmode=c-shared）

## protoc-gen-ygrpc-cgo（当前已实现能力）

该插件用于把 proto 里的 RPC 方法生成一层 Go 导出函数（`//export`），并可用 `go build -buildmode=c-shared` 产出 `.so` + `.h`，供 C/C++/其他语言通过 C ABI 调用。

### 1) 编译插件

在本仓库根目录执行：

```bash
go build -o protoc-gen-ygrpc-cgo ./cmd/protoc-gen-ygrpc-cgo
```

### 2) 用 protoc 触发生成

示例（假设你的 proto 在当前目录，输出到 `./gen`）：

```bash
protoc \
  -I . \
  --plugin=protoc-gen-ygrpc-cgo=./protoc-gen-ygrpc-cgo \
  --ygrpc-cgo_out=./gen \
  your.proto
```

生成目录内会包含：

- `gen/ygrpc_cgo/ygrpc_runtime.go`：运行时支持（错误模型等）
- `gen/ygrpc_cgo/<proto_filename>.ygrpc.cgo.go`：按 proto 生成的导出函数（当前为最小可编译 scaffold）

### 3) 生成 C 共享库与头文件

在生成目录（例如 `./gen`）中创建一个最小 `go.mod`（module 名可自定），然后 build：

```bash
cd gen
cat > go.mod <<'EOF'
module cgogen

go 1.23.0
EOF

go build -buildmode=c-shared -o libygrpc.so ./ygrpc_cgo
```

成功后会产出：

- `libygrpc.so`
- `libygrpc.h`（包含所有导出符号原型）

### 4) 已实现的接口形态（Binary Unary）

当前阶段（scaffold）已经实现并通过测试验证的能力：

- **Binary Mode (Unary)**：每个 unary RPC 生成导出函数 `Service_Method`（返回 `int`）
- **错误模型**：返回值 `0` 表示成功；非 `0` 表示 `errorId`，并可在 3 秒内调用 `Ygrpc_GetErrorMsg(errorId, ...)` 获取错误消息
- **Response 输出**：始终以 `(ptr, len, free)` 三元组输出（通过 output params）
- **Request 输入默认不带 free**：默认导出函数签名仅包含 `(ptr, len)`

### 5) Request Free option（0/1/2）

支持在 request message 上定义一个自定义 option，用于控制 request `free` 的导出策略：

- option=0（默认）：仅生成默认符号 `Service_Method`（不包含 request `free`）
- option=1：仅生成 `Service_Method_TakeReq`（包含 request `free`）
- option=2：同时生成 `Service_Method` + `Service_Method_TakeReq`

注意：当前实现通过读取 `google.protobuf.MessageOptions` 的 unknown fields 来解析该 option，要求该 option 的字段号为 `50001`。

proto 示例：

```proto
import "google/protobuf/descriptor.proto";

extend google.protobuf.MessageOptions {
  int32 ygrpc_cgo_req_free = 50001;
}

message MyReq {
  option (ygrpc_cgo_req_free) = 2; // 0/1/2
  bytes data = 1;
}
```

### 6) 重要说明（当前限制）

当前生成的导出函数仍是“可编译的最小骨架（scaffold）”：函数体不包含真实的 marshal/unmarshal 与业务 handler 调用逻辑，主要用于固定 ABI 形态、验证 cgo/buildmode=c-shared 可用与头文件原型符合约定。
