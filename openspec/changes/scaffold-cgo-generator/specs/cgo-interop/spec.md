# CGO 互操作支持 (Go-Library 模式)

## ADDED Requirements

### Requirement: Export Go Functions

插件必须 (MUST) 为每个输入 proto 文件生成 Go 文件，其中包含允许 C 调用 RPC 逻辑的 `//export` 导出函数。

#### Scenario: Unary Export

给定 `rpc SayHello`，生成的 Go 代码必须包含 `//export Service_SayHello`，该函数同步接收请求字节并返回响应字节/错误。

### Requirement: Generate C Header

插件必须 (MUST) 生成定义了导出 Go 函数原型和辅助类型的 C 头文件。

#### Scenario: Header Content

头文件 `service.h` 中必须包含与 CGO 导出匹配的 `extern int Service_SayHello(...)` 原型。

### Requirement: Middleware Execution

导出的函数必须 (MUST) 在调用服务实现逻辑之前执行配置的 Go gRPC 拦截器链。

#### Scenario: Auth Interceptor

当 C 调用 `SayHello` 时，Go Auth Interceptor 必须运行，检查凭据（如果通过 Context/MD 传递），只有通过后才调用实现。
