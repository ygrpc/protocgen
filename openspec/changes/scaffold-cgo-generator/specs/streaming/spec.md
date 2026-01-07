# 流式支持 (Callback 模式)

## ADDED Requirements

### Requirement: Callback Definitions

插件必须 (MUST) 生成 `OnRead` 和 `OnDone` 回调的 C typedef 定义，以支持异步/Reactor 风格的流式处理。

#### Scenario: Callback Types

头文件必须包含 `typedef void (*OnReadFunc)(void* ctx, char* data, int len)`。

### Requirement: Server Streaming Export

插件必须 (MUST) 导出用于服务端流式处理的函数，该函数接受请求数据和回调。

#### Scenario: Server Push

C 调用 `Service_StreamMethod(req, onRead, onDone)`。Go 实现通过 `onRead` 推送数据，并通过 `onDone` 发送完成信号。

### Requirement: Client/Bidi Streaming Export

插件必须 (MUST) 导出用于客户端/双向流式处理的函数，该函数接受回调并返回流句柄 (Stream Handle)。

#### Scenario: Bidi Flow

C 调用 `h = Service_Bidi_Start(onRead, onDone)`。C 调用 `Send(h, data)`。Go 通过 `onRead` 推送接收到的数据。

### Requirement: Thread Safety Warning

生成的文档必须 (MUST) 警告回调是从任意 Go 线程调用的。

#### Scenario: Documentation

头文件或文档必须说明 `OnRead` 必须是线程安全的。
