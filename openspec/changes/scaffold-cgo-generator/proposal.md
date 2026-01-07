# 脚手架 CGO 生成器 (Scaffold CGO Generator)

## 目标 (Goal)
创建一个新的 `protoc` 插件 `protoc-gen-ygrpc-cgo`，用于将 Go 实现的 RPC 服务（支持 gRPC 或 ConnectRPC 等）导出为 C 接口。这允许 C/C++ 应用程序直接调用 Go 实现的 RPC 方法（将网络层替换为 CGO 调用）。

## 能力 (Capabilities)

### 1. CGO 接口导出 (Export CGO Interface)
- **需求**: 生成包含 `//export` 的 Go 代码，将 RPC 方法暴露给 C 调用。
- **需求**: 生成 C 头文件，定义调用这些 Go 方法的函数原型。
- **需求**: 支持所有 RPC 原语（Unary, Server/Client/Bidi Streaming）。

### 2. 显式生命周期管理 (Explicit Lifecycle Management)
- **需求**: 所有跨语言的数据传递（String/Bytes）必须携带释放函数 (`FreeFunc`)。
- **需求**: C -> Go：Go 负责调用 C 提供的释放函数。
- **需求**: Go -> C：Go 返回用于释放其分配内存的函数指针。

### 3. 基于回调的流式处理 (Callback-based Streaming)
- **需求**: 由于 C 是调用发起方，流式交互通过回调实现。
- **方案**:
    - C 调用 `Start` 获得句柄。
    - C 调用 `Send` 发送数据给 Go。
    - Go 通过 `OnRead` 回调将数据推送给 C。

## 用户收益 (User Benefit)
- 允许 C/C++ 遗留系统零网络开销地集成现代 Go RPC 服务。
- 严格的内存生命周期管理，防止泄漏和 GC 问题。
- 保持 Protobuf 作为单一事实来源。

## 变更 ID (Change ID)
`scaffold-cgo-generator`
