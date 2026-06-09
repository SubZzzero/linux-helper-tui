package executor

import (
	"errors"
	"fmt"

	"linux-helper/internal/models"
)

// ErrConfirmationRequired is returned when a dangerous recipe is not confirmed.
var ErrConfirmationRequired = errors.New("confirmation required")

// ConfirmRisk checks whether the caller confirmed the recipe risk.
func ConfirmRisk(level models.RiskLevel, confirmed bool) error {
	if level == models.RiskDangerous && !confirmed {
		return fmt.Errorf("%w for dangerous recipe", ErrConfirmationRequired)
	}

	return nil
}
