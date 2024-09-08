# ulib
<!-- TOC tocDepth:2..3 chapterDepth:2..6 -->

- [介绍](#介绍)
- [软件架构](#软件架构)
    - [主框架(ctl)](#主框架ctl)
    - [日志框架(log)](#日志框架log)
    - [日志框架(evn)](#日志框架evn)
    - [http框架(htp)](#http框架htp)
    - [orm框架(mdb)](#orm框架mdb)
    - [cache框架(rdb)](#cache框架rdb)

<!-- /TOC -->
## 介绍

典型的api服务器, 适用于开发stateless服务, 此项目重写note的slot游戏服务器.

## 软件架构

### 主框架(ctl)
采用control微框架,简单,易用,灵活.

### 日志框架(log)
- 支持6个日志等级 [TRC] [DBG] [INF] [WRN] [ERR] [FAL]
- 支持不同级别日志输出不同颜色
- 支持2种输出模式 ELM_Std(控制台) ELM_File(文件流,支持轮换)
- 支持阀值告警
- 支持绑定字段
- 支持对接graylog日志管理平台(gelf-udp)

### 日志框架(evn)


### http框架(htp)
网络框架采用知名的gin包, htp包进一步封装了gin. 
- context
- render
- service
- response
- server
- middleware

### orm框架(mdb)
DB-orm框架采用知名的xorm, mdb对它进行了进一步的封装.更易用,高复用.
- table(实体)
- session(事务)

### cache框架(rdb)
cache框架采用知名的redigo, rdb对它进行了进一步的封装. 更易用,高复用.
- 各种类型的key
- 分布式锁
- 统一的reply
- Pipe & Exec