package pattern

type Match struct {
	Index   int // value, relative to the beginning of the block of data where the pattern started
	Pattern int // the pattern that matched (carries the slice index)
}

const Wildcard = 0xFF

// MatchBytes matches byte patterns. This isn't designed to be efficient; it does the job well enough, though.
func MatchBytes(data []byte, patterns [][]byte) []Match {
	var results []Match
	for i := range data {
		for j, pattern := range patterns {
			if i + len(pattern) > len(data) {
				continue
			}
			var matches int
			for k, next := range pattern {
				if next == Wildcard || next == data[i+k] {
					matches++
				} else {
					break
				}
				if matches == len(pattern) {
					results = append(results, Match{
						Index:   i,
						Pattern: j,
					})
				}
			}
		}
	}
	return results
}

