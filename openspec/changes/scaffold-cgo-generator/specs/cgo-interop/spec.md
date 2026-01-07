# CGO 互操作支持 (Go-Library 模式)

## ADDED Requirements

### Requirement: Export Go Functions

插件必须 (MUST) 为每个输入 proto 文件生成 Go 文件，其中包含允许 C 调用 RPC 逻辑的 `//export` 导出函数。

#### Scenario: Unary Export

给定 `rpc SayHello`，生成的 Go 代码包含 `//export Service_SayHello`。

### Requirement: Explicit Lifecycle ABI

所有数据交换必须 (MUST) 携带对应的释放函数 (FreeFunc)，实现所有权的显式移交。

#### Scenario: Input Lifecycle

当 C 向 Go 传递参数 `req` 时，必须同时传入 `FreeFunc req_free`。Go 必须在不再使用 `req` 时（通常是返回前）执行 `req_free(req)`。

#### Scenario: Output Lifecycle

当 Go 向 C 返回参数 `out` 时，必须同时返回 `FreeFunc out_free`。C 必须在不再使用 `out` 时执行 `out_free(out)`。

### Requirement: Pinned Memory

Go 向 C 返回的内存地址必须 (MUST) 是“固定”的（Pinned/Off-Heap），不受 Go GC 移动的影响。

#### Scenario: C Malloc

Go 实现应使用 `C.malloc` 分配返回内存，并返回对应的 `free` 包装函数作为 `out_free`。

### Requirement: Generate C Header

插件必须 (MUST) 生成 C 头文件。

#### Scenario: Prototype

`int Func(const char* req, int req_len, FreeFunc req_free, char** out, int* out_len, FreeFunc* out_free);`
