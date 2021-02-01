package filter

import "github.com/CAVaccineInventory/airtable-export/pipeline/pkg/airtable"

// FilterToAllowedKeys takes a slice of KV objects, and a set of allowed key names.
// For each object in the list, it removes each KV pair where the key is not in allowedKeys,
// then returns this result.
func FilterToAllowedKeys(raw airtable.TableContent, allowedKeys []string) airtable.TableContent {
	// Build a map for fast lookup.
	allowedSet := map[string]struct{}{}
	for _, key := range allowedKeys {
		allowedSet[key] = struct{}{}
	}

	filtered := make([]map[string]interface{}, len(raw))

	for i := range raw {
		filtered[i] = map[string]interface{}{}
		for k, v := range raw[i] {
			if _, found := allowedSet[k]; found {
				filtered[i][k] = v
			}
		}
	}

	return filtered
}
