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
- **参数命名约定**：Binary 的 request/response 参数名统一使用 `in*` / `out*` 前缀，并包含对应消息类型名（例如 `inPingRequestPtr/inPingRequestLen[/inPingRequestFree]`，`outPingResponsePtr/outPingResponseLen/outPingResponseFree`）

### 5) Request Free option（0/1/2）

支持通过 proto option 配置 request `free` 的导出策略：

- option=0（默认）：仅生成默认符号 `Service_Method`（不包含 request `free`）
- option=1：仅生成 `Service_Method_TakeReq`（包含 request `free`）
- option=2：同时生成 `Service_Method` + `Service_Method_TakeReq`

注意：当前实现通过读取 Options 的 unknown fields 来解析该 option，要求字段号为 `50001`。

该策略支持两种配置入口与优先级（Method 优先于 File）：

- MethodOptions（方法级，优先级最高）
- FileOptions（文件级默认值）

由于 protobuf 的扩展名在同一 package 下必须全局唯一，File/Method 的扩展名建议使用不同名字（例如 `*_default` / `*_method`），但字段号保持为 `50001`。

proto 示例：

```proto
import "google/protobuf/descriptor.proto";

extend google.protobuf.FileOptions { int32 ygrpc_cgo_req_free_default = 50001; }
extend google.protobuf.MethodOptions { int32 ygrpc_cgo_req_free_method = 50001; }

// file-level default (applies to all methods unless overridden)
option (ygrpc_cgo_req_free_default) = 2;

message MyReq {
  bytes data = 1;
}

service Svc {
  rpc Ping(MyReq) returns (MyResp) {
    option (ygrpc_cgo_req_free_method) = 1;
  }
}
```

### 6) Native Mode（Unary）与禁用开关

当某个 unary RPC 的 request/response 都满足 “flat message” 条件时，当前实现会额外生成 `Service_Method_Native`（以及可能的 `Service_Method_Native_TakeReq`）导出函数：

- **入参展开**：request 的字段会展开为参数（例如 `string/bytes` 变为 `inXxxPtr/inXxxLen`；数值类标量变为 `inXxx`）
- **出参展开**：response 的 `string/bytes` 仍以 `(ptr, len, free)` 三元组输出（output params）
- **Request free 行为**：默认不带 `free`；当 `ygrpc_cgo_req_free_*` 配置触发 `*_TakeReq` 版本时，native 侧对应 `string/bytes` 入参会额外带 `inXxxFree`

此外支持通过方法级 option 显式关闭某个方法的 native 生成：

- option=0（默认）：生成 native（若满足 flat 条件）
- option=1：不生成 native

注意：当前实现通过读取 Options 的 unknown fields 来解析该 option，要求字段号为 `50002`。由于 protobuf 扩展名全局唯一，FileOptions 建议使用 `ygrpc_cgo_native_default`，MethodOptions 使用 `ygrpc_cgo_native`。

proto 示例：

```proto
import "google/protobuf/descriptor.proto";

extend google.protobuf.FileOptions {
  int32 ygrpc_cgo_native_default = 50002; // 0/1
}

extend google.protobuf.MethodOptions {
  int32 ygrpc_cgo_native = 50002; // 0/1
}

service Svc {
  rpc Ping(MyReq) returns (MyResp) {
    option (ygrpc_cgo_native) = 1; // 禁用 native
  }
}
```

### 7) 重要说明（当前限制）

当前生成的导出函数仍是“可编译的最小骨架（scaffold）”：函数体不包含真实的 marshal/unmarshal 与业务 handler 调用逻辑，主要用于固定 ABI 形态、验证 cgo/buildmode=c-shared 可用与头文件原型符合约定。
