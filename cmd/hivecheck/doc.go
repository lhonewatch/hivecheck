// Package main provides the hivecheck command-line interface.
//
// Usage:
//
//	hivecheck [flags]
//
// Flags:
//
//	-config string
//		Path to the YAML configuration file (default "hivecheck.yaml").
//
//	-format string
//		Output format for the health-check report. Accepted values:
//		  text  — human-readable table (default)
//		  json  — machine-readable JSON object
//
//	-timeout duration
//		Global deadline applied to all checks (default 30s).
//
// Exit codes:
//
//	0  All checks passed (OK or WARN).
//	1  Fatal error (config missing, formatter failure, etc.).
//	2  One or more checks reached CRITICAL status.
//
// Alert webhooks are fired after the report is printed when
// alerts.webhook_url is non-empty in the config file.
package main
