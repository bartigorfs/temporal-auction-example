package auction

import (
	"context"
	"fmt"
)

func StartAuctionActivity(ctx context.Context, auctionID string) error {
	fmt.Printf("Auction %s started\n", auctionID)
	return nil
}

func EndAuctionActivity(ctx context.Context, auctionID string, winningBid Bid) error {
	fmt.Printf("Auction %s ended with winning bid %v\n", auctionID, winningBid)
	return nil
}
