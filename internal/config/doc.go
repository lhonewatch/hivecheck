// Package config handles loading, parsing, and validating the hivecheck
// YAML configuration file.
//
// A typical configuration file looks like:
//
//	checks:
//	  - name: payments-api
//	    url: https://payments.internal/health
//	    interval: 30s
//	    timeout: 5s
//	    warn_at: 200   # response-time ms
//	    crit_at: 500
//
//	alerts:
//	  - url: https://hooks.example.com/hivecheck
//	    headers:
//	      Authorization: Bearer secret
//
// Use [Load] to read a file from disk. The returned [Config] is safe to
// pass directly to the check runner and alert hooks.
package config
