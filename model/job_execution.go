package model

import (
	"github.com/google/uuid"
	"gopkg.in/guregu/null.v4"
	"time"
)

type JobExecution struct {
	ID           int         `json:"id"`
	JobID        uuid.UUID   `json:"job_id"`
	StartTime    time.Time   `json:"start_time"`
	EndTime      time.Time   `json:"end_time"`
	Success      bool        `json:"success"`
	ErrorMessage null.String `json:"error_message,omitempty"`
}

type JobExecutionStatus string

const (
	JobExecutionStatusSuccessful JobExecutionStatus = "SUCCESSFUL"
	JobExecutionStatusFailed     JobExecutionStatus = "FAILED"
)
