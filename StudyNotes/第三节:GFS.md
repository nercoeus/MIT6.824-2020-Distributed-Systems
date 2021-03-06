# 内容

## 为什么读 GFS？
1. 分布式存储是分布式系统十分重要的使用方向
    - 怎么设计接口和语义
    - 内部应该怎么工作
2. 和课程主题相关：高性能，容错，复制(从服务器)，一致性
3. 好的系统论文：详细描述了 apps 到 net 之间的细节
4. 成功的实现了这一设计并应用到工业界

## 为什么分布式存储很难实现？ 可以参考 CAP 理论
1. 高性能 -> 在多个服务器之间共享数据 : 高性能就需要多台计算器,所以不同机器之间需要共享数据
2. 多服务器 -> 不间断的单点故障  : 机器一多,不可避免会有大量的故障
3. 频繁的节点故障 -> 复制  : 通过复制错误的机器的数据到正常机器上进行替换达到容错的目的
4. 复制 -> 不能很好地满足一致性  : 使用复制就会遇到相同内容不同机器不一致的问题
5. 好的一致性 -> 较差的性能  : 为了严格控制这种一致性就会降低性能

## 保证一致性？
理想情况下：
1. 分布式的服务就好像在同一个服务中一样,调用方并不需要考虑过多东西
2. 服务器使用磁盘进行存储
3. 服务器一次执行一个客户端的操作（即使并发情况下）
4. 会加载历史数据，当服务器崩溃重新启动后, GFS 应该仅仅会 master 节点加载元数据


## 通过从服务器来保证容错性很难保证强一致性
一个简单但是弱一致性的方案：两个从服务器，S1 和 S2。客户端会同时向这两个服务器写入，但只会从任意一个服务器读取。
如果客户端发送了 S1 的写入请求，但是在发送 S2 时崩溃，这样二者就不一致了。如果想要保证服务器之间的一致性，需要频繁的进行同步操作，这样的话就会影响性能。
所以需要在一致性和性能之间进行权衡。

## GFS 内容：
1. Google 中许多服务需要一个大型的快速统一的存储系统：MapReduce、网页爬虫/索引、日志存储分析系统、youtube？ 等等
2. 全局(通过单个节点)：任何用户可以读取系统中的任何文件。允许在应用之间共享数据
3. 自动故障恢复
4. 一个数据中心
5. 仅供 Google 的应用和用户进行使用
6. 目的在于大型文件的连续访问、读取或追加. 并不注重小字段的存储

## 整体架构
1. 大量的客户端（数据库，RPC）
2. 一个 master 节点和多个 chunk server 节点. master 来管理所有的 chunk 节点
3. 每个大文件拆分多个为 64MB 大小的块,可以保存在不同的 chunk 节点中
4. 每个块在三台 chunk 节点上进行备份。

## master 节点的元数据
1. 用来管理 chunk 信息的表
2. 用来记录每个 chunk 节点和其上的数据块之间的映射
3. log 文件,用来记录数据的变更,并会生成 checkpoint 用来恢复 master 节点(使用log记录会更加快速,而不是使用 b-tree 会存在寻找插入位置的过程导致较慢,有趣的角度)

## 客户端读取一个文件的步骤
1. client 发送文件名称和偏移量给主服务器 master
2. master 根据偏移量找到第一个需要的数据块的句柄
3. master 把保存了该数据块的 chunk 服务器列表发送给 client
4. client 缓存 chunk 服务器列表,用以多次重复访问
5. client 从最近的 chunk 服务器请求数据块
6. chunk 服务器从磁盘读取文件并返回给 client

## 如何保证 GFS 的强一致性:GFS 并没有实现,成本过于高昂
1. primary 节点需要过滤重复的写入请求,防止数据出现多次
2. 从节点必须正确执行 primary 节点给与的任务
3. 使用两阶段提交保证任务被每个节点成功执行了

## GFS 存在的缺陷
1. 只有一个 master 节点,并且不支持自动回复,出错需要人工干预
2. 由于他可能会执行重复的追加操作,有些客户端不能很好地处理这种情况
3. 单个 master 节点需要处理大量的请求,这是很困难的
# reference
- 谈谈分布式系统的CAP理论:[https://zhuanlan.zhihu.com/p/33999708#:~:text=CAP%E7%90%86%E8%AE%BA%E6%A6%82%E8%BF%B0,%E9%A1%B9%E4%B8%AD%E7%9A%84%E4%B8%A4%E9%A1%B9%E3%80%82](https://zhuanlan.zhihu.com/p/33999708#:~:text=CAP%E7%90%86%E8%AE%BA%E6%A6%82%E8%BF%B0,%E9%A1%B9%E4%B8%AD%E7%9A%84%E4%B8%A4%E9%A1%B9%E3%80%82)
- wiki:CAP theorem:[https://en.wikipedia.org/wiki/CAP_theorem](https://en.wikipedia.org/wiki/CAP_theorem)