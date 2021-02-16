// Package types holds stand-alone type definitions.  This package should have
// minimal (if any) dependencies, and only on the standard library -- to prevent
// import cycles.
package types

// TableContent represents a generic table, consisting of rows which contain
// string keys and arbitrary values.
type TableContent []map[string]interface{}

// Clone does a clone of a TableContent structure returning a new copy.
func (tc TableContent) Clone() TableContent {
	var out TableContent

	for _, row := range tc {
		var new = make(map[string]interface{})
		for k, v := range row {
			new[k] = v
		}
		out = append(out, new)
	}
	return out
}
