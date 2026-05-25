package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/SamyRai/go-jules/internal/apicoverage"
)

func main() {
	discoveryURL := os.Getenv("JULES_DISCOVERY_URL")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	report, err := apicoverage.ValidateURL(ctx, discoveryURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Jules API coverage OK: revision=%s operations=%d schemas=%d enums=%d\n",
		report.Revision,
		report.OperationCount,
		report.SchemaCount,
		report.EnumCount,
	)
}
