/*
 * SPDX-FileCopyrightText: © Hypermode Inc. <hello@hypermode.com>
 * SPDX-License-Identifier: Apache-2.0
 */

package atomspace

import (
	"fmt"
)

// AtomType represents the type of an Atom in the AtomSpace
type AtomType string

const (
	// NodeType represents a node atom (terminal/leaf)
	NodeType AtomType = "Node"
	// LinkType represents a link atom (connects other atoms)
	LinkType AtomType = "Link"

	// --- Node types ---

	// ConceptNodeType represents a concept node
	ConceptNodeType AtomType = "ConceptNode"
	// PredicateNodeType represents a predicate node
	PredicateNodeType AtomType = "PredicateNode"
	// NumberNodeType represents a numeric constant node
	NumberNodeType AtomType = "NumberNode"
	// VariableNodeType represents an unbound variable in patterns
	VariableNodeType AtomType = "VariableNode"
	// SchemaNodeType represents a schema (procedural) node
	SchemaNodeType AtomType = "SchemaNode"
	// GroundedSchemaNodeType represents an executable (grounded) schema node
	GroundedSchemaNodeType AtomType = "GroundedSchemaNode"

	// --- Link types ---

	// InheritanceLinkType represents an inheritance relationship
	InheritanceLinkType AtomType = "InheritanceLink"
	// SimilarityLinkType represents a similarity relationship
	SimilarityLinkType AtomType = "SimilarityLink"
	// EvaluationLinkType represents a predicate evaluation
	EvaluationLinkType AtomType = "EvaluationLink"
	// MemberLinkType represents set membership
	MemberLinkType AtomType = "MemberLink"
	// ListLinkType represents an ordered list of atoms
	ListLinkType AtomType = "ListLink"
	// SetLinkType represents an unordered set of atoms
	SetLinkType AtomType = "SetLink"
	// AndLinkType represents logical conjunction
	AndLinkType AtomType = "AndLink"
	// OrLinkType represents logical disjunction
	OrLinkType AtomType = "OrLink"
	// NotLinkType represents logical negation (unary)
	NotLinkType AtomType = "NotLink"
	// ImplicationLinkType represents logical implication (if A then B)
	ImplicationLinkType AtomType = "ImplicationLink"
	// EquivalenceLinkType represents logical equivalence (A iff B)
	EquivalenceLinkType AtomType = "EquivalenceLink"
	// ExecutionLinkType represents execution of a schema
	ExecutionLinkType AtomType = "ExecutionLink"
	// ContextLinkType represents a contextual assertion
	ContextLinkType AtomType = "ContextLink"
)

// TruthValue represents the truth value of an Atom for uncertain reasoning
// Following OpenCog's PLN (Probabilistic Logic Networks) model
type TruthValue struct {
	Strength   float64 // Probability/confidence [0,1]
	Confidence float64 // Weight of evidence [0,1]
}

// NewTruthValue creates a new TruthValue with the given strength and confidence
func NewTruthValue(strength, confidence float64) *TruthValue {
	return &TruthValue{
		Strength:   clamp(strength, 0.0, 1.0),
		Confidence: clamp(confidence, 0.0, 1.0),
	}
}

// DefaultTruthValue returns a default truth value (0.5 strength, 0.0 confidence)
func DefaultTruthValue() *TruthValue {
	return &TruthValue{
		Strength:   0.5,
		Confidence: 0.0,
	}
}

// clamp restricts a value to the given range
func clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// AttentionValue represents the attention allocation for ECAN
type AttentionValue struct {
	STI  int64 // Short-term importance
	LTI  int64 // Long-term importance
	VLTI int64 // Very long-term importance
}

// NewAttentionValue creates a new AttentionValue
func NewAttentionValue(sti, lti, vlti int64) *AttentionValue {
	return &AttentionValue{
		STI:  sti,
		LTI:  lti,
		VLTI: vlti,
	}
}

// DefaultAttentionValue returns a default attention value
func DefaultAttentionValue() *AttentionValue {
	return &AttentionValue{
		STI:  0,
		LTI:  0,
		VLTI: 0,
	}
}

// Atom represents a fundamental unit in the AtomSpace hypergraph
// Can be either a Node (terminal) or Link (connecting other atoms)
type Atom interface {
	// GetID returns the unique identifier for this atom
	GetID() string
	// GetType returns the type of this atom
	GetType() AtomType
	// GetTruthValue returns the truth value
	GetTruthValue() *TruthValue
	// SetTruthValue sets the truth value
	SetTruthValue(*TruthValue)
	// GetAttentionValue returns the attention value
	GetAttentionValue() *AttentionValue
	// SetAttentionValue sets the attention value
	SetAttentionValue(*AttentionValue)
	// String returns a string representation
	String() string
}

// BaseAtom provides common functionality for all atoms
type BaseAtom struct {
	ID           string
	Type         AtomType
	TruthVal     *TruthValue
	AttentionVal *AttentionValue
}

// GetID returns the atom's unique identifier
func (a *BaseAtom) GetID() string {
	return a.ID
}

// GetType returns the atom's type
func (a *BaseAtom) GetType() AtomType {
	return a.Type
}

// GetTruthValue returns the atom's truth value
func (a *BaseAtom) GetTruthValue() *TruthValue {
	if a.TruthVal == nil {
		return DefaultTruthValue()
	}
	return a.TruthVal
}

// SetTruthValue sets the atom's truth value
func (a *BaseAtom) SetTruthValue(tv *TruthValue) {
	a.TruthVal = tv
}

// GetAttentionValue returns the atom's attention value
func (a *BaseAtom) GetAttentionValue() *AttentionValue {
	if a.AttentionVal == nil {
		return DefaultAttentionValue()
	}
	return a.AttentionVal
}

// SetAttentionValue sets the atom's attention value
func (a *BaseAtom) SetAttentionValue(av *AttentionValue) {
	a.AttentionVal = av
}

// Node represents a terminal atom (leaf node in the hypergraph)
type Node struct {
	BaseAtom
	Name string
}

// NewNode creates a new Node atom
func NewNode(atomType AtomType, name string) *Node {
	return &Node{
		BaseAtom: BaseAtom{
			ID:           generateID(atomType, name),
			Type:         atomType,
			TruthVal:     DefaultTruthValue(),
			AttentionVal: DefaultAttentionValue(),
		},
		Name: name,
	}
}

// String returns a string representation of the Node
func (n *Node) String() string {
	return fmt.Sprintf("(%s \"%s\")", n.Type, n.Name)
}

// Link represents a connection between atoms (edge in the hypergraph)
type Link struct {
	BaseAtom
	Outgoing []Atom // The atoms this link connects
}

// NewLink creates a new Link atom
func NewLink(atomType AtomType, outgoing []Atom) *Link {
	return &Link{
		BaseAtom: BaseAtom{
			ID:           generateLinkID(atomType, outgoing),
			Type:         atomType,
			TruthVal:     DefaultTruthValue(),
			AttentionVal: DefaultAttentionValue(),
		},
		Outgoing: outgoing,
	}
}

// String returns a string representation of the Link
func (l *Link) String() string {
	outgoingStr := ""
	for i, atom := range l.Outgoing {
		if i > 0 {
			outgoingStr += " "
		}
		outgoingStr += atom.String()
	}
	return fmt.Sprintf("(%s %s)", l.Type, outgoingStr)
}

// GetOutgoing returns the atoms connected by this link
func (l *Link) GetOutgoing() []Atom {
	return l.Outgoing
}

// NumberNode is a Node that also carries a float64 value.
// Its Name is the canonical decimal representation of Value.
type NumberNode struct {
	Node
	Value float64
}

// NewNumberNode creates a NumberNode for the given numeric value.
func NewNumberNode(value float64) *NumberNode {
	name := fmt.Sprintf("%g", value)
	return &NumberNode{
		Node: Node{
			BaseAtom: BaseAtom{
				ID:           generateID(NumberNodeType, name),
				Type:         NumberNodeType,
				TruthVal:     DefaultTruthValue(),
				AttentionVal: DefaultAttentionValue(),
			},
			Name: name,
		},
		Value: value,
	}
}

// String returns a string representation of the NumberNode.
func (n *NumberNode) String() string {
	return fmt.Sprintf("(%s %g)", n.Type, n.Value)
}

// IsNode returns true when the atom is a Node (terminal).
func IsNode(a Atom) bool {
	switch a.(type) {
	case *Node, *NumberNode:
		return true
	}
	return false
}

// IsLink returns true when the atom is a Link.
func IsLink(a Atom) bool {
	_, ok := a.(*Link)
	return ok
}

// generateID generates a unique ID for a node
func generateID(atomType AtomType, name string) string {
	return fmt.Sprintf("%s:%s", atomType, name)
}

// generateLinkID generates a unique ID for a link based on its outgoing atoms
func generateLinkID(atomType AtomType, outgoing []Atom) string {
	id := string(atomType) + ":"
	for i, atom := range outgoing {
		if i > 0 {
			id += ","
		}
		id += atom.GetID()
	}
	return id
}
