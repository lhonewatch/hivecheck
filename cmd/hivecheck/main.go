// Package main is the entry point for the hivecheck CLI tool.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/example/hivecheck/internal/alert"
	"github.com/example/hivecheck/internal/check"
	"github.com/example/hivecheck/internal/config"
	"github.com/example/hivecheck/internal/report"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "hivecheck: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfgPath := flag.String("config", "hivecheck.yaml", "path to config file")
	format := flag.String("format", "text", "output format: text or json")
	timeout := flag.Duration("timeout", 30*time.Second, "global timeout for all checks")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	runner := check.NewRunner(cfg.Checks)

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	results := runner.Run(ctx)
	summary := report.NewSummary(cfg.ServiceName, results)

	var formatter report.Formatter
	switch *format {
	case "json":
		formatter = report.NewJSONFormatter()
	default:
		formatter = report.NewTextFormatter()
	}

	output, err := formatter.Format(summary)
	if err != nil {
		return fmt.Errorf("formatting report: %w", err)
	}
	fmt.Print(output)

	if len(cfg.Alerts.WebhookURL) > 0 {
		hook := alert.NewWebhookHook(cfg.Alerts.WebhookURL, cfg.Alerts.TimeoutSeconds)
		dispatcher := alert.NewDispatcher(hook,
			alert.WithMinStatus(cfg.Alerts.MinStatus),
			alert.WithLogFailure(true),
		)
		if dispErr := dispatcher.Dispatch(ctx, summary); dispErr != nil {
			fmt.Fprintf(os.Stderr, "hivecheck: alert dispatch: %v\n", dispErr)
		}
	}

	if summary.Overall() >= check.StatusCritical {
		os.Exit(2)
	}
	return nil
}
