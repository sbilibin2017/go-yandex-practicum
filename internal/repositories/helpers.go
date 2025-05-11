package repositories

import "fmt"

func generateMetricKey(data map[string]any) string {
	return fmt.Sprintf("%s:%s", data["id"], data["type"])
}
