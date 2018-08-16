package testsupport

// CompareActions compares strings a and b in slice actions
// and returns:
// >0 if a.index > b.index
// <0 if a.index < b.index
// 0 if a.index == b.index
func CompareActions(actions []string, a, b string) int {
	m := make(map[string]int)
	for i, e := range actions {
		m[e] = i
	}
	return m[a] - m[b]
}
