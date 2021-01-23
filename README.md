# heartbeating
为了实现一个在线更新客户端软件的服务器，同时考虑流量成本，于是决定采用socket长连接的方式维持客户端与服务端之间的通信  
因此需要实现一个简单且稳定的长连接工具包  
此工具包会分为客户端与服务端  


编写此包将在虚拟宿主机内进行，宿主机会直接来git clone当前远程仓库，并开始开发  
重点需要注意的是客户端与服务端同步开发，如何最大程度做到简洁且不臃肿  


# ~~关于加密~~
~~为了节省流量确实心跳包所发送的数据确实越少越好，但是必要的加密还是要有的  
初步构思可以采取两种加密模式：  
第一种是md5+时间戳+盐  
第二种是针对有能力实现慢哈希md5+时间戳+盐客户端，相关介绍可看如下文章:  
https://www.cnblogs.com/zhangchengye/p/6323409.html~~   
加密会有独立的加密单元

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
这里有另一篇实现心跳的文章:  
https://my.oschina.net/sharelinux/blog/699725  
两篇文章套路是不同的  
前者采用了:
    客户端主逻辑函数拿到数据->将数据传入一个心跳管道->心跳管道和当前conn一起传入心跳逻辑函数->心跳逻辑函数内SetDeadline()  
后者则是：
    客户端主逻辑函数拿到数据->主逻辑函数将除listen之外的逻辑都抽象成函数handler()  
    handler函数内再将读抽象成rhandler，写抽象成whandler，而rhandler内用到了SetDeadline()  
两者都值得借鉴，而我需要做的是将心跳功能抽离出来

后者的设计思路太复杂了  
同时并不是长连接不能使用SetDeadline()来实现心跳  
只不过是短链接只要用一下SetDeadline()就可以实现短链接需求  
长连接比短链接只是稍微复杂一些，但是只要能实现功能也没必要设计的过于复杂

# 心跳包只是心跳包  
要明白一件事，那就是心跳包仅仅是某个项目整体套接字逻辑的一个组件  
同样的组件还有客户端管理组件、以及之前所说的加密，其实也应该是属于整体套接字的组件，而不是心跳包的组件  
这里只实现心跳包，其他的都不去实现

# 准备借鉴go-logger  
go-logger整体的设计思路似乎是适配器模式“adapter”  
主体的骨架是logger.go，各个适配器分别位于console.go、file.go等  

首先需要留意的是logger.go 50~60行的Regiser函数，他会在每个适配器对象的.go文件的最后一行被调用

	func Register(adapterName string, newLog adapterLoggerFunc) {
        	if adapters[adapterName] != nil {
	    		panic("logger: logger adapter " + adapterName + " already registered!")
		}
		if newLog == nil {
	    		panic("logger: logger adapter " + adapterName + " is nil!")
		}	    
		adapters[adapterName] = newLog
    	}  
  
他的作用和logger.go 105~115行的Attach方法是有很大区别的:

	func (logger *Logger) Attach(adapterName string, level int, config Config) error {  
		logger.lock.Lock()  
	    	defer logger.lock.Unlock()  

	    	return logger.attach(adapterName, level, config)  
    	}  

前者的作用是包自身的初始化操作  
包自身的逻辑需要先把console、file等适配器实例化并装入整体逻辑  
之后用户才能自己通过Attach方法自己选择使用哪几个适配器  
就好比玩游戏时先要把不同的道具放入背包，玩家在打野时在针对不同的环境再去把道具装配在身上  

console.go对应并实现了console适配器，在末尾行存在：  

    func init() {  
	    Register(CONSOLE_ADAPTER_NAME, NewAdapterConsole)  
    }  
      
file.go对应并实现了file适配器，在末尾行存在：  

    func init() {  
	    Register(FILE_ADAPTER_NAME, NewAdapterFile)  
    }  
    
api.go对应并实现了api适配器，在末尾行存在：  

    func init() {  
	    Register(API_ADAPTER_NAME, NewAdapterApi)  
    }  
    
这几个init都是为了让适配器可以被直接使用，同时也是为了方便用户设计自己的适配器  

而Register的过程也并不复杂，在logger.go里最在一个适配器的缓存:  

	var adapters = make(map[string]adapterLoggerFunc)  
   
Register的作用其实就是往这个缓存里添加适配器对象  
**同时在这里也可以明确，适配器的对象就是以adapterLoggerFunc存在的**    
而之后的Attach方法的核心功能只是告诉模块需要激活、需要真正用到哪个适配器，并为这个适配器设置好参数  

adapterLoggerFunc是函数数据类型，通过type转化为新的实体，从而能为其设计方法，而这个函数类型是如下所示的样子：  

    func NewAdapterConsole() LoggerAbstract {  
	    consoleWrite := &ConsoleWriter{  
		    writer: os.Stdout,  
	    }  
	    config := &ConsoleConfig{}  
	    return &AdapterConsole{  
		    write:  consoleWrite,  
		    config: config,  
	    }  
    }  
  
然后在logger.go中对其进行type：  

    type adapterLoggerFunc func() LoggerAbstract  

其中，NewAdapterConsole() LoggerAbstract{...}是一个数据类型为func() LoggerAbstract的值,**他返回的是一个接口**其实是可以进行这样的操作的（不过没有太大必要）:  

    f :=func NewAdapterConsole() LoggerAbstract {  
	    consoleWrite := &ConsoleWriter{  
		    writer: os.Stdout,  
	    }  
	    config := &ConsoleConfig{}  
	    return &AdapterConsole{  
		    write:  consoleWrite,  
		    config: config,  
	    }  
    }  
    myadapter :=(adapterLoggerFunc)f  
      
回到正题Register其实是用到了很多函数类型的类型转换特性，从而把NewAdapterConsole()、NewAdapterFile（）、NewAdapterApi（）这样的函数转化成实体再利用实体的接口特性，把这3个“函数”转化成真正意义上的同一种数据类型  

**或者说，每个迭代器的源代码里定义了全局形式的原始的函数，而这个函数的数据类型是func() LoggerAbstract  
其中LoggerAbstract这个数据类型是一个能统一不同迭代器（console、file、api等）通用的接口  
在func() LoggerAbstract这种函数类型的内部（如NewAdapterConsole() LoggerAbstract）
不同的迭代器会把与其对应的迭代器结构类返回  
这是这种设计模式的意义所在：**  
console.go对应NewAdapterConsole()LoggerAbstract对应&AdapterConsole{}   
file.go对应NewAdapterFile()LoggerAbstract对应&AdapterFile{}  
api.go对应NewAdapterApi()LoggerAbstract对应&AdapterApi{}  
如上是一套完整的设计思路  

如下又是另一套完整的设计思路，如果把两者混为一滩就很难理清思路了：  
由于各个“New”开头的函数都属于“func() LoggerAbstract”这一函数类型  
因此是可以试下这样的操作的:  

    func (f func() LoggerAbstract)string{  
        la :=f()  
        _ =la  
    }(NewAdapterConsole())    

于是直接type adapterLoggerFunc func() LoggerAbstract，其实就是为了方便而做了一下替换：  

    func (f adapterLoggerFunc)string{  
        la :=f()  
        _ =la  
    }(NewAdapterConsole())  
目前看来仅仅是替换，因为在logger.go的源代码中，虽然可以，但是并没有发现任何为adapterLoggerFunc这个新类型设计了任何方法  
最终通过Register（）函数放入了这个map里：  

    var adapters = make(map[string]adapterLoggerFunc)  

