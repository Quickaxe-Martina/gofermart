package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/Quickaxe-Martina/gofermart/internal/client"
	"github.com/Quickaxe-Martina/gofermart/internal/logger"
	"github.com/Quickaxe-Martina/gofermart/internal/storage"
	"go.uber.org/zap"
)

// ErrWorkerStopped woker stopped
var ErrWorkerStopped = errors.New("woker stopped")

// CheckOrderTask todo
type checkOrderTask struct {
	OrderNumber   string
	ScheduledTime time.Time
	Attempts      int
}

type resultTask struct {
	result    client.AccrualResponse
	err       error
	checkTask checkOrderTask
}

// OrderWorkers struct
type OrderWorkers struct {
	store    storage.Storage
	client   client.AccrualClient
	inputCh  chan checkOrderTask
	workerCh chan checkOrderTask
	resultCh chan resultTask
	doneCh   chan struct{}
	wg       sync.WaitGroup
}

// NewOrderWorkers create OrderWorkers
func NewOrderWorkers(store storage.Storage, numWorkers int, URL string, poolSize int, poolTimeout time.Duration) *OrderWorkers {
	client, err := client.NewAccrualClient(URL, poolSize, poolTimeout)
	if err != nil {
		logger.Log.Error("", zap.Error(err))
		return nil
	}
	wm := &OrderWorkers{
		store:    store,
		client:   client,
		inputCh:  make(chan checkOrderTask, 100),
		workerCh: make(chan checkOrderTask, numWorkers*3),
		resultCh: make(chan resultTask, 100),
		doneCh:   make(chan struct{}),
	}

	wm.wg.Add(1)
	go wm.aggregator()
	for i := 0; i < numWorkers; i++ {
		wm.wg.Add(1)
		go wm.worker(i)
	}

	return wm
}

func (wm *OrderWorkers) aggregator() {
	logger.Log.Info("aggregator started")
	defer wm.wg.Done()

	var tasks []checkOrderTask
	timer := time.NewTimer(time.Hour * 24)
	timer.Stop()

	for {
		var nextTime time.Time
		var delay time.Duration

		if len(tasks) > 0 {
			nextTime = tasks[0].ScheduledTime
			now := time.Now()
			if nextTime.After(now) {
				delay = nextTime.Sub(now)
			} else {
				delay = 0
			}
			timer.Reset(delay)
		}

		select {
		case task := <-wm.inputCh:
			tasks = append(tasks, task)
			sort.Slice(tasks, func(i, j int) bool {
				return tasks[i].ScheduledTime.Before(tasks[j].ScheduledTime)
			})

		case <-timer.C:
			now := time.Now()
			i := 0

			for ; i < len(tasks); i++ {
				t := tasks[i]
				if t.ScheduledTime.After(now) {
					break
				}
				wm.workerCh <- t
			}
			if i < len(tasks) {
				tasks = tasks[i:]
			} else {
				tasks = tasks[:0]
				timer.Stop()
			}

		case <-wm.doneCh:
			logger.Log.Info("aggregator stopped")
			timer.Stop()
			return
		}
	}
}

func (wm *OrderWorkers) worker(id int) {
	defer wm.wg.Done()
	logger.Log.Info(fmt.Sprintf("worker-%d started", id))
	for {
		select {
		case <-wm.doneCh:
			logger.Log.Info(fmt.Sprintf("worker-%d stopping", id))
			return
		case task := <-wm.workerCh:
			wm.handleCheckOrder(task)
		}
	}
}

func (wm *OrderWorkers) handleCheckOrder(task checkOrderTask) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := wm.client.GetOrder(ctx, task.OrderNumber)
	if err != nil {
		if errors.Is(err, client.ErrToManyRequests) {
			task.Attempts++
			task.ScheduledTime = time.Now().Add(time.Minute * 10)
			wm.inputCh <- task
		}
	}
	if result.Status == "PROCESSING" {
		task.Attempts++
		task.ScheduledTime = time.Now().Add(time.Minute * 5)
		wm.inputCh <- task
	}
}

// AddTask add task to check order
func (wm *OrderWorkers) AddTask(OrderNumber string) error {
	select {
	case <-wm.doneCh:
		return ErrWorkerStopped
	default:
	}

	wm.inputCh <- checkOrderTask{OrderNumber: OrderNumber, ScheduledTime: time.Now(), Attempts: 0}
	return nil
}

// Stop end workers work
func (wm *OrderWorkers) Stop() {
	close(wm.doneCh)
	wm.wg.Wait()
	logger.Log.Info("All workers stopped")
	close(wm.inputCh)
}
