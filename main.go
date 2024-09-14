package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"temporal-test/auction"
	"time"
)

func main() {
	temporalClient, err := client.Dial(client.Options{})
	if err != nil {
		panic(err)
	}
	defer temporalClient.Close()

	w := worker.New(temporalClient, "AuctionTaskQueue", worker.Options{})
	w.RegisterWorkflow(auction.AuctionWorkflow)
	w.RegisterActivity(auction.StartAuctionActivity)
	w.RegisterActivity(auction.EndAuctionActivity)

	go func() {
		err := w.Run(worker.InterruptCh())
		if err != nil {
			panic(err)
		}
	}()

	app := fiber.New()

	app.Post("/auction", func(c *fiber.Ctx) error {
		var req struct {
			AuctionID  string  `json:"auction_id"`
			Duration   int     `json:"duration"`
			StartPrice float64 `json:"start_price"`
		}

		if err := c.BodyParser(&req); err != nil {
			return err
		}

		workflowOptions := client.StartWorkflowOptions{
			ID:        req.AuctionID,
			TaskQueue: "AuctionTaskQueue",
		}

		workflowInput := auction.WorkflowInput{
			AuctionID:  req.AuctionID,
			Duration:   time.Duration(req.Duration) * time.Second,
			StartPrice: req.StartPrice,
		}

		workflow, err := temporalClient.ExecuteWorkflow(c.Context(), workflowOptions, auction.AuctionWorkflow, workflowInput)
		if err != nil {
			return err
		}

		fmt.Println(workflow.GetID(), workflow.GetRunID())

		return c.SendStatus(fiber.StatusCreated)
	})

	app.Post("/auction/:id/bid", func(c *fiber.Ctx) error {
		var req struct {
			UserID string  `json:"user_id"`
			Amount float64 `json:"amount"`
		}

		if err := c.BodyParser(&req); err != nil {
			return err
		}

		auctionID := c.Params("id")
		bid := auction.Bid{
			UserID: req.UserID,
			Amount: req.Amount,
		}

		err := temporalClient.SignalWorkflow(c.Context(), auctionID, "", "BidSignal", bid)
		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusOK)
	})

	err = app.Listen(":3000")
	if err != nil {
		return
	}
}
