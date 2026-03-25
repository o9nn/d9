/*
 * SPDX-FileCopyrightText: © Hypermode Inc. <hello@hypermode.com>
 * SPDX-License-Identifier: Apache-2.0
 */

package atomspace

import (
	"context"
	"encoding/json"
	"math"
	"testing"
)

// ---------------------------------------------------------------------------
// NumberNode tests
// ---------------------------------------------------------------------------

func TestNumberNode(t *testing.T) {
	n := NewNumberNode(3.14)
	if n.Value != 3.14 {
		t.Errorf("expected Value 3.14, got %g", n.Value)
	}
	if n.GetType() != NumberNodeType {
		t.Errorf("expected NumberNodeType, got %s", n.GetType())
	}
	if n.Name != "3.14" {
		t.Errorf("expected name '3.14', got '%s'", n.Name)
	}
}

func TestAddNumberNode(t *testing.T) {
	ctx := context.Background()
	space := NewDgraphSpace(ctx)

	n1, err := space.AddNumberNode(42.0)
	if err != nil {
		t.Fatalf("AddNumberNode: %v", err)
	}

	n2, err := space.AddNumberNode(42.0)
	if err != nil {
		t.Fatalf("AddNumberNode duplicate: %v", err)
	}
	if n1.GetID() != n2.GetID() {
		t.Errorf("duplicate AddNumberNode should return the same atom")
	}

	size, _ := space.Size()
	if size != 1 {
		t.Errorf("expected size 1, got %d", size)
	}
}

func TestNewAtomTypes(t *testing.T) {
	types := []AtomType{
		NumberNodeType, VariableNodeType, SchemaNodeType, GroundedSchemaNodeType,
		MemberLinkType, ListLinkType, SetLinkType,
		AndLinkType, OrLinkType, NotLinkType,
		ImplicationLinkType, EquivalenceLinkType,
		ExecutionLinkType, ContextLinkType,
	}
	for _, at := range types {
		if at == "" {
			t.Errorf("AtomType should not be empty string")
		}
	}
}

func TestIsNodeIsLink(t *testing.T) {
	node := NewNode(ConceptNodeType, "x")
	num := NewNumberNode(1.0)
	link := NewLink(InheritanceLinkType, []Atom{node, node})

	if !IsNode(node) {
		t.Error("IsNode(node) should be true")
	}
	if !IsNode(num) {
		t.Error("IsNode(num) should be true")
	}
	if IsNode(link) {
		t.Error("IsNode(link) should be false")
	}
	if IsLink(node) {
		t.Error("IsLink(node) should be false")
	}
	if !IsLink(link) {
		t.Error("IsLink(link) should be true")
	}
}

// ---------------------------------------------------------------------------
// JSON round-trip tests
// ---------------------------------------------------------------------------

func TestJSONRoundTrip(t *testing.T) {
	ctx := context.Background()
	space := NewDgraphSpace(ctx)

	cat, _ := space.AddNode(ConceptNodeType, "cat")
	animal, _ := space.AddNode(ConceptNodeType, "animal")
	link, _ := space.AddLink(InheritanceLinkType, []Atom{cat, animal})
	link.SetTruthValue(NewTruthValue(0.9, 0.8))
	cat.SetAttentionValue(NewAttentionValue(100, 50, 5))

	data, err := space.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	space2 := NewDgraphSpace(ctx)
	if err := json.Unmarshal(data, space2); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}

	size, _ := space2.Size()
	if size != 3 {
		t.Errorf("expected 3 atoms after round-trip, got %d", size)
	}

	restored, err := space2.GetNodeByName(ConceptNodeType, "cat")
	if err != nil {
		t.Fatalf("GetNodeByName after round-trip: %v", err)
	}
	av := restored.GetAttentionValue()
	if av.STI != 100 {
		t.Errorf("expected STI 100, got %d", av.STI)
	}

	incoming, _ := space2.GetIncoming(restored)
	if len(incoming) != 1 {
		t.Errorf("expected 1 incoming link after round-trip, got %d", len(incoming))
	}
	tv := incoming[0].GetTruthValue()
	if tv.Strength != 0.9 || tv.Confidence != 0.8 {
		t.Errorf("unexpected truth value after round-trip: %+v", tv)
	}
}

func TestJSONRoundTripNumberNode(t *testing.T) {
	ctx := context.Background()
	space := NewDgraphSpace(ctx)
	_, err := space.AddNumberNode(2.718)
	if err != nil {
		t.Fatalf("AddNumberNode: %v", err)
	}

	data, err := space.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	space2 := NewDgraphSpace(ctx)
	if err := json.Unmarshal(data, space2); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}

	atoms := space2.QueryAtoms(ByType(NumberNodeType))
	if len(atoms) != 1 {
		t.Fatalf("expected 1 NumberNode after round-trip, got %d", len(atoms))
	}
	nn, ok := atoms[0].(*NumberNode)
	if !ok {
		t.Fatal("deserialized atom is not a *NumberNode")
	}
	if nn.Value != 2.718 {
		t.Errorf("expected Value 2.718, got %g", nn.Value)
	}
}

// ---------------------------------------------------------------------------
// Pattern matching tests
// ---------------------------------------------------------------------------

func TestQueryAtoms(t *testing.T) {
	ctx := context.Background()
	space := NewDgraphSpace(ctx)

	cat, _ := space.AddNode(ConceptNodeType, "cat")
	dog, _ := space.AddNode(ConceptNodeType, "dog")
	animal, _ := space.AddNode(ConceptNodeType, "animal")
	eats, _ := space.AddNode(PredicateNodeType, "eats")

	cat.SetAttentionValue(NewAttentionValue(120, 0, 0))
	dog.SetAttentionValue(NewAttentionValue(80, 0, 0))
	animal.SetAttentionValue(NewAttentionValue(50, 0, 0))
	eats.SetTruthValue(NewTruthValue(0.95, 0.9))

	// Filter by type
	concepts := space.QueryAtoms(ByType(ConceptNodeType))
	if len(concepts) != 3 {
		t.Errorf("expected 3 ConceptNodes, got %d", len(concepts))
	}

	// Filter by STI
	highSTI := space.QueryAtoms(ByMinSTI(100))
	if len(highSTI) != 1 {
		t.Errorf("expected 1 atom with STI >= 100, got %d", len(highSTI))
	}

	// Combined filter
	result := space.QueryAtoms(ByType(ConceptNodeType), ByMinSTI(90))
	if len(result) != 1 {
		t.Errorf("expected 1 ConceptNode with STI>=90, got %d", len(result))
	}
}

func TestFindLinks(t *testing.T) {
	ctx := context.Background()
	space := NewDgraphSpace(ctx)

	cat, _ := space.AddNode(ConceptNodeType, "cat")
	animal, _ := space.AddNode(ConceptNodeType, "animal")
	mammal, _ := space.AddNode(ConceptNodeType, "mammal")

	l1, _ := space.AddLink(InheritanceLinkType, []Atom{cat, animal})
	l2, _ := space.AddLink(InheritanceLinkType, []Atom{cat, mammal})
	l1.SetTruthValue(NewTruthValue(0.99, 0.95))
	l2.SetTruthValue(NewTruthValue(0.7, 0.3))

	// All links pointing to cat
	links := space.FindLinks(cat)
	if len(links) != 2 {
		t.Errorf("expected 2 links for cat, got %d", len(links))
	}

	// High-confidence links only
	confident := space.FindLinks(cat, ByMinConfidence(0.9))
	if len(confident) != 1 {
		t.Errorf("expected 1 high-confidence link, got %d", len(confident))
	}
}

// ---------------------------------------------------------------------------
// PLN inference tests
// ---------------------------------------------------------------------------

func TestPLNRevision(t *testing.T) {
	tv1 := NewTruthValue(0.8, 0.6)
	tv2 := NewTruthValue(0.2, 0.4)
	result := PLNRevision(tv1, tv2)

	expectedConf := 1.0 // clamped
	if result.Confidence != expectedConf {
		t.Errorf("expected confidence 1.0, got %f", result.Confidence)
	}
	expectedStrength := (0.8*0.6 + 0.2*0.4) / (0.6 + 0.4)
	if math.Abs(result.Strength-expectedStrength) > 1e-9 {
		t.Errorf("expected strength %f, got %f", expectedStrength, result.Strength)
	}
}

func TestPLNAnd(t *testing.T) {
	tv1 := NewTruthValue(0.8, 0.9)
	tv2 := NewTruthValue(0.5, 0.7)
	result := PLNAnd(tv1, tv2)

	if math.Abs(result.Strength-0.4) > 1e-9 {
		t.Errorf("expected strength 0.4, got %f", result.Strength)
	}
	if result.Confidence != 0.7 {
		t.Errorf("expected confidence 0.7, got %f", result.Confidence)
	}
}

func TestPLNOr(t *testing.T) {
	tv1 := NewTruthValue(0.8, 0.9)
	tv2 := NewTruthValue(0.5, 0.7)
	result := PLNOr(tv1, tv2)

	expected := 1 - (1-0.8)*(1-0.5)
	if math.Abs(result.Strength-expected) > 1e-9 {
		t.Errorf("expected strength %f, got %f", expected, result.Strength)
	}
}

func TestPLNNot(t *testing.T) {
	tv := NewTruthValue(0.7, 0.9)
	result := PLNNot(tv)
	if math.Abs(result.Strength-0.3) > 1e-9 {
		t.Errorf("expected strength 0.3, got %f", result.Strength)
	}
	if result.Confidence != 0.9 {
		t.Errorf("expected confidence 0.9, got %f", result.Confidence)
	}
}

func TestPLNDeduction(t *testing.T) {
	tvAB := NewTruthValue(0.9, 0.8)
	tvBC := NewTruthValue(0.8, 0.7)
	tvB := NewTruthValue(0.3, 0.6)

	result := PLNDeduction(tvAB, tvBC, tvB)
	if result.Strength < 0 || result.Strength > 1 {
		t.Errorf("deduction strength out of range: %f", result.Strength)
	}
	if result.Confidence != 0.7 {
		t.Errorf("expected confidence 0.7, got %f", result.Confidence)
	}
}

func TestPLNInversion(t *testing.T) {
	tvAB := NewTruthValue(0.9, 0.8)
	tvA := NewTruthValue(0.4, 0.7)
	tvB := NewTruthValue(0.6, 0.9)

	result := PLNInversion(tvAB, tvA, tvB)
	if result.Strength < 0 || result.Strength > 1 {
		t.Errorf("inversion strength out of range: %f", result.Strength)
	}
}

// ---------------------------------------------------------------------------
// ECAN tests
// ---------------------------------------------------------------------------

func TestECANDecay(t *testing.T) {
	ctx := context.Background()
	space := NewDgraphSpace(ctx)

	cat, _ := space.AddNode(ConceptNodeType, "cat")
	cat.SetAttentionValue(NewAttentionValue(100, 50, 10))

	cfg := DefaultECANConfig()
	space.ECANDecay(cfg)

	av := cat.GetAttentionValue()
	expectedSTI := int64(100 * (1 - cfg.STIDecayRate))
	if av.STI != expectedSTI {
		t.Errorf("expected STI %d after decay, got %d", expectedSTI, av.STI)
	}
}

func TestECANSpread(t *testing.T) {
	ctx := context.Background()
	space := NewDgraphSpace(ctx)

	cat, _ := space.AddNode(ConceptNodeType, "cat")
	animal, _ := space.AddNode(ConceptNodeType, "animal")
	link, _ := space.AddLink(InheritanceLinkType, []Atom{cat, animal})

	link.SetAttentionValue(NewAttentionValue(100, 0, 0))
	cat.SetAttentionValue(NewAttentionValue(0, 0, 0))
	animal.SetAttentionValue(NewAttentionValue(0, 0, 0))

	cfg := DefaultECANConfig()
	space.ECANSpread(cfg)

	linkAV := link.GetAttentionValue()
	if linkAV.STI >= 100 {
		t.Errorf("link STI should decrease after spread, got %d", linkAV.STI)
	}

	catAV := cat.GetAttentionValue()
	animalAV := animal.GetAttentionValue()
	if catAV.STI <= 0 && animalAV.STI <= 0 {
		t.Error("at least one neighbour should gain STI after spread")
	}
}

func TestECANTopN(t *testing.T) {
	ctx := context.Background()
	space := NewDgraphSpace(ctx)

	cat, _ := space.AddNode(ConceptNodeType, "cat")
	dog, _ := space.AddNode(ConceptNodeType, "dog")
	bird, _ := space.AddNode(ConceptNodeType, "bird")

	cat.SetAttentionValue(NewAttentionValue(300, 0, 0))
	dog.SetAttentionValue(NewAttentionValue(100, 0, 0))
	bird.SetAttentionValue(NewAttentionValue(200, 0, 0))

	top2 := space.ECANTopN(2)
	if len(top2) != 2 {
		t.Fatalf("expected 2 top atoms, got %d", len(top2))
	}
	if top2[0].GetAttentionValue().STI != 300 {
		t.Errorf("first top atom should have STI 300, got %d", top2[0].GetAttentionValue().STI)
	}
	if top2[1].GetAttentionValue().STI != 200 {
		t.Errorf("second top atom should have STI 200, got %d", top2[1].GetAttentionValue().STI)
	}
}

func TestECANStimulate(t *testing.T) {
	node := NewNode(ConceptNodeType, "x")
	node.SetAttentionValue(NewAttentionValue(50, 0, 0))

	cfg := DefaultECANConfig()
	ECANStimulate(node, 100, cfg)

	if node.GetAttentionValue().STI != 150 {
		t.Errorf("expected STI 150, got %d", node.GetAttentionValue().STI)
	}

	// Clamp to max
	ECANStimulate(node, 10000, cfg)
	if node.GetAttentionValue().STI != cfg.MaxSTI {
		t.Errorf("expected STI clamped to %d, got %d", cfg.MaxSTI, node.GetAttentionValue().STI)
	}
}
