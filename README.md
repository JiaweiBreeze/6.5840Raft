# RAFT 6.5840 实验

本仓库包含了我在 6.5840 课程中完成的RAFT（故障容错复制算法）实验部分的实现。
## 概述

RAFT算法是一种共识算法，旨在管理分布式系统中的复制日志。它提供了一种在分布式节点之间实现故障容错和一致性的方法。本仓库包含了在 6.5840 课程作业中实现和实验的RAFT算法。



## 功能

- 领导者选举
- 日志复制
- 安全性和故障容错机制
- 客户端交互以进行日志更新
- RAFT基本测试框架

## 文件结构
- src/ 源码目录
-   src/raft/raft.go raft主流程代码
-   src/raft/append_entries.go 日志复制
-   src/raft/election.go 领导选举
-   src/raft/test_test.go 测试用例

## 安装与本地运行

git clone

go test -run 3A
go test -run 3B (关于rpc数量太多的测试用例会失败)

go test -run 3C