/*
 * SPDX-FileCopyrightText: © Hypermode Inc. <hello@hypermode.com>
 * SPDX-License-Identifier: Apache-2.0
 */

package atomspace

// AtomFilter is a predicate used to filter atoms during pattern matching.
type AtomFilter func(Atom) bool

// ByType returns an AtomFilter that matches atoms of the given type.
func ByType(t AtomType) AtomFilter {
	return func(a Atom) bool { return a.GetType() == t }
}

// ByMinStrength returns an AtomFilter that keeps atoms whose truth-value
// strength is at least the given threshold.
func ByMinStrength(threshold float64) AtomFilter {
	return func(a Atom) bool { return a.GetTruthValue().Strength >= threshold }
}

// ByMinConfidence returns an AtomFilter that keeps atoms whose truth-value
// confidence is at least the given threshold.
func ByMinConfidence(threshold float64) AtomFilter {
	return func(a Atom) bool { return a.GetTruthValue().Confidence >= threshold }
}

// ByMinSTI returns an AtomFilter that keeps atoms whose short-term importance
// is at least the given threshold.
func ByMinSTI(threshold int64) AtomFilter {
	return func(a Atom) bool { return a.GetAttentionValue().STI >= threshold }
}

// All returns an AtomFilter that is true only when all provided filters are true.
func All(filters ...AtomFilter) AtomFilter {
	return func(a Atom) bool {
		for _, f := range filters {
			if !f(a) {
				return false
			}
		}
		return true
	}
}

// Any returns an AtomFilter that is true when at least one provided filter is true.
func Any(filters ...AtomFilter) AtomFilter {
	return func(a Atom) bool {
		for _, f := range filters {
			if f(a) {
				return true
			}
		}
		return false
	}
}

// QueryAtoms returns all atoms in the space that satisfy every given filter.
// Filters are ANDed together for convenience; use Any/All combinators for
// more complex logic.
func (ds *DgraphSpace) QueryAtoms(filters ...AtomFilter) []Atom {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	combined := All(filters...)
	var result []Atom
	for _, atom := range ds.atoms {
		if combined(atom) {
			result = append(result, atom)
		}
	}
	return result
}

// FindLinks returns all links that connect a specific atom in their outgoing
// set and that satisfy every given filter.
func (ds *DgraphSpace) FindLinks(atom Atom, filters ...AtomFilter) []Atom {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	linkIDs, ok := ds.incoming[atom.GetID()]
	if !ok {
		return nil
	}

	combined := All(filters...)
	var result []Atom
	for _, id := range linkIDs {
		if link, exists := ds.atoms[id]; exists && combined(link) {
			result = append(result, link)
		}
	}
	return result
}
