/*
 * SPDX-FileCopyrightText: © Hypermode Inc. <hello@hypermode.com>
 * SPDX-License-Identifier: Apache-2.0
 */

package atomspace

import (
	"context"
	"testing"
)

func TestNewNode(t *testing.T) {
	node := NewNode(ConceptNodeType, "cat")
	
	if node.Name != "cat" {
		t.Errorf("Expected name 'cat', got '%s'", node.Name)
	}
	
	if node.GetType() != ConceptNodeType {
		t.Errorf("Expected type ConceptNodeType, got %s", node.GetType())
	}
	
	if node.GetID() != "ConceptNode:cat" {
		t.Errorf("Expected ID 'ConceptNode:cat', got '%s'", node.GetID())
	}
}

func TestNewLink(t *testing.T) {
	cat := NewNode(ConceptNodeType, "cat")
	animal := NewNode(ConceptNodeType, "animal")
	
	link := NewLink(InheritanceLinkType, []Atom{cat, animal})
	
	if link.GetType() != InheritanceLinkType {
		t.Errorf("Expected type InheritanceLinkType, got %s", link.GetType())
	}
	
	if len(link.Outgoing) != 2 {
		t.Errorf("Expected 2 outgoing atoms, got %d", len(link.Outgoing))
	}
	
	if link.Outgoing[0].GetID() != cat.GetID() {
		t.Errorf("Expected first outgoing to be cat")
	}
}

func TestTruthValue(t *testing.T) {
	tv := NewTruthValue(0.8, 0.9)
	
	if tv.Strength != 0.8 {
		t.Errorf("Expected strength 0.8, got %f", tv.Strength)
	}
	
	if tv.Confidence != 0.9 {
		t.Errorf("Expected confidence 0.9, got %f", tv.Confidence)
	}
	
	// Test clamping
	tv2 := NewTruthValue(1.5, -0.2)
	if tv2.Strength != 1.0 {
		t.Errorf("Expected clamped strength 1.0, got %f", tv2.Strength)
	}
	if tv2.Confidence != 0.0 {
		t.Errorf("Expected clamped confidence 0.0, got %f", tv2.Confidence)
	}
}

func TestAttentionValue(t *testing.T) {
	av := NewAttentionValue(100, 50, 10)
	
	if av.STI != 100 {
		t.Errorf("Expected STI 100, got %d", av.STI)
	}
	
	if av.LTI != 50 {
		t.Errorf("Expected LTI 50, got %d", av.LTI)
	}
	
	if av.VLTI != 10 {
		t.Errorf("Expected VLTI 10, got %d", av.VLTI)
	}
}

func TestDgraphSpaceAddNode(t *testing.T) {
	ctx := context.Background()
	space := NewDgraphSpace(ctx)
	
	node, err := space.AddNode(ConceptNodeType, "dog")
	if err != nil {
		t.Fatalf("Failed to add node: %v", err)
	}
	
	if node.GetType() != ConceptNodeType {
		t.Errorf("Expected ConceptNodeType, got %s", node.GetType())
	}
	
	// Adding same node again should return existing
	node2, err := space.AddNode(ConceptNodeType, "dog")
	if err != nil {
		t.Fatalf("Failed to add duplicate node: %v", err)
	}
	
	if node.GetID() != node2.GetID() {
		t.Errorf("Expected same node ID for duplicate add")
	}
	
	size, _ := space.Size()
	if size != 1 {
		t.Errorf("Expected size 1, got %d", size)
	}
}

func TestDgraphSpaceAddLink(t *testing.T) {
	ctx := context.Background()
	space := NewDgraphSpace(ctx)
	
	cat, _ := space.AddNode(ConceptNodeType, "cat")
	animal, _ := space.AddNode(ConceptNodeType, "animal")
	
	link, err := space.AddLink(InheritanceLinkType, []Atom{cat, animal})
	if err != nil {
		t.Fatalf("Failed to add link: %v", err)
	}
	
	if link.GetType() != InheritanceLinkType {
		t.Errorf("Expected InheritanceLinkType, got %s", link.GetType())
	}
	
	size, _ := space.Size()
	if size != 3 { // 2 nodes + 1 link
		t.Errorf("Expected size 3, got %d", size)
	}
}

func TestDgraphSpaceGetAtom(t *testing.T) {
	ctx := context.Background()
	space := NewDgraphSpace(ctx)
	
	node, _ := space.AddNode(ConceptNodeType, "bird")
	
	retrieved, err := space.GetAtom(node.GetID())
	if err != nil {
		t.Fatalf("Failed to get atom: %v", err)
	}
	
	if retrieved.GetID() != node.GetID() {
		t.Errorf("Retrieved wrong atom")
	}
	
	_, err = space.GetAtom("nonexistent")
	if err == nil {
		t.Errorf("Expected error for nonexistent atom")
	}
}

func TestDgraphSpaceGetNodeByName(t *testing.T) {
	ctx := context.Background()
	space := NewDgraphSpace(ctx)
	
	space.AddNode(ConceptNodeType, "fish")
	
	node, err := space.GetNodeByName(ConceptNodeType, "fish")
	if err != nil {
		t.Fatalf("Failed to get node by name: %v", err)
	}
	
	if n, ok := node.(*Node); ok {
		if n.Name != "fish" {
			t.Errorf("Expected name 'fish', got '%s'", n.Name)
		}
	} else {
		t.Errorf("Retrieved atom is not a Node")
	}
	
	_, err = space.GetNodeByName(ConceptNodeType, "nonexistent")
	if err == nil {
		t.Errorf("Expected error for nonexistent node")
	}
}

func TestDgraphSpaceGetIncoming(t *testing.T) {
	ctx := context.Background()
	space := NewDgraphSpace(ctx)
	
	cat, _ := space.AddNode(ConceptNodeType, "cat")
	animal, _ := space.AddNode(ConceptNodeType, "animal")
	mammal, _ := space.AddNode(ConceptNodeType, "mammal")
	
	// cat inherits from animal
	link1, _ := space.AddLink(InheritanceLinkType, []Atom{cat, animal})
	// cat inherits from mammal
	link2, _ := space.AddLink(InheritanceLinkType, []Atom{cat, mammal})
	
	incoming, err := space.GetIncoming(cat)
	if err != nil {
		t.Fatalf("Failed to get incoming: %v", err)
	}
	
	if len(incoming) != 2 {
		t.Errorf("Expected 2 incoming links, got %d", len(incoming))
	}
	
	// Check both links are present
	found := 0
	for _, link := range incoming {
		if link.GetID() == link1.GetID() || link.GetID() == link2.GetID() {
			found++
		}
	}
	
	if found != 2 {
		t.Errorf("Expected to find both links in incoming set")
	}
}

func TestDgraphSpaceGetAllAtoms(t *testing.T) {
	ctx := context.Background()
	space := NewDgraphSpace(ctx)
	
	space.AddNode(ConceptNodeType, "dog")
	space.AddNode(ConceptNodeType, "cat")
	space.AddNode(PredicateNodeType, "eats")
	
	concepts, err := space.GetAllAtoms(ConceptNodeType)
	if err != nil {
		t.Fatalf("Failed to get all atoms: %v", err)
	}
	
	if len(concepts) != 2 {
		t.Errorf("Expected 2 ConceptNodes, got %d", len(concepts))
	}
	
	predicates, _ := space.GetAllAtoms(PredicateNodeType)
	if len(predicates) != 1 {
		t.Errorf("Expected 1 PredicateNode, got %d", len(predicates))
	}
}

func TestDgraphSpaceRemoveAtom(t *testing.T) {
	ctx := context.Background()
	space := NewDgraphSpace(ctx)
	
	node, _ := space.AddNode(ConceptNodeType, "temporary")
	
	size, _ := space.Size()
	if size != 1 {
		t.Errorf("Expected size 1 before removal, got %d", size)
	}
	
	err := space.RemoveAtom(node.GetID())
	if err != nil {
		t.Fatalf("Failed to remove atom: %v", err)
	}
	
	size, _ = space.Size()
	if size != 0 {
		t.Errorf("Expected size 0 after removal, got %d", size)
	}
	
	_, err = space.GetAtom(node.GetID())
	if err == nil {
		t.Errorf("Expected error when getting removed atom")
	}
}

func TestDgraphSpaceRemoveLink(t *testing.T) {
	ctx := context.Background()
	space := NewDgraphSpace(ctx)
	
	cat, _ := space.AddNode(ConceptNodeType, "cat")
	animal, _ := space.AddNode(ConceptNodeType, "animal")
	link, _ := space.AddLink(InheritanceLinkType, []Atom{cat, animal})
	
	// Check incoming before removal
	incoming, _ := space.GetIncoming(cat)
	if len(incoming) != 1 {
		t.Errorf("Expected 1 incoming link before removal")
	}
	
	err := space.RemoveAtom(link.GetID())
	if err != nil {
		t.Fatalf("Failed to remove link: %v", err)
	}
	
	// Check incoming after removal
	incoming, _ = space.GetIncoming(cat)
	if len(incoming) != 0 {
		t.Errorf("Expected 0 incoming links after removal, got %d", len(incoming))
	}
	
	size, _ := space.Size()
	if size != 2 { // Only the 2 nodes should remain
		t.Errorf("Expected size 2 after link removal, got %d", size)
	}
}

func TestDgraphSpaceClear(t *testing.T) {
	ctx := context.Background()
	space := NewDgraphSpace(ctx)
	
	space.AddNode(ConceptNodeType, "cat")
	space.AddNode(ConceptNodeType, "dog")
	space.AddNode(ConceptNodeType, "bird")
	
	size, _ := space.Size()
	if size != 3 {
		t.Errorf("Expected size 3 before clear")
	}
	
	err := space.Clear()
	if err != nil {
		t.Fatalf("Failed to clear: %v", err)
	}
	
	size, _ = space.Size()
	if size != 0 {
		t.Errorf("Expected size 0 after clear, got %d", size)
	}
}

func TestComplexHypergraph(t *testing.T) {
	ctx := context.Background()
	space := NewDgraphSpace(ctx)
	
	// Create a simple knowledge base:
	// (Inheritance (Concept "cat") (Concept "animal"))
	// (Inheritance (Concept "dog") (Concept "animal"))
	// (Similarity (Concept "cat") (Concept "dog"))
	
	cat, _ := space.AddNode(ConceptNodeType, "cat")
	dog, _ := space.AddNode(ConceptNodeType, "dog")
	animal, _ := space.AddNode(ConceptNodeType, "animal")
	
	link1, _ := space.AddLink(InheritanceLinkType, []Atom{cat, animal})
	link2, _ := space.AddLink(InheritanceLinkType, []Atom{dog, animal})
	link3, _ := space.AddLink(SimilarityLinkType, []Atom{cat, dog})
	
	// Set truth values
	link1.SetTruthValue(NewTruthValue(0.9, 0.8))
	link2.SetTruthValue(NewTruthValue(0.9, 0.8))
	link3.SetTruthValue(NewTruthValue(0.7, 0.6))
	
	// Set attention values
	cat.SetAttentionValue(NewAttentionValue(100, 50, 10))
	dog.SetAttentionValue(NewAttentionValue(90, 45, 8))
	
	size, _ := space.Size()
	if size != 6 { // 3 nodes + 3 links
		t.Errorf("Expected size 6, got %d", size)
	}
	
	// Test incoming links for animal
	incoming, _ := space.GetIncoming(animal)
	if len(incoming) != 2 {
		t.Errorf("Expected 2 inheritance links to animal, got %d", len(incoming))
	}
	
	// Test truth values
	tv := link1.GetTruthValue()
	if tv.Strength != 0.9 || tv.Confidence != 0.8 {
		t.Errorf("Truth value not preserved correctly")
	}
	
	// Test attention values
	av := cat.GetAttentionValue()
	if av.STI != 100 {
		t.Errorf("Attention value not preserved correctly")
	}
}

func TestMarshalJSON(t *testing.T) {
	ctx := context.Background()
	space := NewDgraphSpace(ctx)
	
	cat, _ := space.AddNode(ConceptNodeType, "cat")
	animal, _ := space.AddNode(ConceptNodeType, "animal")
	space.AddLink(InheritanceLinkType, []Atom{cat, animal})
	
	data, err := space.MarshalJSON()
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}
	
	if len(data) == 0 {
		t.Errorf("Expected non-empty JSON data")
	}
}
