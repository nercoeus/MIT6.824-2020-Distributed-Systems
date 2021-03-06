# 内容

什么是分布式系统？
* 多台协作的机器
* 大型网站的存储，大数据的计算：MapReduce，p2p 分享
* 分布式的管理多个基础架构

为什么要使用分布式系统？
* 高性能
* 因为物理上的原因不得不使用分布式架构（两家物理机不在一起但是需要一起提供某一服务）
* 通过复制提高可用性（容错性）
* 通过隔离保证可靠性

分布式系统的挑战：
* 分布式系统的构建十分复杂
* 需要应对部分组件的错误
* 难以发挥机器的性能（并不是 1000 台计算机就会有 1000 倍的性能）

Topic：
- 分布式系统的基础组件：RPC通信，线程机制，并发控制 etc
- 高性能：横向扩展，同时也需要实现负载均衡，处理逻辑并发，共享资源瓶颈等等问题
- 容错行：随着计算机数量的升级会频繁的出现程序或者集群节点的崩溃
  - 可用性
  - 可恢复性
- 一致性：弱一致性和强一致性

# 案例研究:mapReduce
## 概述
context：将大数据下的耗时计算操作分散到多台机器上进行计算
总体目标：初级程序员简单的顺序调用 Map 和 Reduce 方法完成数据操作，并发的细节被隐藏掉了
## MapReduce 抽象视图
Input* 是整个计算的输入,可能是大量的待计算数据等. 使用 Map function 来计算大量数据中的一小部分,最后将所有 Map 计算的结果使用 Reduce function 来进行收集并统计结果输出.
```
Input1 -> Map -> a,1 b,1
Input2 -> Map ->     b,1
Input3 -> Map -> a,1     c,1
                  |   |   |
                  |   |   -> Reduce -> c,1
                  |   -----> Reduce -> b,2
                  ---------> Reduce -> a,2
```
讲义中 MapReduce 的例子
通过调用 Map() 为每个出入文件产生一组 k2,v2
并通过 Reduce() 收集上面产生的 k,v 最后输出处理结果
## MapReduce 有很好的横向扩展性
Map()s 和 Reduce()s 方法可以在互不影响的情况下并行运行，所以可以很容易的通过增加机器来提高的吞吐率
## MapReduce 隐藏了很多的细节
* 将任务自动部署到机器集群上进行执行
* 跟踪的任务直至完成
* 控制 Map() 和 Reduce() 之间的数据传输
* 机器间的负载均衡
* 从失败中恢复 recover
## MapReduce 的限制
* Map()s 和 Reduce()s 之间不可以相互作用也没有不同状态
* 没有迭代以及多阶段的管道
* 没有实时或者流式处理,只能根据预先制定好的 Map 和 Reduce 规则进行处理
## 为什么会限制性能
网络和cpu/磁盘会限制性能，在 MapReduce 中网络带宽是很大的阻碍，论文第 8 小结关于网络带宽的优化方法总结如下:
1. 将输入数据保存到不同的机器上并就近原则执行 Map 方法,即 Input 数据和 Map() 在同一台机器上减少网络操作节约性能.
2. 将生成的中间数据的一个副本保存在本地磁盘上，作用和上面的类似，可以就近执行 Reduce 操作.这种并不能避免所有的 Reduce() 的网络带宽.

但是发展到现在，其实网络带宽已经不是主要的问题了。
# 论文

见 Documents 目录下的 MapReduce.pdf 了解更多和 MapReduce 相关知识。

# 实验
![MapReduce 执行流程](https://github.com/nercoeus/MIT-6.824-2020-Distributed-Systems/blob/master/image/mapreduce.jpg)
上图是 MapReduce 的执行流程，这里实验代码实现也尽量参照上面的流程进行实现。实验代码见 src/MapReduce 目录。

自带的例子就不进行展示了：大概流程就是先将一系列的文件通过 Map() 函数进行解析为 {key:name,value:1} 的集合，然后在排序后并使用 Reduce() 统计相同词的频率然后统一写入输出文件即可。

需要我们实现一个分布式的 MapReduce，通过 RPC 进行通信工作进程向主进程请求任务，从一个文件或多个文件中读取输入，并将任务的输出写入到一个文件或多个中。主进程并监控如果从进程在一定时间内没有完成任务就需要切换一个进程执行任务。