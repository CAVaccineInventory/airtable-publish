package types

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestClone ensures that copies match the source, and modifying the copy doesn't affect the source.
func TestClone(t *testing.T) {
	src := TableContent{
		map[string]interface{}{
			"a": "1",
			"b": 2,
		},
	}
	// want should match the first element of src above.
	want := map[string]interface{}{
		"a": "1",
		"b": 2,
	}

	cpy := src.Clone()

	// First, make sure the clone matches the source.
	if diff := cmp.Diff(src, cpy); diff != "" {
		t.Errorf("src and copy are different: -src +cpy:\n%v\n", diff)
	}

	// Then, modify the clone, and make sure the source isn't modified.
	cpy[0]["a"] = "X"
	cpy[0]["b"] = "Y"

	if diff := cmp.Diff(want, src[0]); diff != "" {
		t.Errorf("src modified unexpectedly: -want +src[0]:\n%v\n", diff)
	}

}
