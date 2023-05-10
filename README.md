# EIM(Elegant Instant Messenger)

## 项目描述
本项目是一个采用**Golang**语言编写的**即时通讯IM系统**. 提供单聊, 群聊, 离线消息同
步功能. SDK暂时未编写, 所以无法直接使用服务.

ps: 本项目参考了稀土掘金小册《**分布式IM原理与实战: 从0到1打造即时通讯云**》, 
仅用于学习.

## 技术选型
1. 连接协议: websocket、tcp
2. 序列化: protobuf
3. 分布式唯一id: 雪花算法
4. 注册中心: consul
5. 日志: logrus
6. 数据库: mysql
7. 缓存: redis
8. web框架: iris

## TODO
- [x] 基础层
    - [x] 逻辑协议定义
    - [x] 基础协议定义(用于心跳)
    - [x] jwt
    - [x] 日志logger
- [x] 通信层
  - [x] websocket服务端及客户端实现
  - [x] tcp服务端及客户端实现
- [x] 容器层
- [x] 业务层
  - [x] 网关
  - [x] 登录服务
  - [x] 单聊及群聊逻辑
  - [x] 离线消息同步
  - [x] 群管理
  - [x] 消息管理
- [ ] 测试
  - [ ] mock测试
  - [ ] benchmark测试
  - [ ] 集成测试echo
- [ ] docker部署

## 目录结构
```
-EIM
    |-container 容器层
    |-examples 测试
        |-mock mock测试
    |-logger 日志
    |-naming 注册中心
        |-consul consul接口
    |-services 业务服务
        |-gateway 网关
        |-router 路由
        |-server 各种服务
        |-service 群管理及消息管理
    |-storage redis缓存
    |-tcp tcp连接实现
    |-websocket websocket连接实现
    |-wire 基础层
        |-pkt 基础层协议定义
        |-proto proto文件
        |-token jwt
    |-channel.go channel实现
    |-channels.go channel的map管理
    |-context.go 自定义上下文
    |-dispatcher.go 调度器
    |-location.go 用户抽象地址
    |-router.go 指令路由
    |-server.go 接口文件
    |-storage.go 会话管理
```

## 使用说明
### 1. 前期准备(安装中间件)
因为docker部署暂未制作, 所以请读者自行将mysql, redis以及consul环境安装到本
地并启动.

### 2. 数据库表
```shell
$ create database eim_base default character set utf8mb4 collate utf8mb4_unicode_ci;
$ create database eim_message default character set utf8mb4 collate utf8mb4_unicode_ci;
```

### 3. 启动服务
```shell
$ cd ./services
$ go run main.go gateway
$ go run main.go server
$ go run main.go royal
```

### 4. 访问Consul即可查看服务的启动状态：
http://localhost:8500/ui

## 未来展望
可尝试加入传输语音, 图片, 视频等功能. 