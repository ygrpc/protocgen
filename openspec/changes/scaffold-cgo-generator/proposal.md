# 脚手架 CGO 生成器 (Scaffold CGO Generator)

## 目标 (Goal)

创建一个新的 `protoc` 插件 `protoc-gen-ygrpc-cgo`，用于将 Go 实现的 RPC 服务导出为 C 接口。

## 能力 (Capabilities)

### 1. 基础 CGO 导出

- 支持 Unary 和 Streaming 调用。
- 支持基于 Protobuf 二进制数据的标准交换模式。

### 2. 原生模式 (Native Mode)

- **需求**: 对于仅包含基本类型（数值、字符串、字节）的扁平消息，支持跳过序列化步骤，直接通过 C 函数参数传递字段值。
- **命名**: 生成后缀为 `_Native` 的接口。
- **场景**: 高频调用的简单接口，提升性能。

### 3. 显式生命周期管理

- 所有堆内存传递必须携带释放函数。

## 变更 ID (Change ID)

`scaffold-cgo-generator`
