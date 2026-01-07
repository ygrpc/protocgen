# CGO 互操作支持 (Go-Library 模式)

## ADDED Requirements

### Requirement: Export Go Functions

插件必须 (MUST) 导出 `//export` 函数。

#### Scenario: Unary Export

- **WHEN** proto 定义包含 `rpc SayHello(SayHelloReq) returns (SayHelloResp)`
- **THEN** 生成的 Go 代码包含 `//export Service_SayHello` 并提供可被 C 调用的导出符号。

### Requirement: Binary Mode Support

插件必须 (MUST) 始终生成接受 Protobuf 二进制数据的标准接口。

#### Scenario: Fallback Interface

- **WHEN** 某个 RPC 的 request/response 满足 Native Mode 可用条件
- **THEN** 仍必须同时生成 Binary 接口（接受 protobuf bytes 三元组），不得仅生成 Native 接口。

### Requirement: Native Arguments Support

对于仅仅包含基本类型的扁平消息，插件必须 (MUST) 额外生成一个 "Native" 版本的接口。

扁平消息定义：仅允许数值类标量、`bool`、`string`、`bytes` 字段；且不允许 `optional` / `map` / `enum` / `repeated` / `oneof` / 嵌套 `message`。

#### Scenario: Flat Message Input

- **WHEN** request 为 `message Log { string msg = 1; }`
- **THEN** 生成的 Native 接口签名包含 `(const char* msg, int msg_len, FreeFunc msg_free)`（或等价的 `void*` 指针类型），且函数名后缀为 `_Native`。

#### Scenario: Non-Flat Message Skips Native

- **WHEN** request/response 包含任意 `optional` / `map` / `enum` / `repeated` / `oneof` / 嵌套 `message`
- **THEN** 不得生成 `_Native` 接口，仅生成 Binary 接口。

### Requirement: Explicit Lifecycle ABI

所有引用类型的数据交换必须 (MUST) 携带释放函数参数位置。

#### Scenario: Input FreeFunc Optionality

当 C 向 Go 传递参数 `req` 时，同时传入 `FreeFunc req_free`。

- 如果 `req_free` **不为 NULL**，Go 必须 (MUST) 在不再使用 `req` 时执行 `req_free(req)`。
- 如果 `req_free` **为 NULL**，Go 必须 (MUST) **不执行** 任何释放操作，仅读取数据。这也适用于 Native 模式下的字符串参数。

#### Scenario: len=0 Means Absent Input

- **WHEN** C 传入 `len == 0` 的 `string/bytes` 三元组参数
- **THEN** Go 不得读取 `ptr` 内容，并将该字段视为“未传入”（保持默认值）。
- **AND WHEN** `ptr != NULL` 且 `free != NULL`
- **THEN** Go 必须调用 `free(ptr)` 以避免内存泄漏。

#### Scenario: Output FreeFunc Obligation

当 Go 向 C 返回参数 `out` 时，必须 (MUST) 总是返回有效的 `out_free` 函数指针，C 必须调用它。

### Requirement: Error Reporting

所有导出函数必须 (MUST) 以 `int` 返回错误码，并通过输出参数 `msg_error` 返回错误消息（三元组）。

#### Scenario: Error Code + Error Message

- **WHEN** 导出函数执行成功
- **THEN** 返回值为 `0`，且 `msg_error_len == 0`（或 `msg_error_ptr == NULL`）。
- **WHEN** 导出函数执行失败
- **THEN** 返回值为非 `0`，且 `msg_error` 三元组包含错误消息，并提供可调用的 `msg_error_free`（允许为 `NULL` 仅当 `msg_error_ptr == NULL`）。

### Requirement: Generate C Header

头文件必须 (MUST) 包含 Binary 和 Native (如果可用) 两种接口的原型。

#### Scenario: Header Protos

- **WHEN** 某个 RPC 方法生成导出符号
- **THEN** 头文件必须包含对应的 `Service_Method` 声明。
- **AND WHEN** 该方法满足 Native 条件
- **THEN** 头文件还必须包含 `Service_Method_Native` 声明。
