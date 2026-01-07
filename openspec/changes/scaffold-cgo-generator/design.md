# 设计：RPC CGO 生成器 (Design: CGO Generator for RPC)

## 概述 (Overview)

`protoc-gen-ygrpc-cgo` 插件生成一个 Go 适配层，将 Go 实现的 RPC 服务导出给 C 使用。

## 架构 (Architecture)

```mermaid
graph LR
    CApp(C Application) -->|Call + FreeCallback| GoExport[Go Exported Function]
    GoExport -->|Run| Middleware[Go Interceptors]
    Middleware -->|Call| GoHandler[Go Service Implementation]
    GoHandler -->|Result| Middleware
    Middleware -->|Return Ptr + FreeFunc| GoExport
    GoExport -->|Return| CApp
```

## 数据传递与内存生命周期 (ABI & Lifecycle)

### 1. 核心原则：显式生命周期管理

为了实现对内存的绝对控制，所有跨语言的内存传递都必须携带“销毁器 (Destructor)”。

### 2. C -> Go (Input String/Bytes)

**规则**:

- C 传入数据指针 (`val`) 和长度 (`len`)。
- C **必须** 同时传入一个释放函数 (`free_func`)。
- Go 在使用完该数据（通常是函数返回前，或异步调用结束后）**必须** 调用该 `free_func`。

**接口定义**:

```c
typedef void (*FreeFunc)(void* ptr);

int MyService_MyUnary(
    // Input
    const char* req, int req_len,
    FreeFunc req_free,            // C 提供的释放函数，Go 用完 req 后调用它

    // Output
    char** out, int* out_len,
    FreeFunc* out_free            // Go 返回的释放函数，C 用完 out 后调用它
);
```

### 3. Go -> C (Output String/Bytes)

**规则**:

- Go 分配内存（必须固定/Pinned，通常使用 `C.malloc` 分配堆外内存以避开 GC 移动）。
- Go **必须** 返回一个专门用于释放该内存的函数指针 (`out_free`) 给 C。
- C 在使用完数据后，调用 `out_free(out)`。

**Go 实现逻辑**:

```go
//export MyService_MyUnary
func MyService_MyUnary(
    reqBuf *C.char, reqLen C.int, reqFree C.FreeFunc,     // IN
    outBuf **C.char, outLen *C.int, outFree *C.FreeFunc,  // OUT
) C.int {
    // 1. 处理输入
    // 使用 reqBuf...
    // 关键：确保在不再需要 reqBuf 后调用释放
    defer C.call_free_func(reqFree, reqBuf)

    // 2. 业务逻辑...

    // 3. 处理输出
    // Go 使用 C.malloc 分配 Pinned Memory
    *outBuf = C.malloc(respLen)
    copy(*outBuf, respData)
    *outLen = respLen

    // 返回标准释放函数 (wrapper for free)
    *outFree = C.get_standard_free_func()

    return 0
}
```

### 4. 流式定义 (基于回调)

**回调定义**:
`OnRead` 回调同样遵循此规则：

```c
// Go 推送数据给 C
// Go 提供 data, len, 以及 data_free 函数。
// C 处理完 data 后，必须调用 data_free(data)。
typedef void (*OnReadFunc)(void* ctx, char* data, int len, FreeFunc data_free);
```

## 字符串编码

- 依旧强制 UTF-8。

## 内存安全总结

- **Input**: Go 负责 Driver (C) 传入内存的生命周期结束（调用 C 提供的 Free）。
- **Output**: C 负责 Service (Go) 返回内存的生命周期结束（调用 Go 提供的 Free）。
