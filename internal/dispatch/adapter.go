package dispatch

import (
	"context"

	"fixapp/internal/domain"
	
	"github.com/google/uuid"
)

// JobDispatcher adapts the dispatch.Service to the job.Dispatcher interface.
type JobDispatcher struct {
	service *Service
}

// NewJobDispatcher creates a new dispatcher adapter.
func NewJobDispatcher(service *Service) *JobDispatcher {
	return &JobDispatcher{service: service}
}

// DispatchJob implements job.Dispatcher.
func (d *JobDispatcher) DispatchJob(ctx context.Context, job *domain.Job, handymanIDs []uuid.UUID) error {
	_, err := d.service.DispatchJob(ctx, job, handymanIDs)
	return err
}
