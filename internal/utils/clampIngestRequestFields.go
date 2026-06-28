package utils

import (
	"strconv"

	"github.com/dishan1223/mutt/consts"
	"github.com/dishan1223/mutt/models"
)

// The `ClampIngestRequestFields` function takes an `IngestRequest` object as input and ensures
// that the lengths of the `Log` and `StackTrace` fields do not exceed predefined maximum sizes.
// It retrieves these maximum sizes from environment variables, converts them to integers, and truncates
// the fields if they exceed the limits.
// If any error occurs during this process (e.g., invalid size values), it returns an error.
func ClampIngestRequestFields(body *models.IngestRequest) error {
	maxLogSize, err := strconv.Atoi(consts.MAX_LOG_SIZE)
	if err != nil {
		return err
	}
	if len(body.Log) > maxLogSize {
		body.Log = body.Log[:maxLogSize]
	}

	maxStackTrace, err := strconv.Atoi(consts.MAX_STACK_TRACE)
	if err != nil {
		return err
	}
	if len(body.StackTrace) > maxStackTrace {
		body.StackTrace = body.StackTrace[:maxStackTrace]
	}

	return nil
}
