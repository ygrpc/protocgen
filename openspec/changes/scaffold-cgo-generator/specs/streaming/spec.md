# 流式支持 (Callback 模式)

## ADDED Requirements

### Requirement: Callback Definitions
插件必须 (MUST) 生成 `OnRead` 和 `OnDone` 回调的 C typedef 定义。

#### Scenario: Callback Types

- **WHEN** 生成 streaming 相关头文件内容
- **THEN** 必须包含形如 `typedef void (*OnReadFunc)(void* ctx, char* data, int len, FreeFunc data_free);` 的定义。
- **AND** `FreeFunc` 的 typedef 必须出现在所有使用 `FreeFunc` 的回调 typedef 之前。

### Requirement: Streaming Has Binary And Native Variants

Streaming 必须 (MUST) 支持 Binary 与 Native 两个版本：

- **Binary streaming**：通过 `(ptr,len,free)` 传递 protobuf bytes。
- **Native streaming**：当消息满足 flat 限制时，额外生成 `_Native` 版本并展开字段参数。

#### Scenario: Binary Always Exists

- **WHEN** proto 中存在 streaming RPC
- **THEN** 必须生成 Binary streaming 导出函数与对应头文件原型。

#### Scenario: Native Only For Flat

- **WHEN** streaming RPC 的 request/response 满足 flat 限制
- **THEN** 必须额外生成 `_Native` 版本。
- **WHEN** 不满足 flat 限制
- **THEN** 不得生成 `_Native` 版本。

### Requirement: Server Streaming Export
插件必须 (MUST) 导出用于服务端流式处理的函数，该函数接受请求数据和回调，并在独立的 Goroutine 中执行 Handler。

#### Scenario: Server Push

- **WHEN** C 调用 `Service_StreamMethod`（Binary 版本）
- **THEN** Go 在 goroutine 中执行 handler，并通过 `onRead` 回调推送每条响应消息的 `(data,len,free)`。
- **AND** 在 stream 结束时必须调用 `onDone`（成功或失败）。

### Requirement: Client/Bidi Streaming Export
插件必须 (MUST) 导出用于客户端/双向流式处理的函数，该函数接受回调并返回流句柄。

#### Scenario: Bidi Flow

- **WHEN** C 调用 `h = Service_Bidi_Start`（Binary 版本）
- **THEN** 返回一个可用于后续操作的 stream 句柄。
- **AND WHEN** C 调用 `Service_Bidi_Send(h, data, len, free)`
- **THEN** Go 将该消息发送到 handler。
- **AND WHEN** handler 推送响应消息
- **THEN** Go 通过 `onRead(ctx, data, len, free)` 回调推送。

### Requirement: Stream Handle Lifecycle

插件必须 (MUST) 为 streaming 句柄提供显式生命周期管理入口。

#### Scenario: Cancel And Free

- **WHEN** C 调用 cancel/close
- **THEN** Go 必须尽快停止该 stream 的处理，并最终调用 `onDone`。
- **AND WHEN** C 调用 free/destroy
- **THEN** 释放句柄相关资源（重复调用必须是安全的或明确禁止并在文档中说明）。

### Requirement: Streaming Error Reporting

streaming 相关导出函数必须 (MUST) 以 `int` 返回错误结果：`0` 表示成功；非 `0` 表示失败并作为全局唯一的 errorId。错误信息通过全局 `Ygrpc_GetErrorMsg` 获取（不通过函数签名输出参数返回）。

#### Scenario: Start Fails

- **WHEN** `Start` 无法创建 stream
- **THEN** 返回非 0 errorId，并可在 3s 内通过 `Ygrpc_GetErrorMsg(errorId, ...)` 获取错误信息。
