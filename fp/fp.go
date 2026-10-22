// Package fp contains thin adapters over github.com/samber/lo
// which omit the index parameter from the callbacks.
package fp

import "github.com/samber/lo"

// Map applies the mapper to every item of the collection.
func Map[T, R any](collection []T, mapper func(T) R) []R {
	return lo.Map(collection, func(item T, _ int) R { return mapper(item) })
}

// Filter returns the items of the collection for which the predicate is true.
func Filter[T any](collection []T, predicate func(T) bool) []T {
	return lo.Filter(collection, func(item T, _ int) bool { return predicate(item) })
}
