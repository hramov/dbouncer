package utils

func MergeStatMaps(m1 map[string]int, m2 map[string]int) {
	for k, v := range m2 {
		m1[k] = m1[k] + v
	}
}

func MergeStatMapsAll(m1 map[string]int, m2 map[string]int) int {
	res := 0
	for k, v := range m2 {
		m1[k] = m1[k] + v
		res += m1[k] + v
	}

	return res
}
