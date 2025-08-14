package jobs

import (
	"context"
	"my-api/gmail"
)

type Job interface {
	Name() string
	Schedule() string
	Run(context.Context) error
}

type FoodSpotThreadsJob struct{}

// Runs every 2 hours
func (j FoodSpotThreadsJob) Name() string { return "FoodSpotThreadsJob" }

// included extra 6 fields, leftmost for seconds
func (j FoodSpotThreadsJob) Schedule() string { return "0 0 */2 * * *" }

// Runs every 2 hours
func (j FoodSpotThreadsJob) Run(ctx context.Context) error {
	return gmail.GetNewThreadsByLabel(ctx, "Mangopost/FoodSpot Requests")
}
