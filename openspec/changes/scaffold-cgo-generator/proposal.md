# 脚手架 CGO 生成器 (Scaffold CGO Generator)

## Why

需要一个新的 `protoc` 插件 `protoc-gen-ygrpc-cgo`，将 Go 实现的 RPC 服务以稳定的 C ABI 导出，便于非 Go 进程/语言集成。

## What Changes

- 新增 `cmd/protoc-gen-ygrpc-cgo`：实现标准 `protoc` 插件入口。
- 生成两类接口：
	- **Binary Mode**（默认，必有）：通过 Protobuf 二进制进行请求/响应交换。
	- **Native Mode**（可选，按消息类型判定）：对“扁平消息”展开字段，跳过序列化；并支持通过 method option 显式关闭某个方法的 native 接口生成。
- 支持 Unary 与 Streaming（server streaming + client/bidi streaming），两者均提供 Binary + Native 两版本（Native 仅在可用时生成）。
- ABI 明确生命周期：涉及堆内存的指针参数必须携带释放函数指针；允许输入 free 为空指针。
- ABI 支持 request 内存释放策略可配置：默认情况下导出函数签名不包含 request 的 `free` 参数；可通过 message option 强制包含；或生成双版本。
- 两个 option（request free 策略、native 生成开关）支持两种配置方案：
	- **FileOptions（整个 proto 文件级）**：作为默认值（由于 protobuf 扩展名全局唯一，建议使用 `*_default` 命名）
	- **MethodOptions（方法级）**：覆盖 FileOptions（Method 优先；建议使用 `*_method` 命名）
	- request free 策略仅支持 FileOptions/MethodOptions 两级配置（MethodOptions > FileOptions）。
- 统一错误模型：导出函数返回 `int`（0=成功；非 0=错误 ID）；错误消息不再通过函数签名输出参数返回，而是通过全局导出函数 `GetErrorMsg(error_id, ptr,len,free)` 获取（Go 侧维护全局 map，保存 3s）。

### Options（字段号约定）

- `ygrpc_cgo_req_free`：`0/1/2`（与现有策略一致），字段号 `50001`
- `ygrpc_cgo_native`：`0/1`（0 默认生成；1 禁用 native），字段号 `50002`

建议的扩展声明（命名示例，字段号保持一致）：

- `extend google.protobuf.FileOptions   { int32 ygrpc_cgo_req_free_default = 50001; int32 ygrpc_cgo_native_default = 50002; }`
- `extend google.protobuf.MethodOptions { int32 ygrpc_cgo_req_free_method = 50001; int32 ygrpc_cgo_native = 50002; }`

### ABI 命名约定（参数前缀）

- Binary 接口：request 侧参数使用 `in*` 前缀，response 侧参数使用 `out*` 前缀；并且参数名必须包含对应的 **消息类型名** 以提升可读性。
	- request bytes：`in<ReqMsg>Ptr / in<ReqMsg>Len [/ in<ReqMsg>Free]`
	- response bytes：`out<RespMsg>Ptr / out<RespMsg>Len / out<RespMsg>Free`

示例：`rpc Ping(PingRequest) returns (PingResponse)`
	- 默认 Binary：`inPingRequestPtr, inPingRequestLen, outPingResponsePtr, outPingResponseLen, outPingResponseFree`
	- `*_TakeReq` Binary：额外包含 `inPingRequestFree`

## Definitions

### Flat Message（Native Mode 判定）

仅支持 Go/Protobuf 的基本标量字段：数值（各类 int/uint/sint/fixed）、`bool`、`string`、`bytes`。

不支持（遇到则 **不生成** Native 接口，仅生成 Binary 接口）：`enum`、`optional`、`repeated`、`map`、`oneof`、任何嵌套 `message`。

## Impact

- Affected change specs:
	- `cgo-interop`
	- `streaming`
- Affected code:
	- 新增 `cmd/protoc-gen-ygrpc-cgo/main.go`
	- 复用或扩展现有 `protocplugin/` 的通用能力（按需要最小化修改）

## Change ID

`scaffold-cgo-generator`
