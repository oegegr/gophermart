package services

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/oegegr/gophermart/internal/models"
	"github.com/oegegr/gophermart/internal/models/accrual"
	"github.com/oegegr/gophermart/internal/models/api"
	"github.com/oegegr/gophermart/internal/storage"
)

func NewAccrualProcessor(
	client *accrual.ClientWithResponses,
	storage storage.Storage,
	timeout time.Duration,
	workerNum int,
	workerTask int,
) *AccrualProcessor {
	return &AccrualProcessor{
		client:     client,
		storage:    storage,
		timeout:    timeout,
		workerNum:  workerNum,
		orderQueue: make(chan processOrderTask, workerTask),
	}
}

type processOrderTask struct {
	ctx   context.Context
	order models.Order
}

type AccrualProcessor struct {
	client     *accrual.ClientWithResponses
	storage    storage.Storage
	timeout    time.Duration
	workerNum  int
	orderQueue chan processOrderTask
}

func (a *AccrualProcessor) Start(ctx context.Context) {
	var wg sync.WaitGroup

	for i := 0; i < a.workerNum; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()
			a.processOrders(ctx)
		}(i + 1)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(a.timeout)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				orders, err := a.storage.FindOrdersByStatus(ctx, string(api.NEW))
				if err != nil {
					log.Printf("Error finding orders: %v", err)
					continue
				}
				for _, o := range orders {
					a.storage.UpdateOrderStatus(ctx, string(api.PROCESSING), o.Number, 0.0)
					task := processOrderTask{ctx, o}
					select {
					case a.orderQueue <- task:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	wg.Wait()
}

func (a *AccrualProcessor) processOrders(ctx context.Context) {
	for {
		select {
		case orderTask, ok := <-a.orderQueue:
			if !ok {
				return
			}
			if err := a.processOrder(orderTask.ctx, orderTask.order); err != nil {
				log.Printf("failed to process order %s: %v", orderTask.order.Number, err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (a *AccrualProcessor) processOrder(ctx context.Context, order models.Order) error {
	log.Printf("Processing order %s", order.Number)
	resp, err := a.client.GetApiOrdersNumberWithResponse(ctx, order.Number)
	if err != nil {
		return err
	}

	if resp.StatusCode() == http.StatusNoContent {
		log.Printf("ACCRUAL response no content: %d", resp.StatusCode())
		return nil
	}

	if resp.StatusCode() != http.StatusOK {
		log.Printf("ACCRUAL response error: %d", resp.StatusCode())
		return errors.New("ACCRUAL response error")
	}

	accr := resp.JSON200

	status := *accr.Status
	err = a.storage.UpdateOrderStatus(ctx, string(status), *accr.Order, *accr.Accrual)
	if err != nil {
		log.Printf("failed to update order %s status: %v", *accr.Order, err)
		return err
	}

	if status == accrual.PROCESSED {
		err = a.storage.UpdateUserBalance(ctx, order.Login, *accr.Accrual)
		if err != nil {
			log.Printf("failed to update user %s balance: %v", order.Login, err)
			return err
		}
	}
	return nil
}
