package auction

import (
	"go.temporal.io/sdk/workflow"
	"time"
)

type WorkflowInput struct {
	AuctionID  string
	Duration   time.Duration
	StartPrice float64
}

type Bid struct {
	UserID string
	Amount float64
}

func AuctionWorkflow(ctx workflow.Context, input WorkflowInput) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	err := workflow.ExecuteActivity(ctx, StartAuctionActivity, input.AuctionID).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	var bids []Bid
	signalChan := workflow.GetSignalChannel(ctx, "BidSignal")
	timerFuture := workflow.NewTimer(ctx, input.Duration)

	selector := workflow.NewSelector(ctx)

	auctionCompleted := false

	selector.AddFuture(timerFuture, func(f workflow.Future) {
		auctionCompleted = true
	})

	selector.AddReceive(signalChan, func(c workflow.ReceiveChannel, more bool) {
		var bid Bid
		c.Receive(ctx, &bid)
		bids = append(bids, bid)
	})

	for !auctionCompleted {
		selector.Select(ctx)
	}

	if len(bids) == 0 {
		return "", nil
	}

	winningBid := determineWinningBid(bids)
	err = workflow.ExecuteActivity(ctx, EndAuctionActivity, input.AuctionID, winningBid).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return winningBid.UserID, nil
}

func determineWinningBid(bids []Bid) Bid {
	var maxBid Bid
	for _, bid := range bids {
		if bid.Amount > maxBid.Amount {
			maxBid = bid
		}
	}
	return maxBid
}
