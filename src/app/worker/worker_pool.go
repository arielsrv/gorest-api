package worker2

import (
	"github.com/alitto/pond"
	"github.com/prometheus/client_golang/prometheus"
	log "gitlab.com/iskaypetcom/digital/sre/tools/dev/go-logger"
)

type Pool struct {
	*pond.WorkerPool
}

func New(maxWorkers, maxCapacity int, options ...pond.Option) *Pool {
	pool := pond.New(maxWorkers, maxCapacity, options...)

	// Worker pool metrics
	err := prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "pool_workers_running",
			Help: "Number of running worker goroutines",
		},
		func() float64 {
			return float64(pool.RunningWorkers())
		}))

	if err != nil {
		log.Error(err)
	}

	err = prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "pool_workers_idle",
			Help: "Number of idle worker goroutines",
		},
		func() float64 {
			return float64(pool.IdleWorkers())
		}))

	if err != nil {
		log.Error(err)
	}

	// Task metrics
	err = prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "pool_tasks_submitted_total",
			Help: "Number of tasks submitted",
		},
		func() float64 {
			return float64(pool.SubmittedTasks())
		}))

	if err != nil {
		log.Error(err)
	}

	err = prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "pool_tasks_waiting",
			Help: "Number of tasks waiting in the queue",
		},
		func() float64 {
			return float64(pool.WaitingTasks())
		}))

	if err != nil {
		log.Error(err)
	}

	err = prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "pool_tasks_successful_total",
			Help: "Number of tasks that completed successfully",
		},
		func() float64 {
			return float64(pool.SuccessfulTasks())
		}))

	if err != nil {
		log.Error(err)
	}

	err = prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "pool_tasks_failed_total",
			Help: "Number of tasks that completed with panic",
		},
		func() float64 {
			return float64(pool.FailedTasks())
		}))

	if err != nil {
		log.Error(err)
	}

	err = prometheus.Register(prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: "pool_tasks_completed_total",
			Help: "Number of tasks that completed either successfully or with panic",
		},
		func() float64 {
			return float64(pool.CompletedTasks())
		}))

	if err != nil {
		log.Error(err)
	}

	return &Pool{
		WorkerPool: pool,
	}
}
