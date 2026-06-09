package search

import (
	"sort"

	"github.com/sahilm/fuzzy"
)

// sortMatches keeps the fuzzy search output stable across ties.
func sortMatches(matches []fuzzy.Match) {
	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Score == matches[j].Score {
			return matches[i].Str < matches[j].Str
		}

		return matches[i].Score > matches[j].Score
	})
}
