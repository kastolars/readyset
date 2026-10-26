// Package setlaws checks the algebraic laws that every set type in this module
// must satisfy. It exists so the same properties are tested against each
// implementation. It is only meant to be imported from tests.
package setlaws

import "testing"

// Set is the algebra shared by the set types in this module. S is the set type
// itself, so that operations take and return the same type.
type Set[S any] interface {
	Len() int
	IsEmpty() bool
	Clone() S
	Equal(S) bool
	IsSubset(S) bool
	IsSuperset(S) bool
	IsProperSubset(S) bool
	IsProperSuperset(S) bool
	IsDisjoint(S) bool
	Union(S) S
	Intersection(S) S
	Difference(S) S
	SymmetricDifference(S) S
}

// Check builds three sets from x, y and z with newSet and fails t if any
// identity of set algebra does not hold for them.
func Check[S Set[S]](t testing.TB, newSet func(...byte) S, x, y, z []byte) {
	t.Helper()

	a, b, c := newSet(x...), newSet(y...), newSet(z...)
	empty := newSet()
	u := a.Union(b).Union(c) // a universe containing all three sets

	// Snapshots, to confirm that no operation modifies its operands.
	a0, b0, c0 := a.Clone(), b.Clone(), c.Clone()

	check := func(name string, ok bool) {
		t.Helper()
		if !ok {
			t.Fatalf("law violated: %s\n x=%v\n y=%v\n z=%v", name, x, y, z)
		}
	}

	// Commutativity.
	check("a∪b = b∪a", a.Union(b).Equal(b.Union(a)))
	check("a∩b = b∩a", a.Intersection(b).Equal(b.Intersection(a)))
	check("a△b = b△a", a.SymmetricDifference(b).Equal(b.SymmetricDifference(a)))

	// Associativity.
	check("(a∪b)∪c = a∪(b∪c)", a.Union(b).Union(c).Equal(a.Union(b.Union(c))))
	check("(a∩b)∩c = a∩(b∩c)", a.Intersection(b).Intersection(c).Equal(a.Intersection(b.Intersection(c))))

	// Distributivity.
	check("a∩(b∪c) = (a∩b)∪(a∩c)", a.Intersection(b.Union(c)).Equal(a.Intersection(b).Union(a.Intersection(c))))
	check("a∪(b∩c) = (a∪b)∩(a∪c)", a.Union(b.Intersection(c)).Equal(a.Union(b).Intersection(a.Union(c))))

	// Idempotence and identity.
	check("a∪a = a", a.Union(a).Equal(a))
	check("a∩a = a", a.Intersection(a).Equal(a))
	check("a∖a = ∅", a.Difference(a).Equal(empty))
	check("a∪∅ = a", a.Union(empty).Equal(a))
	check("a∩∅ = ∅", a.Intersection(empty).Equal(empty))
	check("a△a = ∅", a.SymmetricDifference(a).Equal(empty))

	// Inclusion–exclusion.
	check("|a∪b| = |a|+|b|-|a∩b|", a.Union(b).Len() == a.Len()+b.Len()-a.Intersection(b).Len())

	// De Morgan, relative to the universe u.
	check("u∖(a∪b) = (u∖a)∩(u∖b)", u.Difference(a.Union(b)).Equal(u.Difference(a).Intersection(u.Difference(b))))
	check("u∖(a∩b) = (u∖a)∪(u∖b)", u.Difference(a.Intersection(b)).Equal(u.Difference(a).Union(u.Difference(b))))

	// Symmetric difference, defined two ways.
	check("a△b = (a∪b)∖(a∩b)", a.SymmetricDifference(b).Equal(a.Union(b).Difference(a.Intersection(b))))
	check("a△b = (a∖b)∪(b∖a)", a.SymmetricDifference(b).Equal(a.Difference(b).Union(b.Difference(a))))

	// Subset, superset and disjointness agree with the operations.
	check("a⊆a∪b", a.IsSubset(a.Union(b)))
	check("a∩b⊆a", a.Intersection(b).IsSubset(a))
	check("a⊆b ⇔ a∩b = a", a.IsSubset(b) == a.Intersection(b).Equal(a))
	check("a⊆b ⇔ b⊇a", a.IsSubset(b) == b.IsSuperset(a))
	check("a⊂b ⇔ a⊆b and a≠b", a.IsProperSubset(b) == (a.IsSubset(b) && !a.Equal(b)))
	check("a⊂b ⇔ b⊃a", a.IsProperSubset(b) == b.IsProperSuperset(a))
	check("a⊄a", !a.IsProperSubset(a) && !a.IsProperSuperset(a))
	check("a∖b is disjoint from b", a.Difference(b).IsDisjoint(b))
	check("disjoint ⇔ a∩b = ∅", a.IsDisjoint(b) == a.Intersection(b).IsEmpty())
	check("a⊆b and b⊆a ⇔ a = b", (a.IsSubset(b) && b.IsSubset(a)) == a.Equal(b))

	// Operands are unchanged.
	check("operands unchanged", a.Equal(a0) && b.Equal(b0) && c.Equal(c0))
}
