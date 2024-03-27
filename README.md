# protocgen

the repo for protoc generators

ygrpc is a low code system base on grpc(protobuf) idl, we adopt protobuf as the base for all info and  communication, and generate all the code for you, so you can focus on the business logic.

ygrpc是一个基于grpc(protobuf) idl的低代码系统，我们采用protobuf作为所有数据和通信的基础，并为您生成所有的代码，这样您就可以专注于业务逻辑。
虽然ygrpc会为您生成大部分代码，但是你还是得需要理解整个流程，我们并不是一个全自动的系统，我们只是帮助您减少重复的工作，提高开发效率，底层的流程还是很清晰的，你可以很容易的理解。

ygrpc的整个流程如下：

- 编写proto文件，定义好数据库表结构
  - 根据proto文件生成sql初始化脚本
  - 根据proto文件生成一些辅助的proto message, 如 message_list, crud_rpc
  - 根据proto文件生成crud操作代码，包括后端go代码，前端ts代码, 和rpc代码

cmd下面的工具说明：

- protoc-gen-ygrpc-sql: 生成sql初始化脚本
- protoc-gen-ygrpc-msglist: 生成protobuf message list
