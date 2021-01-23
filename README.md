# heartbeating
为了实现一个在线更新客户端软件的服务器，同时考虑流量成本，于是决定采用socket长连接的方式维持客户端与服务端之间的通信  
因此需要实现一个简单且稳定的长连接工具包  
此工具包会分为客户端与服务端  

# 此包应该是隶属一个底层基于net.Conn接口所实现封装结构类的，一个类似适配器的东西  
# 我要先设计出个这结构类的demo才能开拓思路：  

	github.com/ziyouzy/zconn/


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

# 不能完全照搬go-logger的设计模式  
**从go-logger各个迭代器都需要实现LoggerAbstract接口这一点来看，“心跳包适配器”所包含的功能方法或许不会被之后可能会去设计的“crc校验适配器”、“读取解码适配器”、“发送加密适配器”等在功能上发生重叠**
**说白了，他们并不回去搞同一类型的事，以至于在设计逻辑上没有足够的理由让这些模块去实现同一个接口**  
唯一值得借鉴的只是attach这个方法，~~以及他的参数表~~（参数表不值得借鉴，值得借鉴的是他通过反射拿到config结构类实体的这种技巧）：  
    
	func (logger *Logger) Attach(adapterName string, level int, config Config) error {  }
    
	vc := reflect.ValueOf(consoleConfig)
	cc := vc.Interface().(*ConsoleConfig)
	adapterConsole.config = cc
***      
***
***
***
  
  
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
~~包自身的逻辑需要先把console、file等适配器实例化并装入整体逻辑~~不是转入实例化所得的结构类或实现它的接口，而是实现实例化所需的“函数数据类型”  
之后用户才能自己通过Attach方法自己选择使用哪几个适配器  
就好比玩游戏时先要把不同的道具放入背包，玩家在打野时在针对不同的环境再去把道具装配在身上  
**补充：Attach方法内会真正去执行所需要的“函数数据类型”，从而拿到各个实体，这种设计模式是为了节省资源**   

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
而之后的Attach方法的核心功能只是告诉模块需要激活、需要真正用到哪个适配器，**并为这个适配器设置好参数**  

adapterLoggerFunc是函数数据类型，通过type转化为新的实体，~~从而能为其设计方法~~(但是并没有设计方法)，而这个函数类型是如下所示的样子：  

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

其中，NewAdapterConsole() LoggerAbstract{...}是一个数据类型为func() LoggerAbstract的值,**他返回的是一个接口**~~其实是可以进行这样的操作的~~（没有太大必要）:  

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
      
~~回到正题Register其实是用到了很多函数类型的类型转换特性，从而把NewAdapterConsole()、NewAdapterFile（）、NewAdapterApi（）这样的函数转化成实体再利用实体的接口特性，把这3个“函数”转化成真正意义上的同一种数据类型~~这句话的逻辑实在混乱  

**或者说，每个迭代器的源代码里定义了全局形式的原始的函数，而这个函数的数据类型是func() LoggerAbstract  
其中LoggerAbstract这个数据类型是一个能统一不同迭代器（console、file、api等）通用的接口
其在之后所接收的是各个迭代器真正的结构类实体，这些实体都会实现这个接口
也就是说，在func() LoggerAbstract这种函数类型的内部（如NewAdapterConsole() LoggerAbstract）
不同的迭代器会把与其对应的迭代器结构类返回  
这是这种设计模式的意义所在：**  

console.go对应NewAdapterConsole()LoggerAbstract对应&AdapterConsole{}结构类实体   
file.go对应NewAdapterFile()LoggerAbstract对应&AdapterFile{}结构类实体  
api.go对应NewAdapterApi()LoggerAbstract对应&AdapterApi{}结构类实体  
如上是一套完整的设计思路  

**如下又是另一套完整的设计思路，如果把两者混为一滩就很难理清思路了：**  
由于上述各个“New”开头的函数都属于“func() LoggerAbstract”这一函数类型  
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
**目前看来仅仅是替换**，因为在logger.go的源代码中，虽然可以，但是并没有发现任何为adapterLoggerFunc这个新类型设计了任何方法  
最终通过Register（）函数放入了这个map里：  

    var adapters = make(map[string]adapterLoggerFunc)  

这两个类型都存在于logger.go（程序主逻辑中），他们是各个适配器与主逻辑协调运作的桥梁：  

	type adapterLoggerFunc func() LoggerAbstract

	type LoggerAbstract interface {
		Name() string
		Init(config Config) error
		Write(loggerMsg *loggerMessage) error
		Flush()
	}

实现LoggerAbstract是在各个迭代器文件里的结构类  
实现func() LoggerAbstract以及其别名adapterLoggerFunc是在各个迭代器文件里的New函数  
真正实例化迭代器的操作存在于logger.go文件的115~155行的attach方法：   

	func (logger *Logger) attach(adapterName string, level int, config Config) error {
		for _, output := range logger.outputs {
			if output.Name == adapterName {
				printError("logger: adapter " + adapterName + "already attached!")
			}
		}
		logFun, ok := adapters[adapterName]
		if !ok {
			printError("logger: adapter " + adapterName + "is nil!")
		}
		adapterLog := logFun()
		err := adapterLog.Init(config)
		if err != nil {
			printError("logger: adapter " + adapterName + " init failed, error: " + err.Error())
		}

		output := &outputLogger{
			Name:           adapterName,
			Level:          level,
			LoggerAbstract: adapterLog,
		}

		logger.outputs = append(logger.outputs, output)
		return nil
	}
	
其中的logFun, ok := adapters[adapterName]可以像取出一个变量一样取出一个函数数据类型  
然后adapterLog := logFun()才是真正执行了这个函数，从而拿到一个LoggerAbstract接口，等同于拿到了某个适配器的结构类  
**虽然也可以在进行Register等操作时直接难道adapterLog，或许是为了节省资源才这么设计的：**  

	var adapters = make(map[string]adapterLoggerFunc)也需要改成var adapters = make(map[string]LoggerAbstract)  
	
**这样的设计模式是耗费资源的，因为缓存里会存在所有的适配器结构类对象实体**  

**而detach采用的不是从现有logger.outputs（map）移除某个适配器，而是创建新map，替换掉旧的：**

	func (logger *Logger) detach(adapterName string) error {
		outputs := []*outputLogger{}
		for _, output := range logger.outputs {
			if output.Name == adapterName {
				continue
			}
			outputs = append(outputs, output)
		}
		logger.outputs = outputs
		return nil
	}
	
**现在是全新的知识点：**

	adapterLog := logFun()
	err := adapterLog.Init(config)

在attach方法内部会对各个适配器进行真正的初始化，参数只有一个，是个名为config的接口类型，本以为这个接口里会包含很多方法标签，但是错了，在config.go中只有：  

	package go_logger

	// logger config interface
	type Config interface {
		Name() string
	}
	
于是还是重点看看"适配器.Init()"这个方法吧： 

	func (adapterConsole *AdapterConsole) Init(consoleConfig Config) error {
		if consoleConfig.Name() != CONSOLE_ADAPTER_NAME {
			return errors.New("logger console adapter init error, config must ConsoleConfig")
		}

		vc := reflect.ValueOf(consoleConfig)
		cc := vc.Interface().(*ConsoleConfig)
		adapterConsole.config = cc

		if cc.JsonFormat == false && cc.Format == "" {
			cc.Format = defaultLoggerMessageFormat
		}

		return nil
	}
	
**大佬用了反射，这难道就是反射的正确使用场景吗**  

还是先学习下consoleConfig吧，毕竟这个最简单:  

	type ConsoleConfig struct {
		// console text is show color
		Color bool

		// is json format
		JsonFormat bool

		// jsonFormat is false, please input format string
		// if format is empty, default format "%millisecond_format% [%level_string%] %body%"
		//
		//  Timestamp "%timestamp%"
		//	TimestampFormat "%timestamp_format%"
		//	Millisecond "%millisecond%"
		//	MillisecondFormat "%millisecond_format%"
		//	Level int "%level%"
		//	LevelString "%level_string%"
		//	Body string "%body%"
		//	File string "%file%"
		//	Line int "%line%"
		//	Function "%function%"
		//
		// example: format = "%millisecond_format% [%level_string%] %body%"
		Format string
	}
	
还是用fileConfig对比着看吧：

	type FileConfig struct {

		// log filename
		Filename string

		// level log filename
		LevelFileName map[int]string

		// max file size
		MaxSize int64

		// max file line
		MaxLine int64

		// file slice by date
		// "y" Log files are cut through year
		// "m" Log files are cut through mouth
		// "d" Log files are cut through day
		// "h" Log files are cut through hour
		DateSlice string

		// is json format
		JsonFormat bool

		// jsonFormat is false, please input format string
		// if format is empty, default format "%millisecond_format% [%level_string%] %body%"
		//
		//  Timestamp "%timestamp%"
		//	TimestampFormat "%timestamp_format%"
		//	Millisecond "%millisecond%"
		//	MillisecondFormat "%millisecond_format%"
		//	Level int "%level%"
		//	LevelString "%level_string%"
		//	Body string "%body%"
		//	File string "%file%"
		//	Line int "%line%"
		//	Function "%function%"
		//
		// example: format = "%millisecond_format% [%level_string%] %body%"
		Format string
	}
	
基本上都是简单数据类型，切片已经算是最复杂的数据结构了，同时初始的状态他们的内部所有字段没有赋任何值，赋值的操作存在于各个“适配器.go”的“适配器.Init(适配器Config Config))方法  
同时这个方法的参数表并不是各个config的结构类，而是实现结构类的接口  
他的外层是func (logger *Logger) attach(adapterName string, level int, config Config) error {}  
**config接口作为参数一直会传递到最内部**

回到反射这个事，其实用法也并不复杂：  

	vc := reflect.ValueOf(consoleConfig)
	cc := vc.Interface().(*ConsoleConfig)
	adapterConsole.config = cc

基本上可以理解成json的序列化与反序列化的操作consoleConfig以接口形式传进来  
其真正的模板就是ConsoleConfig结构类，类内部也都是golang内置的数据类型，最复杂的也就是个切片  
反序列化过程很安全，cc已经是个有效的结构类了，里面各个字段都已经是有具体值的了  

**至此，接下来该去研究具体怎么使用log了，也就是适配器结构类的Write()这些方法：**  
设计到的核心结构类为下面两个：

	// adapter console
	type AdapterConsole struct {
		write  *ConsoleWriter
		config *ConsoleConfig
	}

	// console writer
	type ConsoleWriter struct {
		lock   sync.Mutex
		writer io.Writer
	}
	
后者扮演了前者内置字段的角色，同时io.Writer的功能其实就是命令行输出  
AdapterConsole的Write方法是此包的功能核心最具代表性的内容：  

	func (adapterConsole *AdapterConsole) Write(loggerMsg *loggerMessage) error {

		msg := ""
		if adapterConsole.config.JsonFormat == true {
			//jsonByte, _ := json.Marshal(loggerMsg)
			jsonByte, _ := loggerMsg.MarshalJSON()
			msg = string(jsonByte)
		} else {
			msg = loggerMessageFormat(adapterConsole.config.Format, loggerMsg)
		}
		consoleWriter := adapterConsole.write

		if adapterConsole.config.Color {
			colorAttr := adapterConsole.getColorByLevel(loggerMsg.Level, msg)
			consoleWriter.lock.Lock()
			color.New(colorAttr).Println(msg)
			consoleWriter.lock.Unlock()
			return nil
		}

		consoleWriter.lock.Lock()
		consoleWriter.writer.Write([]byte(msg + "\n"))
		consoleWriter.lock.Unlock()

		return nil
	}
	
主要做了下面两件事：  
1.合成字符串：  
  1.创建空msg  
  2.基于loggerMsg拿到具体的“内容”  
  3.选择内容的格式是json格式还是普通文本格式
  4.赋给msg
2.选择输出方式：  
  1.基于布尔值adapterConsole.config.Color决定是否进行彩色输出  
  2.是则结束color包实现彩色输出，否则借助io.writer实现普通输出  
  
**主逻辑差不多就是这样，先写这么多**
