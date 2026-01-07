# 脚手架 CGO 生成器 (Scaffold CGO Generator)

## Why

需要一个新的 `protoc` 插件 `protoc-gen-ygrpc-cgo`，将 Go 实现的 RPC 服务以稳定的 C ABI 导出，便于非 Go 进程/语言集成。

## What Changes

- 新增 `cmd/protoc-gen-ygrpc-cgo`：实现标准 `protoc` 插件入口。
- 生成两类接口：
	- **Binary Mode**（默认，必有）：通过 Protobuf 二进制进行请求/响应交换。
	- **Native Mode**（可选，按消息类型判定）：对“扁平消息”展开字段，跳过序列化。
- 支持 Unary 与 Streaming（server streaming + client/bidi streaming），两者均提供 Binary + Native 两版本（Native 仅在可用时生成）。
- ABI 明确生命周期：涉及堆内存的指针参数必须携带释放函数指针；允许输入 free 为空指针。
- 统一错误模型：函数返回值为错误码；错误信息通过 `msg_error`（三元组）输出参数回传。

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
