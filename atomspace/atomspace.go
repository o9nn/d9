/*
 * SPDX-FileCopyrightText: © Hypermode Inc. <hello@hypermode.com>
 * SPDX-License-Identifier: Apache-2.0
 */

package atomspace

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// AtomSpace represents the hypergraph knowledge store using Dgraph as backend
// This implements the OpenCog AtomSpace concept using DgraphSpace
type AtomSpace interface {
	// AddNode adds a node atom to the space
	AddNode(atomType AtomType, name string) (Atom, error)
	
	// AddLink adds a link atom connecting other atoms
	AddLink(atomType AtomType, outgoing []Atom) (Atom, error)
	
	// GetAtom retrieves an atom by ID
	GetAtom(id string) (Atom, error)
	
	// GetNodeByName retrieves a node by type and name
	GetNodeByName(atomType AtomType, name string) (Atom, error)
	
	// GetIncoming returns all links that include the given atom in their outgoing set
	GetIncoming(atom Atom) ([]Atom, error)
	
	// GetAllAtoms returns all atoms of a given type
	GetAllAtoms(atomType AtomType) ([]Atom, error)
	
	// RemoveAtom removes an atom from the space
	RemoveAtom(id string) error
	
	// Clear removes all atoms from the space
	Clear() error
	
	// Size returns the total number of atoms in the space
	Size() (int, error)
}

// DgraphSpace implements AtomSpace using an in-memory store
// In a full implementation, this would interface with actual Dgraph storage
type DgraphSpace struct {
	mu      sync.RWMutex
	atoms   map[string]Atom      // ID -> Atom
	nodes   map[string]Atom      // "type:name" -> Node
	incoming map[string][]string // atom ID -> list of link IDs that include it
	ctx     context.Context
}

// NewDgraphSpace creates a new AtomSpace backed by Dgraph
func NewDgraphSpace(ctx context.Context) *DgraphSpace {
	return &DgraphSpace{
		atoms:    make(map[string]Atom),
		nodes:    make(map[string]Atom),
		incoming: make(map[string][]string),
		ctx:      ctx,
	}
}

// AddNode adds a node atom to the space
func (ds *DgraphSpace) AddNode(atomType AtomType, name string) (Atom, error) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// Check if node already exists
	key := generateID(atomType, name)
	if existing, ok := ds.nodes[key]; ok {
		return existing, nil
	}

	// Create new node
	node := NewNode(atomType, name)
	ds.atoms[node.GetID()] = node
	ds.nodes[key] = node

	return node, nil
}

// AddLink adds a link atom connecting other atoms
func (ds *DgraphSpace) AddLink(atomType AtomType, outgoing []Atom) (Atom, error) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// Verify all outgoing atoms exist in the space
	for _, atom := range outgoing {
		if _, exists := ds.atoms[atom.GetID()]; !exists {
			return nil, fmt.Errorf("outgoing atom %s not found in atomspace", atom.GetID())
		}
	}

	// Check if link already exists
	link := NewLink(atomType, outgoing)
	linkID := link.GetID()
	if existing, ok := ds.atoms[linkID]; ok {
		return existing, nil
	}

	// Add link to space
	ds.atoms[linkID] = link

	// Update incoming references for all outgoing atoms
	for _, atom := range outgoing {
		atomID := atom.GetID()
		ds.incoming[atomID] = append(ds.incoming[atomID], linkID)
	}

	return link, nil
}

// GetAtom retrieves an atom by ID
func (ds *DgraphSpace) GetAtom(id string) (Atom, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	atom, exists := ds.atoms[id]
	if !exists {
		return nil, fmt.Errorf("atom with id %s not found", id)
	}
	return atom, nil
}

// GetNodeByName retrieves a node by type and name
func (ds *DgraphSpace) GetNodeByName(atomType AtomType, name string) (Atom, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	key := generateID(atomType, name)
	node, exists := ds.nodes[key]
	if !exists {
		return nil, fmt.Errorf("node %s:%s not found", atomType, name)
	}
	return node, nil
}

// GetIncoming returns all links that include the given atom in their outgoing set
func (ds *DgraphSpace) GetIncoming(atom Atom) ([]Atom, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	linkIDs, exists := ds.incoming[atom.GetID()]
	if !exists {
		return []Atom{}, nil
	}

	incoming := make([]Atom, 0, len(linkIDs))
	for _, linkID := range linkIDs {
		if link, exists := ds.atoms[linkID]; exists {
			incoming = append(incoming, link)
		}
	}

	return incoming, nil
}

// GetAllAtoms returns all atoms of a given type
func (ds *DgraphSpace) GetAllAtoms(atomType AtomType) ([]Atom, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var result []Atom
	for _, atom := range ds.atoms {
		if atom.GetType() == atomType {
			result = append(result, atom)
		}
	}

	return result, nil
}

// RemoveAtom removes an atom from the space
func (ds *DgraphSpace) RemoveAtom(id string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	atom, exists := ds.atoms[id]
	if !exists {
		return fmt.Errorf("atom with id %s not found", id)
	}

	// Remove from atoms map
	delete(ds.atoms, id)

	// If it's a node, remove from nodes map
	if node, ok := atom.(*Node); ok {
		key := generateID(node.GetType(), node.Name)
		delete(ds.nodes, key)
	}

	// If it's a link, remove incoming references
	if link, ok := atom.(*Link); ok {
		for _, outgoing := range link.GetOutgoing() {
			atomID := outgoing.GetID()
			if links, exists := ds.incoming[atomID]; exists {
				// Remove this link from the incoming list
				newLinks := make([]string, 0, len(links)-1)
				for _, linkID := range links {
					if linkID != id {
						newLinks = append(newLinks, linkID)
					}
				}
				ds.incoming[atomID] = newLinks
			}
		}
	}

	// Remove any incoming references to this atom
	delete(ds.incoming, id)

	return nil
}

// Clear removes all atoms from the space
func (ds *DgraphSpace) Clear() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.atoms = make(map[string]Atom)
	ds.nodes = make(map[string]Atom)
	ds.incoming = make(map[string][]string)

	return nil
}

// Size returns the total number of atoms in the space
func (ds *DgraphSpace) Size() (int, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return len(ds.atoms), nil
}

// MarshalJSON serializes the AtomSpace to JSON
func (ds *DgraphSpace) MarshalJSON() ([]byte, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	type atomData struct {
		ID           string                 `json:"id"`
		Type         string                 `json:"type"`
		Name         string                 `json:"name,omitempty"`
		Outgoing     []string               `json:"outgoing,omitempty"`
		TruthValue   *TruthValue            `json:"truthValue"`
		AttentionValue *AttentionValue      `json:"attentionValue"`
	}

	atoms := make([]atomData, 0, len(ds.atoms))
	for _, atom := range ds.atoms {
		data := atomData{
			ID:             atom.GetID(),
			Type:           string(atom.GetType()),
			TruthValue:     atom.GetTruthValue(),
			AttentionValue: atom.GetAttentionValue(),
		}

		if node, ok := atom.(*Node); ok {
			data.Name = node.Name
		} else if link, ok := atom.(*Link); ok {
			outgoing := make([]string, len(link.Outgoing))
			for i, out := range link.Outgoing {
				outgoing[i] = out.GetID()
			}
			data.Outgoing = outgoing
		}

		atoms = append(atoms, data)
	}

	return json.Marshal(map[string]interface{}{
		"atoms": atoms,
		"size":  len(ds.atoms),
	})
}
