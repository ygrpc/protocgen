# 设计：RPC CGO 生成器 (Design: CGO Generator for RPC)

## 概述 (Overview)

`protoc-gen-ygrpc-cgo` 插件生成一个 Go 适配层，将 Go 实现的 RPC 服务导出给 C 使用。

## 架构 (Architecture)

```mermaid
graph LR
    CApp(C Application) -->|Call| GoExport[Go Exported Function]
    GoExport -->|Native Args / Proto Bytes| GoHandler[Go Service Implementation]
    GoHandler -->|Return| GoExport
    GoExport -->|Return| CApp
```

## 数据传递与内存生命周期 (ABI & Lifecycle)

### 1. 核心 ABI 原则 (Core ABI Principles)
- **显式生命周期 (Explicit Lifecycle)**:
    - **C -> Go**: 必须传输 `(DataPtr, Length, FreeCallback)`。Go 在使用完数据后（通常在函数返回前）调用 `FreeCallback`。
    - **Go -> C**: 必须传输 `(DataPtr, Length, FreeCallback)`。Go 使用 `C.malloc` 分配 Pinned Memory，并返回标准 `free` 的包装函数。C 在使用完数据后调用 `FreeCallback`。
- **字符串编码 (String Encoding)**:
    - 所有的 `char*` 必须是 UTF-8 编码。
    - 禁止仅依赖 Null-Terminator，必须显式传递长度。

### 2. 模式 A：二进制模式 (Binary Mode) - 默认
- **适用场景**: 复杂消息对象，嵌套结构。
- **签名**: 接收序列化的 Protobuf 二进制 `(req_buf, req_len, req_free)`，返回 `(resp_buf, resp_len, resp_free)`。

### 3. 模式 B：原生模式 (Native Mode)
- **命名**: 接口函数名后缀为 `_Native` (例如 `Service_Method_Native`)。
- **适用场景**: Request/Response 仅包含基本类型（int, long, double, bool, string, bytes）且无嵌套 Message 的 RPC 方法。
- **映射规则**:
    - **基本数值**: 直接映射 (`int32` -> `int`, `int64` -> `long long`, `double` -> `double`)。
    - **String/Bytes**: 展开为三元组 `(char* ptr, int len, FreeFunc free)`。
- **实现逻辑**:
    - Go 导出函数接收展开后的参数。
    - Go 内部构造 Go Struct。
    - 调用业务 Handler。
    - 返回值同样展开。

**接口示例:**
假设 `rpc Login(LoginReq) returns (LoginResp)`
`message LoginReq { string user = 1; int32 age = 2; }`
`message LoginResp { int32 code = 1; string msg = 2; }`

```c
int MyService_Login_Native(
    // Input Fields
    const char* user, int user_len, FreeFunc user_free, // String 展开为三元组
    int age,                                            // Int 直接传递

    // Output Fields
    int* code,                                          // Int 输出
    char** msg, int* msg_len, FreeFunc* msg_free        // String 输出展开为三元组
);
```

### 4. 流式定义
流式接口也支持 Native 模式，`OnRead` 回调将展开参数列表。

## 内存管理细节 (Memory Management Details)
对于 `string` 类型的字段，无论是 Binary Mode 还是 Native Mode，都严禁简化处理：
1.  **Input String**: C 传入 `(ptr, len, free)`。Go 使用 `C.GoStringN(ptr, len)` 创建 Go String（发生拷贝）。Go 随即调用 `free(ptr)`。
2.  **Output String**: Go 计算结果。Go 调用 `C.malloc(len)` 分配内存。Go 将结果拷贝入 `malloc` 的内存。Go 返回 `(malloc_ptr, len, wrap_free)`。
