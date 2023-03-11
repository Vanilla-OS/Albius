package albius

import (
	"fmt"
	"os"
)

// Set locale
// Set keyboard
// Set user

func SetTimezone(tz string, targetRoot string) error {
	tzPath := targetRoot + "/etc/timezone"

	err := os.WriteFile(tzPath, []byte(tz), 0644)
	if err != nil {
		return fmt.Errorf("Failed to set timezone: %s", err)
	}

	return nil
}
