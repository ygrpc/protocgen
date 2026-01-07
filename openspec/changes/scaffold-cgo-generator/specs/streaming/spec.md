# 流式支持 (Callback 模式)

## ADDED Requirements

### Requirement: Callback Definitions
插件必须 (MUST) 生成 `OnRead` 和 `OnDone` 回调的 C typedef 定义。

#### Scenario: Callback Types
`typedef void (*OnReadFunc)(void* ctx, char* data, int len, FreeFunc data_free)`。

### Requirement: Server Streaming Export
插件必须 (MUST) 导出用于服务端流式处理的函数，该函数接受请求数据和回调，并在独立的 Goroutine 中执行 Handler。

#### Scenario: Server Push
C 调用 `Service_StreamMethod`。Go 实现通过 `onRead` 推送数据。

### Requirement: Client/Bidi Streaming Export
插件必须 (MUST) 导出用于客户端/双向流式处理的函数，该函数接受回调并返回流句柄。

#### Scenario: Bidi Flow
C 调用 `h = Service_Bidi_Start`。C 调用 `Send(h, data)`。Go 通过 `onRead` 推送数据。
