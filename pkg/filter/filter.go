package filter

// FilterToAllowedKeys takes a slice of KV objects, and a set of allowed key names.
// For each object in the list, it removes each KV pair where the key is not in allowedKeys,
// then returns this result.
func FilterToAllowedKeys(raw []map[string]interface{}, allowedKeys map[string]struct{}) []map[string]interface{} {
	copy := make([]map[string]interface{}, len(raw))

	for i := range raw {
		copy[i] = map[string]interface{}{}
		for k, v := range raw[i] {
			if _, found := allowedKeys[k]; found {
				copy[i][k] = v
			}
		}
	}

	return copy
}
