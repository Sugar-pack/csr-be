package domain

var (
	OrderStatusInReview   = "in review"
	OrderStatusApproved   = "approved"
	OrderStatusInProgress = "in progress"
	OrderStatusRejected   = "rejected"
	OrderStatusClosed     = "closed"
	OrderStatusPrepared   = "prepared"
	OrderStatusOverdue    = "overdue"
	OrderStatusBlocked    = "blocked"

	// Aggregated states
	OrderStatusAll      = "all"
	OrderStatusActive   = "active"
	OrderStatusFinished = "finished"

	// AllOrderStatuses contains all allowed OrderStatus values
	AllOrderStatuses = map[string]struct{}{
		OrderStatusInReview:   {},
		OrderStatusApproved:   {},
		OrderStatusInProgress: {},
		OrderStatusRejected:   {},
		OrderStatusClosed:     {},
		OrderStatusPrepared:   {},
		OrderStatusOverdue:    {},
		OrderStatusBlocked:    {},
		// Aggregated states
		OrderStatusAll:      {},
		OrderStatusActive:   {},
		OrderStatusFinished: {},
	}

	OrderStatusAggregation = map[string][]string{
		OrderStatusActive: {
			OrderStatusInReview,
			OrderStatusApproved,
			OrderStatusOverdue,
			OrderStatusInProgress,
			OrderStatusPrepared,
		},
		OrderStatusFinished: {
			OrderStatusRejected,
			OrderStatusBlocked,
			OrderStatusClosed,
		},
	}
)
