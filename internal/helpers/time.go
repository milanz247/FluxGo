package helpers

import (
	"fmt"
	"time"
)

func formatDate(layout string, value any) (string, error) {
	switch date := value.(type) {
	case time.Time:
		return date.Format(layout), nil
	case *time.Time:
		if date == nil {
			return "", nil
		}
		return date.Format(layout), nil
	default:
		return "", fmt.Errorf("date requires time.Time, got %T", value)
	}
}

func now() time.Time {
	return time.Now()
}
