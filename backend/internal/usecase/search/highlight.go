package search

import (
	"slices"
	"unicode"
	"unicode/utf16"
)

// highlightRanges は body 内で語に大文字小文字を区別せず一致した範囲を、重なりをまとめて昇順で返します
func highlightRanges(body string, terms []string) []TextRange {
	runes := foldRunes(body)
	// runes[i] の開始位置を UTF-16 コード単位で表したもの
	offsets := make([]int, len(runes)+1)
	for i, r := range []rune(body) {
		offsets[i+1] = offsets[i] + utf16.RuneLen(r)
	}

	ranges := []TextRange{}
	for _, term := range terms {
		needle := foldRunes(term)
		if len(needle) == 0 {
			continue
		}
		for i := 0; i+len(needle) <= len(runes); i++ {
			if slices.Equal(runes[i:i+len(needle)], needle) {
				ranges = append(ranges, TextRange{Start: offsets[i], End: offsets[i+len(needle)]})
			}
		}
	}

	slices.SortFunc(ranges, func(a, b TextRange) int { return a.Start - b.Start })
	merged := []TextRange{}
	for _, r := range ranges {
		if last := len(merged) - 1; last >= 0 && r.Start <= merged[last].End {
			merged[last].End = max(merged[last].End, r.End)
			continue
		}
		merged = append(merged, r)
	}
	return merged
}

func foldRunes(s string) []rune {
	runes := []rune(s)
	for i, r := range runes {
		runes[i] = unicode.ToLower(r)
	}
	return runes
}
