package util

import (
	"fmt"
	"os"

	"github.com/avast/retry-go"
)

/* Function checkFilesExist:
 * returns an err if any of the given filePaths do not exist.
 */
func CheckFilesExist(filePaths []string) error {
	for _, filePath := range filePaths {
		if !fileExists(filePath) {
			return fmt.Errorf("could not find file: %v", filePath)
		}
	}
	return nil
}

/* Function fileExists:
 * checks to see if a file exists.
 */
func fileExists(filePath string) bool {
	err := retry.Do(
		func() error {
			_, err := os.Stat(filePath)
			return err
		},
		// Set to four retries;
		// Default is 10 defined in newDefaultRetryConfig()
		retry.Attempts(4),
	)

	return err == nil
}
