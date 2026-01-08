## 1. Scaffolding & Plumbing

- [x] 1.1 新增插件入口目录与最小可运行框架
    - 交付物：`cmd/protoc-gen-ygrpc-cgo/main.go` 能作为标准 `protoc` 插件从 stdin 读 request、向 stdout 写 response。
    - 验收点：`go test ./...`（或至少 `go test ./cmd/...`）通过；手工运行 `protoc --ygrpc-cgo_out=.` 能产出空/最小文件而不崩溃。

## 2. Binary Mode（Unary）

- [x] 2.1 定义并生成基础 ABI 类型（FreeFunc / Buf 三元组）
    - 交付物：生成的 Go 文件包含 `import "C"` 的 cgo 注释块，内含 `typedef void (*FreeFunc)(void*);` 及相关声明（不使用 C struct，三元组通过参数列表表达）。
    - 验收点：生成代码 `go test` 可编译（至少到 cgo 语法层面，不要求链接到真实 C 程序）。

- [x] 2.2 生成 Unary 的 Binary 接口（必有）
    - 交付物：每个 unary rpc 生成一个导出函数，request 默认使用 `(ptr, len)` 输入，response 使用 `(ptr, len, free)` 输出。
    - 验收点：对 sample proto 编译生成的 cgo 头文件（buildmode=c-shared 自动产出）/Go 导出函数名与参数满足 change specs（`cgo-interop`）。

- [x] 2.3 统一错误模型（Binary）：ErrorId + GetErrorMsg
    - 交付物：导出函数返回 `int`（0=成功；非 0=errorId）；不再在函数签名中输出错误消息；额外生成 `Ygrpc_GetErrorMsg(errorId, ptr,len,free)`。
    - 验收点：生成的 C 原型不再出现 `msg_error` 输出参数；存在 `Ygrpc_GetErrorMsg` 原型；文档/注释说明 errorId 的 3s 有效期。

- [x] 2.4 Request Free 参数策略（Binary）
    - 交付物：默认导出函数签名不包含 request `free`；支持通过 file/method option 控制 request free 策略（method 优先）：option=0 默认；option=1 仅生成 `_TakeReq`；option=2 同时生成默认名 + `_TakeReq`。
    - 验收点：样例 proto 覆盖 option=0/1/2；生成物符号命名与参数形态匹配变更规格。

## 3. Native Mode（Unary）

- [x] 3.1 Flat Message 判定器
    - 交付物：实现“仅基本标量字段、且不含 optional/map/enum/repeated/oneof/嵌套 message”的判定。
    - 验收点：sample proto 同时包含支持与不支持的 message；支持的 rpc 生成 `_Native`，不支持的不生成。

- [x] 3.2 Native 接口签名生成（含 string/bytes 三元组）
    - 交付物：对 flat rpc 生成 `*_Native` 导出函数（可通过 method/file option 禁用 native：option=0 默认生成；option=1 不生成，method 优先）；response 侧 string/bytes 仍展开为 `(ptr, len, free)`；request 侧默认展开为 `(ptr, len)`，并受 file/method option 控制 request free 策略（method 优先）：option=0 默认；option=1 仅生成 `*_Native_TakeReq`；option=2 同时生成 `*_Native` + `*_Native_TakeReq`。
    - 验收点：生成的 C 原型字段顺序与形态满足 change specs（`cgo-interop`）。

- [ ] 3.3 Native 的 Go 侧装配逻辑
    - 交付物：Go 侧直接构造 request struct、读取 response struct（替代 marshal/unmarshal）。
    - 验收点：至少对一条 unary rpc 的 generated code 走通编译与基础单测（如有）。

## 4. Streaming（Binary + Native）

- [ ] 4.1 生成 Streaming 的回调 typedef 与句柄类型
    - 交付物：头文件包含 `OnRead` / `OnDone` 等回调 typedef；包含 stream handle 的不透明类型/约定。
    - 验收点：原型满足 change specs（`streaming`），且 `FreeFunc` 在这些 typedef 之前已定义。

- [ ] 4.2 Server streaming：Binary 版本
    - 交付物：导出函数接收请求数据与回调，在 goroutine 内执行并通过 onRead 推送；错误返回改为 errorId + GetErrorMsg。
    - 验收点：输出侧回调参数包含 `(ptr,len,free)` 三元组；导出函数不包含 `msg_error` 输出参数。

- [ ] 4.3 Server streaming：Native 版本（flat 可用时生成）
    - 交付物：`*_Native` 版本导出函数按 native 展开；request free 默认不生成并受 option 控制（0/1/2）；option=2 生成双版本；错误返回改为 errorId + GetErrorMsg。
    - 验收点：对 flat rpc 生成 native streaming 原型，且无 `msg_error` 输出参数。

- [ ] 4.4 Client/Bidi streaming：Binary + Native 版本
    - 交付物：Start 返回句柄；提供 Send/Close/Cancel/Free 等操作；Native 版本按规则生成。
    - 验收点：原型与生命周期要求写进生成的头文件注释/README 片段。

## 5. Verification（Sample Protos）

- [ ] 5.1 新增 sample proto 覆盖 unary + streaming、flat + non-flat、request free option=0/1/2
    - 验收点：能用 `protoc` 触发生成并产出可检视的 C 头文件。

- [ ] 5.2 最小集成验证（可先仅编译级）
    - 验收点：`go test ./...` 通过；至少检查生成物包含所需符号与参数形态。
