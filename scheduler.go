package cronSchedule

import (
	"fmt"
	"sort"
	"time"
)

//
// Phase 相对于零点的偏移量（以秒为单位），可以有多个或为空。
// Period 周期
type CronJob interface {
	// 任务的唯一标识
	Name() string

	// 具体的任务执行方法
	Process() error

	// 是否启用该任务
	IfActive() bool

	// 发生 panic 后是否自动重启
	IfReboot() bool
}

// 对CronJob接口的包装
type CronJobWrapper struct {
	phase  []int
	period int
	job    CronJob
	count  int
}

func (c *CronJobWrapper) name() string {
	return c.job.Name()
}

func (c *CronJobWrapper) Process() error {
	return c.job.Process()
}

func (c *CronJobWrapper) ifActive() bool {
	return c.job.IfActive()
}

func (c *CronJobWrapper) ifReboot() bool {
	return c.job.IfReboot()
}

type Scheduler struct {
	jobs    []*CronJobWrapper
	nameSet map[string]bool
}

// 生成一个Scheduler实例
func New() *Scheduler {
	return &Scheduler{
		jobs:    nil,
		nameSet: make(map[string]bool),
	}
}

// 将任务注册到Scheduler
// phase: 相对于零点的偏移量（以秒为单位），可以有多个或为空。例如
// []int{} => 从启动时刻开始执行
// []int{3600, 22*3600} => 在 1:00 和 22：00 执行
//
// period: 任务循环的周期（单位为秒）
// 若Phase为空，则Period为任意大于0的值
// 若Phase不为空，则Period必须能被24*3600整除
func (sche *Scheduler) Register(phase []int, period int, job CronJob) {
	// phase 排序
	if len(phase) > 1 {
		sort.Slice(phase, func(i, j int) bool {
			return phase[i] < phase[j]
		})
	}
	jobName := job.Name()
	if _, ok := sche.nameSet[jobName]; jobName == "" || ok {
		// 任务名为空或重复的情况
	} else {
		sche.jobs = append(sche.jobs, &CronJobWrapper{
			job:    job,
			phase:  phase,
			period: period,
		})
		sche.nameSet[job.Name()] = true
	}
}

func (sche *Scheduler) Start() {
	for i := 0; i < len(sche.jobs); i++ {
		if sche.jobs[i].ifActive() && validateJob(sche.jobs[i]) {
			go sche.run(sche.jobs[i])
		}
	}
}

func (sche *Scheduler) run(job *CronJobWrapper) {
	for {
		// 计算下一次运行时间
		nextTimeInterval := calculateNextTime(job.phase, job.period, job.count)
		if nextTimeInterval >= 0 {
			time.Sleep(time.Duration(nextTimeInterval) * time.Second)
		} else {
			// 日志、报警
			// 退出
			break
		}
		err := func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("job %s panic, recover=%v", job.name(), r)
				}
			}()
			return job.Process()
		}()
		// 对任务执行进行计数
		job.count++
		// 如果出错不选择重启，那么直接退出
		if err != nil && !job.ifReboot() {
			break
		}
	}
}

var (
	SecondOfDay = 24 * 3600
)

// 计算下一次任务运行的时间
// phase 初相位
// period 周期
func calculateNextTime(phase []int, period int, calCount int) int {
	if len(phase) == 0 {
		if calCount == 0 {
			return 0
		} else {
			return period
		}
	} else {
		nowTimePhase := int((time.Now().Unix() + 28800) % 86400)
		var i = 0
		for i = 0; i < len(phase); i++ {
			if nowTimePhase < phase[i] {
				break
			}
		}
		if i == len(phase) {
			return period - nowTimePhase + phase[0]
		} else {
			return phase[i] - nowTimePhase
		}
	}
}

// 校验任务是否可以运行
func validateJob(job *CronJobWrapper) bool {
	if len(job.phase) == 0 {
		return job.period > 0
	} else {
		return job.period >= SecondOfDay && job.period%SecondOfDay == 0
	}
}
