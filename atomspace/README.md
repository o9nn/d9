# AtomSpace as DgraphSpace

This package implements OpenCog's AtomSpace hypergraph knowledge representation system using Dgraph as the underlying storage backend.

## Overview

AtomSpace is a hypergraph knowledge representation framework from the OpenCog project. This implementation provides:

- **Atoms**: Fundamental units representing nodes (concepts) and links (relationships)
- **TruthValues**: Probabilistic logic values for uncertain reasoning (PLN)
- **AttentionValues**: Support for Economic Attention Networks (ECAN)
- **Hypergraph Operations**: Add, retrieve, query, and manipulate atoms

## Core Concepts

### Atoms

Atoms are the basic building blocks:

- **Nodes**: Terminal atoms representing concepts, predicates, etc.
- **Links**: Connections between atoms forming relationships

### Atom Types

Predefined atom types include:

- `ConceptNode`: Represents concepts (e.g., "cat", "animal")
- `PredicateNode`: Represents predicates (e.g., "eats", "loves")
- `InheritanceLink`: Represents inheritance relationships
- `SimilarityLink`: Represents similarity relationships
- `EvaluationLink`: Represents predicate evaluations

### Truth Values

Each atom has a `TruthValue` with:

- **Strength**: Probability/confidence [0, 1]
- **Confidence**: Weight of evidence [0, 1]

This supports Probabilistic Logic Networks (PLN) for uncertain reasoning.

### Attention Values

Each atom has an `AttentionValue` with:

- **STI**: Short-term importance
- **LTI**: Long-term importance
- **VLTI**: Very long-term importance

This supports Economic Attention Networks (ECAN) for resource allocation.

## Usage

### Creating an AtomSpace

```go
import (
    "context"
    "github.com/hypermodeinc/dgraph/v25/atomspace"
)

ctx := context.Background()
space := atomspace.NewDgraphSpace(ctx)
```

### Adding Nodes

```go
cat, _ := space.AddNode(atomspace.ConceptNodeType, "cat")
animal, _ := space.AddNode(atomspace.ConceptNodeType, "animal")
```

### Adding Links

```go
// Create: (InheritanceLink (ConceptNode "cat") (ConceptNode "animal"))
link, _ := space.AddLink(
    atomspace.InheritanceLinkType,
    []atomspace.Atom{cat, animal},
)
```

### Setting Truth Values

```go
// Set truth value: strength=0.9, confidence=0.8
tv := atomspace.NewTruthValue(0.9, 0.8)
link.SetTruthValue(tv)
```

### Setting Attention Values

```go
// Set attention: STI=100, LTI=50, VLTI=10
av := atomspace.NewAttentionValue(100, 50, 10)
cat.SetAttentionValue(av)
```

### Querying

```go
// Get atom by ID
atom, _ := space.GetAtom(cat.GetID())

// Get node by type and name
node, _ := space.GetNodeByName(atomspace.ConceptNodeType, "cat")

// Get all atoms of a type
concepts, _ := space.GetAllAtoms(atomspace.ConceptNodeType)

// Get incoming links (backlinks)
incoming, _ := space.GetIncoming(animal)
```

### Complete Example

```go
package main

import (
    "context"
    "fmt"
    "github.com/hypermodeinc/dgraph/v25/atomspace"
)

func main() {
    ctx := context.Background()
    space := atomspace.NewDgraphSpace(ctx)
    
    // Create knowledge: cat and dog are animals, and they are similar
    cat, _ := space.AddNode(atomspace.ConceptNodeType, "cat")
    dog, _ := space.AddNode(atomspace.ConceptNodeType, "dog")
    animal, _ := space.AddNode(atomspace.ConceptNodeType, "animal")
    
    // Add inheritance relationships
    catIsAnimal, _ := space.AddLink(
        atomspace.InheritanceLinkType,
        []atomspace.Atom{cat, animal},
    )
    dogIsAnimal, _ := space.AddLink(
        atomspace.InheritanceLinkType,
        []atomspace.Atom{dog, animal},
    )
    
    // Add similarity relationship
    similarity, _ := space.AddLink(
        atomspace.SimilarityLinkType,
        []atomspace.Atom{cat, dog},
    )
    
    // Set truth values
    catIsAnimal.SetTruthValue(atomspace.NewTruthValue(0.95, 0.9))
    dogIsAnimal.SetTruthValue(atomspace.NewTruthValue(0.95, 0.9))
    similarity.SetTruthValue(atomspace.NewTruthValue(0.7, 0.6))
    
    // Set attention for important concepts
    cat.SetAttentionValue(atomspace.NewAttentionValue(100, 50, 10))
    
    // Query the knowledge base
    size, _ := space.Size()
    fmt.Printf("AtomSpace contains %d atoms\n", size)
    
    incoming, _ := space.GetIncoming(animal)
    fmt.Printf("Animal has %d incoming links\n", len(incoming))
    
    for _, link := range incoming {
        fmt.Printf("  %s\n", link.String())
    }
}
```

## Architecture

### DgraphSpace Structure

The `DgraphSpace` implementation uses:

- **atoms map**: ID → Atom mapping for fast lookups
- **nodes map**: "type:name" → Node mapping for node queries
- **incoming map**: Atom ID → Link IDs for backlink tracking
- **Mutex protection**: Thread-safe operations

### Future Enhancements

In a full Dgraph integration, the implementation would:

1. Use Dgraph mutations for atom storage
2. Leverage Dgraph's GraphQL/DQL for pattern matching
3. Implement distributed storage and querying
4. Add persistent storage with Badger backend
5. Support ACID transactions for atomic updates

## OpenCog Compatibility

This implementation follows OpenCog AtomSpace principles:

- Hypergraph knowledge representation
- Probabilistic truth values (PLN)
- Attention allocation (ECAN)
- Node and link atom types
- Incoming set tracking

## Testing

Run the test suite:

```bash
cd atomspace
go test -v
```

## References

- [OpenCog AtomSpace](https://wiki.opencog.org/w/AtomSpace)
- [OpenCog Prime](https://wiki.opencog.org/w/CogPrime_Overview)
- [Probabilistic Logic Networks (PLN)](https://wiki.opencog.org/w/PLN)
- [Economic Attention Networks (ECAN)](https://wiki.opencog.org/w/ECAN)
- [Dgraph Documentation](https://docs.hypermode.com/dgraph)
