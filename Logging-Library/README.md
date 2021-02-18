# Golang 实现一个异步日志库

## 前言
上班这一年时间天天跟日志打交道什么es，filebeat折腾了一大堆，但是自己还没正儿八经写过日志，正好心血来潮找了个视频学学日志库咋写的，但是像我这种人肯定是不愿意把十几个小时的视频全看完的，直接最后一章把代码扣下来，慢慢研究，果然没跑通，中间查了很多问题，排查过程中也算是对这个程序有了比较深入的理解了，里面用到了不少东西，光是写排序或者别的啥算法肯定是用不到的

## 功能与设计
一个最基本的日志库往往需要具备如下功能：
- 支持输出各种级别的日志
- 输出至文件
- 可以选择输出所需级别的日志
- 日志格式符合常见的日志规范
- 日志文件可以自动轮转切割（按照时间或是大小）
- 日志异步写入

### 初步实现

定义一个FileLogger结构体，包含日志文件的各种属性，业务程序产生的日志都会发送至chanel里面，日志服务单独开启goroutine来消费

``` golang
type FileLogger struct {
	Level       LogLevel
	filePath    string
	fileName    string
	fileObj     *os.File
	errFileObj  *os.File
	maxFileSize int64
	logChan     chan *logMsg
}
```

再定义一条日志的结构体

``` golang
type logMsg struct {
	level     LogLevel
	msg       string
	funcName  string
	fileName  string
	timestamp string
	line      int
}
```

为了实现 选择性输出大于某一级别的日志 这一功能，定义一批常量（iota是一个从0开始的行数索引器，只能在常量里面使用，返回0，1，2，3....)

``` golang
const (
	UNKNOWN LogLevel = iota
	DEBUG
	TRACE
	INFO
	WARNING
	ERROR
	FATAL
)
```

接下来就是一些功能函数了，实现一个log 方法，不断地将日志写入chanel

这里需要注意两点：
* a ...interface{}  是不定参数，也就是可以传入任意参数，因为日志的格式只能确定一个大致框架，具体内容还需要进行填充，因此支持对日志的消息部分进行组合，后续的使用如下： 
``` 
log.Warning("this is warning log")   
log.Error("this is error log,id:%d,name:%s", id, name)
```

* timestamp: now.Format("2006-01-02 15:04:05") 这一段必须怎么写，牢牢记住这个日期！

``` golang
func (f *FileLogger) log(lv LogLevel, format string, a ...interface{}) {
	if f.enable(lv) {
		msg := fmt.Sprintf(format, a...)
		now := time.Now()
		funcName, fileName, lineNo := getInfo(3)
		logtmp := &logMsg{
			level:     lv,
			msg:       msg,
			funcName:  funcName,
			fileName:  fileName,
			timestamp: now.Format("2006-01-02 15:04:05"),
			line:      lineNo,
		}
		select {
		case f.logChan <- logtmp:
		default:

		}
	}
}
```

日志规范一般会要求打印出对应的文件名，函数以及行号，形成一个完整的调用链便于查找错误，因此构造一个getInfo方法来获取这些值


``` golang
func getInfo(n int) (funcName, fileName string, lineNo int) {
	pc, file, lineNo, ok := runtime.Caller(n)
	if !ok {
		fmt.Printf("runtime.Caller() failed\n")
		return
	}
	funcName = runtime.FuncForPC(pc).Name()
	fileName = path.Base(file)
	funcName = strings.Split(funcName, ".")[1]
	return
}
```
该处涉及到runtime包的使用了
- Caller 方法反应的是堆栈信息中某堆栈帧所在文件的绝对路径和语句所在文件的行数
- func Caller(skip int) (pc uintptr, file string, line int, ok bool)

      　　pc是uintptr这个返回的是函数指针
      　　file是函数所在文件名目录
      　　line所在行号
      　　ok 是否可以获取到信息

- 再通过runtime.FuncForPC方法获取到具体的方法名称

生产者这一端差不多了，再来看消费者

这里直接初始化结构体，创建了两个文件权限是644（感觉对文件的操作go语言比python要复杂很多。。。），os.OpenFile返回的是文件的指针，一并写入结构体中返回给调用者，创建完文件之后开启一个线程来消费管道中的日志

这里能不能开多个线程呢？当然可以，前提是不会发生日志轮转，当某个线程关闭了文件，开启新的文件，其他线程使用的还是旧文件的指针，这里靠加锁可没办法解决，我个人想法是声明一个全局的指针用来保存文件信息，每个goroutine运行前检查一下有没有变化（或者变化后主动通知？）变了就都用新的，没变就继续。。。太复杂了
``` go
func NewFileLogger(levelStr, fp, fn string, maxSize int64) *FileLogger {
	loglevel, err := parseLogLevel(levelStr)
	if err != nil {
		panic(err)
	}
	f1 := &FileLogger{
		Level:       loglevel,
		filePath:    fp,
		fileName:    fn,
		maxFileSize: maxSize,
		logChan:     make(chan *logMsg, MaxSize),
	}
	err = f1.initFile()
	if err != nil {
		panic(err)
	}
	return f1
}
```
``` go
func (f *FileLogger) initFile() error {
	fullFileName := path.Join(f.filePath, f.fileName)
	fileObj, err := os.OpenFile(fullFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	errFileObj, err := os.OpenFile(fullFileName+".err", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	f.fileObj = fileObj
	f.errFileObj = errFileObj
	go f.writeLogBackground()
	return nil
}
```

最后看这个writeLogBackground跟splitFile方法，一个死循环，每次先判断一下是否达到大小上线了，达到了就调用一下split，获取当前文件的信息和当前的时间信息，把旧的文件关了，重命名一下，再开启一个新的文件，注意之前init函数只会调用一次，后续日志轮转都是通过split来完成的

如果消费管道里的日志时发现级别大于ERROR，就再记录一份到error log里面

select通常用来取管道中的值，如果单独取可能出现管道为空的情况，这样一般是返回0值，或者报错，使用select配合deefault就可以避免，如果有多个 case 都可以运行，Select 会随机公平地选出一个执行。其他不会执行。
如果有 default 子句，则执行该语句。
如果没有 default 子句，select 将阻塞，直到某个通信可以运行；Go 不会重新对 channel 或值进行求值。

``` go
func (f *FileLogger) writeLogBackground() {
	for {
		if f.checkSize(f.fileObj) {
			newFile, err := f.splitFile(f.fileObj)
			if err != nil {
				return
			}
			f.fileObj = newFile
		}
		select {
		case logtmp := <-f.logChan:
			logInfo := fmt.Sprintf("[%s] [%s] [%s:%s:%d] %s\n", logtmp.timestamp, getLogString(logtmp.level), logtmp.fileName, logtmp.funcName, logtmp.line, logtmp.msg)
			fmt.Fprintf(f.fileObj, logInfo)
			if logtmp.level >= ERROR {
				if f.checkSize(f.errFileObj) {
					newerrFile, err := f.splitFile(f.errFileObj)
					if err != nil {
						return
					}
					f.errFileObj = newerrFile
				}
				fmt.Fprintf(f.errFileObj, logInfo)

			}
		default:
			time.Sleep(time.Millisecond * 500)
		}
	}
}
```
``` go
func (f *FileLogger) splitFile(file *os.File) (*os.File, error) {
	//need to split
	nowStr := time.Now().Format("20060102150405000") //must be this time
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Printf("get file name info faled ,err:%v\n", err)
		return nil, err
	}
	logName := path.Join(f.filePath, fileInfo.Name())
	newLogName := fmt.Sprintf("%s.bak%s", logName, nowStr)
	file.Close()
	os.Rename(logName, newLogName)
	fileObj, err := os.OpenFile(logName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("create file faled ,err:%v\n", err)
		return nil, err
	}
	return fileObj, nil
}
```

到这里原理差不多就完成了，之后再写一个main函数来测试它就好了

这里直接实例化了这个结构，往里传值就好了
``` go

func main() {
	log := mylog.NewFileLogger("INFO", "./", "wth.log", 10*1024*1024)
	for {
		log.Debug("this is debug log")
		log.Info("this is info log")
		log.Warning("this is warning log")
		id := 10010
		name := "test"
		log.Error("this is error log,id:%d,name:%s", id, name)
		log.Fatal("this is fatal log")
	}
}
```
有个地方一开始不太了解，就是为什么要定义一个接口，接口里面的函数都已经实现了呀？

```go
type Logger interface {
	Debug(format string, a ...interface{})
	Info(format string, a ...interface{})
	Warning(format string, a ...interface{})
	Error(format string, a ...interface{})
	Fatal(format string, a ...interface{})
}
```

后面详细看了一下，发现视频里还实现了一个日志标准输出的功能console，相当于既能支持写文件，也可以标准输出，提供的方法是相同的，如果要同时使用的话，需要写成如下形式，

``` go
func main() {
	log := mylog.NewFileLogger("INFO", "./", "wth.log", 10*1024*1024)
	logConsole := mylog.Newconsole("INFO")
	for {
		id := 10010
		name := "test"
		log.Debug("this is debug log")
		log.Info("this is info log")
		log.Warning("this is warning log")
		log.Error("this is error log,id:%d,name:%s", id, name)
		log.Fatal("this is fatal log")
		logConsole.Debug("this is debug log")
		logConsole.Info("this is info log")
		logConsole.Warning("this is warning log")
		logConsole.Error("this is error log,id:%d,name:%s", id, name)
		logConsole.Fatal("this is fatal log")
	}
}
```

修改之后声明了一个全局的接口，再到函数里面分别赋值
``` go
var log mylog.Logger

func main() {
	log = mylog.NewFileLogger("INFO", "./", "wth.log", 10*1024*1024)
	for {
		log.Debug("this is debug log")
		log.Info("this is info log")
		log.Warning("this is warning log")
		id := 10010
		name := "test"
		log.Error("this is error log,id:%d,name:%s", id, name)
		log.Fatal("this is fatal log")
	}
}
```

 留个坑，目前看起来轮转规则是按照大小，之后补一个支持时间的，先算大小，到一定时间还没达到设定值依然轮转

 另外轮转还有一个重要的功能就是压缩以及定期删除，之后有空写一下