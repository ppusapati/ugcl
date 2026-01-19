package repository

// RepositoryManager combines all repository interfaces for insightviewer module
type RepositoryManager struct {
	Execution    IExecutionRepository
	Export       IExportRepository
	Scheduling   ISchedulingRepository
	Subscription ISubscriptionRepository
}

// NewRepositoryManager creates a new repository manager instance
func NewRepositoryManager(
	executionRepo IExecutionRepository,
	exportRepo IExportRepository,
	schedulingRepo ISchedulingRepository,
	subscriptionRepo ISubscriptionRepository,
) *RepositoryManager {
	return &RepositoryManager{
		Execution:    executionRepo,
		Export:       exportRepo,
		Scheduling:   schedulingRepo,
		Subscription: subscriptionRepo,
	}
}