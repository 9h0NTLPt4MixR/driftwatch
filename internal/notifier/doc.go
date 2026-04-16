// Package notifier provides lightweight alerting for driftwatch scan results.
//
// After a scan completes, Notify can be called to emit a human-readable
// summary to stdout or stderr. It supports an OnlyDrift mode that suppresses
// output entirely when no drift is detected, making it suitable for use in
// CI pipelines where silent success is preferred.
//
// Example usage:
//
//	notifier.Notify(results, notifier.Options{
//		Channel:   notifier.ChannelStderr,
//		OnlyDrift: true,
//	})
package notifier
