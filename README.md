# heartbeating
为了实现一个在线更新客户端软件的服务器，同时考虑流量成本，于是决定采用socket长连接的方式维持客户端与服务端之间的通信  
因此需要实现一个简单且稳定的长连接工具包  
此工具包会分为客户端与服务端  


编写此包将在虚拟宿主机内进行，宿主机会直接来git clone当前远程仓库，并开始开发  
重点需要注意的是客户端与服务端同步开发，如何最大程度做到简洁且不臃肿  


# 关于加密
为了节省流量确实心跳包所发送的数据确实越少越好，但是必要的加密还是要有的  
初步构思可以采取两种加密模式：  
第一种是md5+时间戳+盐  
第二种是针对有能力实现慢哈希md5+时间戳+盐客户端，相关介绍可看如下文章:  
https://www.cnblogs.com/zhangchengye/p/6323409.html

# 关于日志系统  
使用了  
    github.com/phachon/go-logger  
超级好用，努力做好集成

# 关于测试单元  
这次开始使用testing测试单元进行代码的调试，而不是和以前那样go build了  
过程中遇到了cgo找不到的问题，似乎已经找到了方法

# 关于长连接  
这有篇文章不错，如果之后有更值得借鉴的文章会在这里补充  
https://blog.csdn.net/zhizhengguan/article/details/108026066

# 心跳包只是心跳包  
要明白一件事，那就是心跳包仅仅是某个项目整体套接字逻辑的一个组件  
同样的组件还有客户端管理组件、以及之前所说的加密，其实也应该是属于整体套接字的组件，而不是心跳包的组件  
这里只实现心跳包，其他的都不去实现

