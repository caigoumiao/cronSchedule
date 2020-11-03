# cronSchedule
cronSchedule 是一个基于Golang 实现的定时任务框架。

当前支持两种定时模式：
1. 设定时间间隔k，每隔时间k 运行一次
2. 设定初相位数组，在一天的多个时间点运行

## 安装
````
go get -u -v github.com/caigoumiao/cronSchedule
````
推荐使用go.mod
<br>
````
require github.com/caigoumiao/cronSchedule latest
````

## 快速使用
1、初始化
```go
sche := cronSchedule.New()
```
cronSchedule 支持自定义日志实现，如果不提供则选用默认实现。
```go
type Logger interface {
	InfoF(format string, args ...interface{})
	ErrorF(format string, args ...interface{})
}

// logger 为用户自己的 Logger 实现
sche.SetLogger(logger)
```
2、实现CronJob接口，注册定时任务
```go
type CronJobPrintf struct{}

func (CronJobPrintf) Name() string {
	return "cronJobPrintf"
}

func (CronJobPrintf) Process() error {
	fmt.Println("cronjob process printf")
	return nil
}

func (CronJobPrintf) IfActive() bool {
	return true
}

func (CronJobPrintf) IfReboot() bool {
	return true
}

// 每5秒执行一次CronJobPrintf
sche.Register([]int{}, 5, CronJobPrintf{})
// 每天的1:00、10:20执行一次
sche.Register([]int{1*3600, 10*3600+20*60}, 86400, CronJobPrintf{})
```
3、开启调度器
```go
sche.Start()
```

## 待完善
目前的定时模式过于简单，并不能涵盖大多数需求。举例如下：
+ 每周三执行、每月的3号执行
+ 在22:00-1:00之间，每隔10分钟执行

下一版本考虑加入cron表达式，力求提供一个更加抽象好用的定时模式。

## 致谢
相遇是缘！感恩🙏🙏🙏

如果你喜欢本项目或本项目有帮助到你，希望你可以帮忙 star 一下。

如果你有任何意见或建议，欢迎提 issue 或联系我本人。联系方式如下：
+ 微信：wo4qiaoba
