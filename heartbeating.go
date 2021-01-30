
/** HeartBeating不仅仅可以用作tcp的心跳包，其他的链接类型，如果有长连接需求也适用
 * 具体的使用方式是，当外层完成通过net.Conn封装自定义Conn时后，将自定义Conn作为参数传入Handler方法
 * 自定义Conn（如下面的ZYUnifiedConn）是个接口，实现了Attach方法
 * (如：
	    ZYHB :=(HeartBeatHandler)HeartBeating(zqy_go_logger)
		ZYUnifiedConn ：=NewZYUnifiedConn(xxx,xxx,xxx)
		ZYUnifiedConn.Attach("heartbeating", ZYHB)
 * )
 * 函数数据类型是引用类型,最后的函数类型参数是比较重要的核心
 * 他被设计成自定义Conn的一个内部字段
 * 这样就可以把心跳包逻辑从整体套接字通信逻辑中抽离出来
 * 同时，ZYHB是个单例，他的Handler可以分别应用于多个自定义conn，自定义conn的内部是tcp，udp，snmp也都是可以的
 */

 
package heartbeating

import (
	logger "github.com/phachon/go-logger"

	"heartbeating/conf"
	"heartbeating/log"

	"fmt"
	"strconv"
)


type HeartBeatHandler HeartBeating
func HeartBeating(l *go_logger.Logger){
	defer ConfFlush()
	Conf(l *go_logger.Logger)	
}

type Conn interface{
	HeartBeatChan() chan byte
	LocalAddr() string
	ClientAddrAndType() string,string
	DisconnectionFromServer() error
}

/** 将会在循环里运行核心逻辑：
 * 以接收到心跳事件作为触发条件
 * 刷新所设定的超时秒数
 */
func (p *HeartBeating)Handler(conn Conn, timeoutSce int){
	HBch :=conn.HeartBeatChan()
	defer close(HBch)

	clientaddr, clienttype := conn.ClientAddrAndType();    localaddr :=conn.LocalAddr()
	logger.Debug("地址为%s的系统开始进行对地址为%s、"+
				 "通信类型为%s的连接进行心跳监控",
				 localaddr, clientaddr, DefineString(clienttype))

	timer := time.NewTimer(time.Duration(timeoutSec) * time.Second)
	for {
		select {
		/*由于心跳刷新/主动Stop()发生在到期之前，基于先后顺序原则，Stop()有效、到期无效*/
		case ok :=<-HBch:
			if !ok { continue }
			/** Reset()之前必须先正确的Stop()
			 * 这里的使用场景决定了如果真的Stop失败
			 * 接下来必然会有数据在timer.C内
			 * 并不会出现死锁
			 */
			if timer.Stop == STOP_AFTER_EXPIRE{ 
				logger.Warning("当心跳包进行timer的Reset操作时与timer自身的到期事件"+
							   "发生了race condition（竞争条件之下心跳事件发生在前")
				_ <-timer.C 
			} 
			timer.Reset(time.Duration(timeoutSec) * time.Second)
		/*由于先到期后存检测到了新事件，基于先后顺序原则，到期有效、该事件无效*/
		case <-timer.C:
			/** 是有可能会出现到期后，析构前HBch里存在数据情况
			 * 但是所对应的事件是无效事件 
			 */
			if len(HBch)>0 { 
				logger.Warning("当心跳包的timer到期时恰好有新的心跳事件，"+
				               "两者之间发生了race condition（竞争条件之下到期事件发生在前）")
				_ <-HBch 
			}

			err :=conn.DisconnectionFromServer()
			if err ==nil{
				logger.Info(fmt.Sprintf("IP地址为%s的心跳包服务检测到类型为%s、"+
							"地址为%s的客户端连接超时，"+
							"并成功从服务端主动断开",
							localaddr,DefineString(clienttype),clientaddr)
			}else{
				logger.Error("IP地址为%s的心跳包服务检测到类型为%s、"+
							 "地址为%s的客户端连接超时，"+
							 "但尝试从服务端主动断开时发生如下错误：%s",
							 localaddr,DefineString(clienttype),clientaddr,err.String())
			}

			return
		}
	}
}