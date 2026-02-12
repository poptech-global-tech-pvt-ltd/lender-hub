package entity

import "time"

// BlockInfo is a value object for user block details
type BlockInfo struct {
	IsBlocked     bool
	Reason        string
	Source        string      // e.g. "FRAUD", "MANUAL", "DELINQUENCY"
	NextEligibleAt *time.Time // nil if permanent
}
