# AtomSpace Integration with Dgraph

This document describes how the AtomSpace implementation integrates with the Dgraph ecosystem and OpenCog principles.

## Architecture Overview

### AtomSpace as DgraphSpace

The implementation provides a hypergraph knowledge representation layer on top of Dgraph's graph database infrastructure:

```
┌─────────────────────────────────────────────┐
│         AtomSpace API Layer                 │
│  (Nodes, Links, TruthValues, Attention)    │
├─────────────────────────────────────────────┤
│         DgraphSpace Implementation          │
│  (In-memory storage with Dgraph semantics) │
├─────────────────────────────────────────────┤
│      Future: Dgraph Backend Integration     │
│  (Badger storage, distributed queries)     │
└─────────────────────────────────────────────┘
```

## OpenCog CogPrime Integration

This implementation aligns with the OpenCog CogPrime cognitive architecture described in `.github/agents/cogprime.md`:

### 1. Knowledge Representation (AtomSpace)

**From CogPrime:**
> AtomSpace: Hypergraph knowledge representation for flexible associations

**Implementation:**
- Nodes represent concepts, predicates, and other terminal entities
- Links represent relationships between atoms (inheritance, similarity, etc.)
- Hypergraph structure allows arbitrary n-ary relationships

### 2. Uncertain Reasoning (PLN)

**From CogPrime:**
> PLN: Probabilistic Logic Networks for reasoning under uncertainty

**Implementation:**
```go
type TruthValue struct {
    Strength   float64 // Probability/confidence [0,1]
    Confidence float64 // Weight of evidence [0,1]
}
```

Truth values enable:
- Probabilistic inference
- Uncertainty quantification
- Evidence accumulation
- Strength-confidence reasoning

### 3. Attention Allocation (ECAN)

**From CogPrime:**
> ECAN: Economic Attention Network for dynamic resource allocation

**Implementation:**
```go
type AttentionValue struct {
    STI  int64 // Short-term importance
    LTI  int64 // Long-term importance
    VLTI int64 // Very long-term importance
}
```

Attention values enable:
- Resource prioritization
- Working memory management
- Focus of processing
- Importance-based filtering

## Cognitive Synergy

The AtomSpace implementation supports cognitive synergy through:

### Multi-Modal Knowledge Representation

```go
// Declarative: Concepts and relationships
cat := space.AddNode(ConceptNodeType, "cat")
animal := space.AddNode(ConceptNodeType, "animal")
inheritance := space.AddLink(InheritanceLinkType, []Atom{cat, animal})

// Procedural: Action predicates
eats := space.AddNode(PredicateNodeType, "eats")
evaluation := space.AddLink(EvaluationLinkType, []Atom{eats, cat, meat})
```

### Emergent Network Structure

The implementation supports emergent hierarchical and heterarchical networks:

- **Hierarchical**: Inheritance hierarchies (cat → mammal → animal)
- **Heterarchical**: Cross-cutting relationships (similarity, association)
- **Self-Network**: Atoms can reference the system itself
- **Mirror Networks**: Atoms can model other agents

## Integration Points

### Current Implementation

The current implementation provides:

1. **In-Memory Storage**: Fast, thread-safe atom storage
2. **CRUD Operations**: Create, Read, Update, Delete atoms
3. **Query Capabilities**: Type-based queries, incoming links
4. **Serialization**: JSON export for persistence/interchange

### Future Dgraph Backend Integration

For full integration with Dgraph's distributed storage:

```go
// Future: DgraphSpace with actual Dgraph backend
type DgraphSpace struct {
    client *dgo.Dgraph
    schema *schema.Schema
    // ... existing fields
}

// Future: Persist atoms to Dgraph
func (ds *DgraphSpace) AddNode(atomType AtomType, name string) (Atom, error) {
    // Create Dgraph mutation
    mu := &api.Mutation{
        SetNquads: []byte(fmt.Sprintf(`
            _:atom <dgraph.type> "Atom" .
            _:atom <atom.type> "%s" .
            _:atom <atom.name> "%s" .
            _:atom <atom.truthValue> "%.2f,%.2f" .
        `, atomType, name, tv.Strength, tv.Confidence)),
    }
    
    // Execute mutation
    resp, err := ds.client.NewTxn().Mutate(ctx, mu)
    // ...
}
```

### Dgraph Schema for AtomSpace

Future schema definition:

```graphql
type Atom {
  id: ID!
  type: String! @index(exact)
  name: String @index(fulltext)
  outgoing: [Atom]
  truthStrength: Float
  truthConfidence: Float
  stiValue: Int
  ltiValue: Int
  vltiValue: Int
}
```

## Usage Patterns

### Knowledge Base Construction

```go
space := atomspace.NewDgraphSpace(ctx)

// Build taxonomy
human := space.AddNode(ConceptNodeType, "human")
mammal := space.AddNode(ConceptNodeType, "mammal")
animal := space.AddNode(ConceptNodeType, "animal")

space.AddLink(InheritanceLinkType, []Atom{human, mammal})
space.AddLink(InheritanceLinkType, []Atom{mammal, animal})
```

### Probabilistic Reasoning

```go
// Assert with uncertainty
link := space.AddLink(InheritanceLinkType, []Atom{cat, mammal})
link.SetTruthValue(NewTruthValue(0.99, 0.95))

// Query with confidence
if link.GetTruthValue().Strength > 0.9 {
    // High-confidence assertion
}
```

### Attention-Based Processing

```go
// Set importance
importantAtom.SetAttentionValue(NewAttentionValue(150, 75, 20))

// Filter by attention
concepts, _ := space.GetAllAtoms(ConceptNodeType)
for _, atom := range concepts {
    if atom.GetAttentionValue().STI > 100 {
        // Process high-importance atoms
    }
}
```

## Compatibility with OpenCog

This implementation maintains compatibility with OpenCog concepts:

| OpenCog Concept | DgraphSpace Implementation |
|-----------------|---------------------------|
| Atom | Atom interface |
| Node | Node struct |
| Link | Link struct |
| TruthValue | TruthValue struct |
| AttentionValue | AttentionValue struct |
| AtomSpace | DgraphSpace struct |
| getIncoming() | GetIncoming() method |
| addNode() | AddNode() method |
| addLink() | AddLink() method |

## Performance Considerations

### Current Implementation

- Thread-safe with RWMutex
- O(1) atom lookup by ID
- O(1) node lookup by type:name
- O(n) type-based queries
- O(1) incoming link tracking

### Future Optimizations

With Dgraph backend:
- Distributed storage and queries
- Index-based fast lookups
- GraphQL query language
- ACID transactions
- Horizontal scaling

## References

- [OpenCog AtomSpace](https://wiki.opencog.org/w/AtomSpace)
- [CogPrime Architecture](https://wiki.opencog.org/w/CogPrime_Overview)
- [PLN Theory](https://wiki.opencog.org/w/PLN)
- [ECAN Theory](https://wiki.opencog.org/w/ECAN)
- [Dgraph Documentation](https://docs.hypermode.com/dgraph)
