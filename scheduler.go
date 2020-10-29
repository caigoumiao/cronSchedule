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

	// 相对于零点的偏移量（以秒为单位），可以有多个或为空。例如
	// []int{} => 从启动时刻开始执行
	// []int{3600, 22*3600} => 在 1:00 和 22：00 执行
	Phase() []int

	// 任务循环的周期（单位为秒）
	// 若Phase为空，则Period为任意大于0的值
	// 若Phase不为空，则Period必须能被24*3600整除
	Period() int

	// 具体的任务执行方法
	Process() error

	// 是否启用该任务
	IfActive() bool

	// 发生 panic 后是否自动重启
	IfReboot() bool
}

type Scheduler struct {
	jobs  []CronJob
	count map[string]int
}

// 生成一个Scheduler实例
func New() *Scheduler {
	return &Scheduler{
		jobs:  nil,
		count: make(map[string]int),
	}
}

// 将任务注册到Scheduler
func (sche *Scheduler) Register(job CronJob) {
	tmpArr := job.Phase()
	sort.Slice(tmpArr, func(i, j int) bool {
		return tmpArr[i] < tmpArr[j]
	})
	if _, ok := sche.count[job.Name()]; job.Name() == "" || ok {
		// 任务名为空或重复需要提示一下
	} else {
		sche.jobs = append(sche.jobs, job)
	}
}

func (sche *Scheduler) Start() {
	for i := 0; i < len(sche.jobs); i++ {
		if sche.jobs[i].IfActive() && validateJob(sche.jobs[i]) {
			go sche.run(sche.jobs[i])
		}
	}
}

func (sche *Scheduler) run(job CronJob) {
	for {
		// 计算下一次运行时间
		nextTimeInterval := calculateNextTime(job.Phase(), job.Period(), sche.count[job.Name()])
		if nextTimeInterval > 0 {
			time.Sleep(time.Duration(nextTimeInterval) * time.Second)
		} else {
			// 日志、报警
			// 退出
			break
		}
		err := func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("job %s panic, recover=%v", job.Name(), r)
				}
			}()
			return job.Process()
		}()
		if err != nil {
			sche.count[job.Name()]++
		}
		// 如果出错不选择重启，那么直接退出
		if err != nil && !job.IfReboot() {
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
			return phase[i]
		}
	}
}

// 校验任务是否可以运行
func validateJob(job CronJob) bool {
	if len(job.Phase()) == 0 {
		return job.Period() > 0
	} else {
		return job.Period() >= SecondOfDay && job.Period()%SecondOfDay == 0
	}
}
