package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/Quickaxe-Martina/gofermart/internal/client"
	"github.com/Quickaxe-Martina/gofermart/internal/storage"
	"go.uber.org/zap"
)

// ErrWorkerStopped woker stopped
var ErrWorkerStopped = errors.New("woker stopped")

// CheckOrderTask task model for check order
type checkOrderTask struct {
	OrderNumber   string
	ScheduledTime time.Time
	Attempts      int
	UserID        int
}

type resultTask struct {
	result    client.AccrualResponse
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
	logging  *zap.Logger
	ctx      context.Context
	cancel   context.CancelFunc
}

// OrderWorkersConfig params for init OrderWorkers
type OrderWorkersConfig struct {
	Store       storage.Storage
	NumWorkers  int
	URL         string
	PoolSize    int
	PoolTimeout time.Duration
	Logging     *zap.Logger
}

// NewOrderWorkers create OrderWorkers
func NewOrderWorkers(cfg OrderWorkersConfig) (*OrderWorkers, error) {
	client, err := client.NewAccrualClient(cfg.URL, cfg.PoolSize, cfg.PoolTimeout)
	if err != nil {
		cfg.Logging.Error("", zap.Error(err))
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())

	wm := &OrderWorkers{
		store:    cfg.Store,
		client:   client,
		inputCh:  make(chan checkOrderTask, 100),
		workerCh: make(chan checkOrderTask, cfg.NumWorkers*3),
		resultCh: make(chan resultTask, 100),
		doneCh:   make(chan struct{}),
		logging:  cfg.Logging,
		ctx:      ctx,
		cancel:   cancel,
	}

	wm.wg.Add(1)
	go wm.aggregator()
	for i := 0; i < cfg.NumWorkers; i++ {
		wm.wg.Add(1)
		go wm.worker(i)
		wm.wg.Add(1)
		go wm.resultWorker(i)
	}

	return wm, nil
}

func (wm *OrderWorkers) aggregator() {
	wm.logging.Info("aggregator started")
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

		case <-wm.ctx.Done():
			wm.logging.Info("aggregator stopped")
			timer.Stop()
			return
		}
	}
}

func (wm *OrderWorkers) worker(id int) {
	defer wm.wg.Done()
	wm.logging.Info(fmt.Sprintf("worker-%d started", id))
	for {
		select {
		case <-wm.ctx.Done():
			wm.logging.Info(fmt.Sprintf("worker-%d stopping", id))
			return
		case task := <-wm.workerCh:
			wm.handleCheckOrder(task)
		}
	}
}

func (wm *OrderWorkers) handleCheckOrder(task checkOrderTask) {
	ctx, cancel := context.WithTimeout(wm.ctx, 5*time.Second)
	defer cancel()

	result, err := wm.client.GetOrder(ctx, task.OrderNumber)
	if err != nil {
		if errors.Is(err, client.ErrTooManyRequest) {
			task.Attempts++
			task.ScheduledTime = time.Now().Add(time.Second * 10)
			wm.inputCh <- task
		}
	}
	if result.Status == storage.StatusProcessing || result.Status == storage.StatusNew {
		task.Attempts++
		task.ScheduledTime = time.Now().Add(time.Second * 5)
		wm.inputCh <- task
	} else if result.Status == storage.StatusProcessed || result.Status == storage.StatusInvalid {
		wm.resultCh <- resultTask{
			result:    result,
			checkTask: task,
		}
	}
}

func (wm *OrderWorkers) resultWorker(id int) {
	ctx, cancel := context.WithTimeout(wm.ctx, 5*time.Second)
	defer cancel()
	defer wm.wg.Done()
	wm.logging.Info(fmt.Sprintf("result worker-%d started", id))
	for {
		select {
		case <-wm.ctx.Done():
			wm.logging.Info(fmt.Sprintf("result worker-%d stopping", id))
			return
		case task := <-wm.resultCh:
			wm.store.AccrueUser(ctx, task.checkTask.UserID, task.result.Accrual, task.result.Status, task.checkTask.OrderNumber)
		}
	}
}

// AddTask add task to check order
func (wm *OrderWorkers) AddTask(OrderNumber string, userID int) error {
	select {
	case <-wm.doneCh:
		return ErrWorkerStopped
	default:
	}

	wm.inputCh <- checkOrderTask{OrderNumber: OrderNumber, ScheduledTime: time.Now(), Attempts: 0, UserID: userID}
	return nil
}

// Stop end workers work
func (wm *OrderWorkers) Stop() {
	wm.cancel()
	close(wm.doneCh)
	wm.wg.Wait()
	wm.logging.Info("All workers stopped")
	close(wm.inputCh)
}
