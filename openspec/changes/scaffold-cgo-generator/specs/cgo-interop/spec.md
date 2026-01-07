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
对于仅仅包含基本类型（Scalars, Strings, Bytes）的消息，插件必须 (MUST) 额外生成一个 "Native" 版本的接口，将字段展开为函数参数。

#### Scenario: Flat Message Input
给定 `message Log { string msg = 1; int32 level = 2; }`，Native 接口签名应为 `Func(const char* msg, int msg_len, FreeFunc msg_free, int level, ...)`。接口名称必须以 `_Native` 结尾。

#### Scenario: Flat Message Output
对于上述消息作为返回值，Native 接口应包含输出参数 `char** out_msg, int* out_msg_len, FreeFunc* out_msg_free, int* out_level`。

### Requirement: Explicit Lifecycle ABI
所有引用类型（String/Bytes/Message）的数据交换必须 (MUST) 携带释放函数。

#### Scenario: Input String Full Cycle
当 C 向 Go 传递 String 字段时（无论是 Binary 还是 Native 模式），必须传入 `(ptr, len, free)`。Go 必须使用 `GoStringN` 读取数据，并在使用后调用 `free(ptr)`。

### Requirement: Generate C Header
头文件必须 (MUST) 包含 Binary 和 Native (如果可用) 两种接口的原型。

#### Scenario: Header Protos
头文件必须包含 `Service_Method` (二进) 和 `Service_Method_Native` (原生) 的定义。
