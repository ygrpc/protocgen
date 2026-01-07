# 脚手架 CGO 生成器 (Scaffold CGO Generator)

## 目标 (Goal)

创建一个新的 `protoc` 插件 `protoc-gen-ygrpc-cgo`，用于将 Go 实现的 gRPC 服务导出为 C 接口。这允许 C/C++ 应用程序直接调用 Go 实现的 RPC 方法（等同于将网络层替换为 CGO 调用），并确保调用链包含 Go 侧配置的中间件（Interceptors）。

## 能力 (Capabilities)

### 1. CGO 接口导出 (Export CGO Interface)

- **需求**: 生成包含 `//export` 的 Go 代码，将 RPC 方法暴露给 C 调用。
- **需求**: 生成 C 头文件，定义调用这些 Go 方法的函数原型。
- **需求**: 支持所有 gRPC 原语（Unary, Server/Client/Bidi Streaming）。

### 2. 中间件集成 (Middleware Integration)

- **需求**: 关键能力。当 C 调用导出函数时，生成的代码必须构造上下文并手动执行 Go 侧配置的 `UnaryServerInterceptor` 或 `StreamServerInterceptor` 链，最后调用实际的 Go 业务逻辑。

### 3. 基于回调的流式处理 (Callback-based Streaming)

- **需求**: 由于 C 是调用发起方，流式交互通过回调实现。
- **方案**:
  - C 调用 `Start` 获得句柄。
  - C 调用 `Send` 发送数据给 Go。
  - Go 通过 `OnRead` 回调将数据推送给 C。

## 用户收益 (User Benefit)

- 允许 C/C++ 遗留系统零网络开销地集成现代 Go gRPC 服务。
- 复用 Go 强大的生态（日志、监控、鉴权中间件）。
- 保持 Protobuf 作为单一事实来源。

## 变更 ID (Change ID)

`scaffold-cgo-generator`
