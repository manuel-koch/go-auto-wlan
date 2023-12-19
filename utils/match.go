package utils

import "regexp"

func MatchNamedExpression(expression *regexp.Regexp, haystack string) map[string]string {
	names := expression.SubexpNames()
	if len(haystack) > 0 {
		subMatches := expression.FindAllStringSubmatch(haystack, -1)
		if len(subMatches) > 0 {
			results := subMatches[0]
			matches := map[string]string{}
			for i, match := range results {
				matches[names[i]] = match
			}
			return matches
		}
	}
	return nil
}
