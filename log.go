/** logger.Attach(adapterName string, level int, config Config) error
 * 第一个参数是适配器的名称，是模块的内置适配器，直接使用即可
 * 必须确保输入名称字符串是正确，才能让内置的功能函数里遍历到对应的适配器对象
 * 第二个参数是日志级别
 * 第三个参数是适配器的参数对象，不同适配器参数对象的内部字段与构造都是不同的
 */

/* Detach->拆卸		Attach->装配 */

package heartbeating


import (
	logger "github.com/phachon/go-logger"
)

/** 
 * 需要先完成下面的初始化后在执行Log()
 * 在别的包使用他的方式:
 * log.Log().Info("xxxxxxx")
 * log.Log().Warning("xxxxxxx")
 */
func Log() *logger.Logger{	return l	}


/*单例实体与初始化函数*/
var l *logger.Logger
func Logger(lp *logger.Logger){
	if lp !=nil{ l = lp }//上层已创建好了一个go-loggr对象，就不用自己创建了

	l = logger.NewLogger()

	/** 在这里设定异步方式
	 * 程序结束前必须调用 Flush
	 * 已简单实现了析构函数LogFlush()
	 */
	l.SetAsync()

	/** 由于准备设置多个输出
	 * 需要先拆卸默认的"console"适配器
	 * 然后自定义新的适合多个输出场景的新“console”适配器
	 * 同时还需要自定义好诸如"file"等适配器，一并重新装载
	 */
	l.Detach("console")


	// 新命令行输出配置
	consoleConfig := &logger.ConsoleConfig{
        Color: true, // 命令行输出字符串是否显示颜色
        JsonFormat: true, // 命令行输出字符串是否格式化
        Format: "", // 如果输出的不是 json 字符串，JsonFormat: false, 自定义输出的格式
    }
    // 添加 console 为 logger 的一个输出
    l.Attach("console", logger.LOGGER_LEVEL_DEBUG, consoleConfig)


    // 文件输出配置
    fileConfig := &go_logger.FileConfig {
		Filename : "./test.log", // 日志输出文件名，不自动存在
		
		/** 如果要将单独的日志分离为文件
		 * 请配置LealFrimeNem参数
		 */
        LevelFileName : map[int]string {
            l.LoggerLevel("error"): "./error.log",    // Error 级别日志被写入 error .log 文件
			l.LoggerLevel("warning"): "./warning.log",// Warn 级别日志被写入到 Warn.log 文件中
			l.LoggerLevel("info"): "./info.log",      // Info 级别日志被写入到 info.log 文件中
            l.LoggerLevel("debug"): "./debug.log",    // Debug 级别日志被写入到 debug.log 文件中
		},
		
        MaxSize : 1024 * 1024,  // 文件最大值（KB），默认值0不限
		
		MaxLine : 100000, // 文件最大行数，默认 0 不限制
		
		DateSlice : "m",  // 文件根据日期切分， 支持 "Y" (年), "m" (月), "d" (日), "H" (时), 默认 "no"， 不切分
		
		JsonFormat: true, // 写入文件的数据是否 json 格式化
		
		Format: "", // 如果写入文件的数据不 json 格式化，自定义日志格式

	}	
    // 添加 file 为 logger 的一个输出
    l.Attach("file", go_logger.LOGGER_LEVEL_DEBUG, fileConfig)
}

/*析构函数，由于开启了异步模式，所以必须在创建实体后立刻做好defer*/
func LogFlush(){
	l.Flush()
}