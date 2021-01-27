package filter

// FilterToAllowedKeys takes a slice of KV objects, and a set of allowed key names.
// For each object in the list, it removes each KV pair where the key is not in allowedKeys.
func FilterToAllowedKeys(jsonMap []map[string]interface{}, allowedKeys map[string]struct{}) {
	for i := range jsonMap {
		for k := range jsonMap[i] {
			if _, found := allowedKeys[k]; !found {
				delete(jsonMap[i], k)
			}
		}
	}
}
