// Package types holds stand-alone type definitions.  This package should have
// minimal (if any) dependencies, and only on the standard library -- to prevent
// import cycles.
package types

// TableContent represents a generic table, consisting of rows which contain
// string keys and arbitrary values.
type TableContent []map[string]interface{}
