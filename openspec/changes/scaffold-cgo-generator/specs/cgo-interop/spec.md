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

#### Scenario: Request FreeFunc Default Omitted

- **WHEN** 生成 unary 或 streaming 的导出函数签名（Binary 或 Native）
- **THEN** 默认情况下 **不得**在函数签名中包含 request 的 `FreeFunc` 参数（包括 Binary 模式下的 `req_free`，以及 Native 模式下 string/bytes 参数对应的 `*_free`）。
- **AND THEN** Go 必须 (MUST) **不执行**任何 request 释放操作，仅读取数据。

#### Scenario: Request FreeFunc Via Message Option

- **GIVEN** request message 声明了自定义 option（详见下方 “Request Free Option”）
- **WHEN** 生成以该 message 作为 request 的导出函数签名
- **THEN** 生成器必须 (MUST) 按 option 值决定是否包含 request `FreeFunc`，并在包含时遵守：
	- 如果 `req_free`（或 `*_free`）**不为 NULL**，Go 必须 (MUST) 在不再使用对应 request 内存时调用该 free。
	- 如果 `req_free`（或 `*_free`）**为 NULL**，Go 必须 (MUST) 不执行释放。

#### Scenario: Dual Variant Generation

- **GIVEN** request message option=2
- **WHEN** 生成以该 message 为 request 的导出函数
- **THEN** 必须 (MUST) 同时生成两套导出符号：
	- 默认符号：不包含 request `free`
	- `*_TakeReq` 符号：包含 request `free`（表示将 request 内存所有权交给 Go，由 Go 负责在合适时机释放）

#### Scenario: len=0 Means Absent Input

- **WHEN** C 传入 `len == 0` 的 `string/bytes` 三元组参数
- **THEN** Go 不得读取 `ptr` 内容，并将该字段视为“未传入”（保持默认值）。
- **AND WHEN** `ptr != NULL` 且 `free != NULL`
- **THEN** Go 必须调用 `free(ptr)` 以避免内存泄漏。

#### Scenario: Output FreeFunc Obligation

当 Go 向 C 返回参数 `out` 时，必须 (MUST) 总是返回有效的 `out_free` 函数指针，C 必须调用它。

### Requirement: Error Reporting (ErrorId + GetErrorMsg)

所有导出函数必须 (MUST) 以 `int` 返回错误结果：`0` 表示成功，非 `0` 表示失败并作为全局唯一的 **errorId**。

#### Scenario: Export Function Returns ErrorId

- **WHEN** 导出函数执行成功
- **THEN** 返回值必须为 `0`。
- **WHEN** 导出函数执行失败
- **THEN** 返回值必须为非 `0` 的 errorId。

### Requirement: GetErrorMsg Export

生成器必须 (MUST) 生成一个全局导出函数用于根据 errorId 获取错误消息（三元组）。

#### Scenario: GetErrorMsg ABI

- **WHEN** 生成任意导出符号
- **THEN** 同一生成产物中必须存在如下（或等价）导出函数原型：

```c
int Ygrpc_GetErrorMsg(int error_id, void** msg_ptr, int* msg_len, FreeFunc* msg_free);
```

- **AND THEN** `msg_ptr/msg_len/msg_free` 必须遵守 `(ptr,len,free)` 生命周期约定。

#### Scenario: GetErrorMsg TTL

- **GIVEN** 某次导出函数失败并返回 errorId
- **WHEN** 在 3s 内调用 `Ygrpc_GetErrorMsg(errorId, ...)`
- **THEN** 必须可获取到对应错误消息。
- **WHEN** 超过 3s 后调用
- **THEN** 必须返回 `1` 表示“未找到/已过期”（实现可清理记录）。

### Requirement: Request Free Option

生成器必须 (MUST) 支持在 proto message 上声明一个 option，用于控制该 message 作为 request 时是否需要在导出函数签名中包含 request `free`。

#### Scenario: Option Semantics

- **WHEN** option=0（默认）
- **THEN** request 不生成 `free` 参数。
- **WHEN** option=1
- **THEN** 仅生成 `*_TakeReq` 导出符号（包含 request `free`），默认符号不生成。
- **WHEN** option=2
- **THEN** 必须同时生成两套导出符号（默认 + `_TakeReq`）。
- **AND THEN** response 侧仍必须始终提供可调用的 `out_free`（不受该 option 影响）。

### Requirement: Generate C Header

头文件必须 (MUST) 包含 Binary 和 Native (如果可用) 两种接口的原型。

#### Scenario: Header Protos

- **WHEN** 某个 RPC 方法生成导出符号
- **THEN** 头文件必须包含对应的 `Service_Method` 声明。
- **AND WHEN** 该方法满足 Native 条件
- **THEN** 头文件还必须包含 `Service_Method_Native` 声明。
