# CGO 互操作支持 (Go-Library 模式)

## ADDED Requirements

### Requirement: Export Go Functions

插件必须 (MUST) 导出 `//export` 函数。

#### Scenario: Unary Export

给定 `rpc SayHello`，生成的 Go 代码包含 `//export Service_SayHello`。

### Requirement: Binary Mode Support

插件必须 (MUST) 始终生成接受 Protobuf 二进制数据的标准接口。

#### Scenario: Fallback Interface

即便 Native Mode 可用，生成的代码也必须保留接受 `reqbuf` 和 `respbuf` 的通用接口。

### Requirement: Native Arguments Support

对于仅仅包含基本类型的扁平消息，插件必须 (MUST) 额外生成一个 "Native" 版本的接口。

#### Scenario: Flat Message Input

给定 `message Log { string msg = 1; }`，Native 接口签名包含 `(const char* msg, int msg_len, FreeFunc msg_free)`。

### Requirement: Explicit Lifecycle ABI

所有引用类型的数据交换必须 (MUST) 携带释放函数参数位置。

#### Scenario: Input FreeFunc Optionality

当 C 向 Go 传递参数 `req` 时，同时传入 `FreeFunc req_free`。

- 如果 `req_free` **不为 NULL**，Go 必须 (MUST) 在不再使用 `req` 时执行 `req_free(req)`。
- 如果 `req_free` **为 NULL**，Go 必须 (MUST) **不执行** 任何释放操作，仅读取数据。这也适用于 Native 模式下的字符串参数。

#### Scenario: Output FreeFunc Obligation

当 Go 向 C 返回参数 `out` 时，必须 (MUST) 总是返回有效的 `out_free` 函数指针，C 必须调用它。

### Requirement: Generate C Header

头文件必须 (MUST) 包含 Binary 和 Native (如果可用) 两种接口的原型。

#### Scenario: Header Protos

头文件必须包含 `Service_Method` 和 `Service_Method_Native` (如果有) 的定义。
