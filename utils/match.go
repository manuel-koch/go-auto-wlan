// This file is part of go-auto-wlan.
//
// go-auto-wlan is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-auto-wlan is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-auto-wlan. If not, see <http://www.gnu.org/licenses/>.
//
// Copyright 2023 Manuel Koch
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
