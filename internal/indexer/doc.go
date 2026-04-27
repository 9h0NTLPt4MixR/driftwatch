// Package indexer provides fast in-memory lookup structures over drift scan
// results.
//
// After a scan produces a []drift.Result, call Build to construct an Index.
// The Index supports O(1) lookup by service name and O(1) lookup of all
// services that drifted on a particular configuration key.
//
// Example:
//
//	idx, err := indexer.Build(results)
//	if err != nil { ... }
//
//	if r, ok := idx.LookupService("payments"); ok {
//	    fmt.Println(r.Drifted)
//	}
//
//	for _, r := range idx.LookupKey("timeout") {
//	    fmt.Println(r.Service)
//	}
package indexer
