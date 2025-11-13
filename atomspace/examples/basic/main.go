/*
 * SPDX-FileCopyrightText: © Hypermode Inc. <hello@hypermode.com>
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hypermodeinc/dgraph/v25/atomspace"
)

func main() {
	fmt.Println("=== AtomSpace as DgraphSpace Demo ===\n")

	// Create a new AtomSpace
	ctx := context.Background()
	space := atomspace.NewDgraphSpace(ctx)

	// Example 1: Building a simple taxonomy
	fmt.Println("1. Creating a simple taxonomy:")
	cat, _ := space.AddNode(atomspace.ConceptNodeType, "cat")
	dog, _ := space.AddNode(atomspace.ConceptNodeType, "dog")
	mammal, _ := space.AddNode(atomspace.ConceptNodeType, "mammal")
	animal, _ := space.AddNode(atomspace.ConceptNodeType, "animal")

	// Create inheritance relationships
	catMammal, _ := space.AddLink(atomspace.InheritanceLinkType, []atomspace.Atom{cat, mammal})
	dogMammal, _ := space.AddLink(atomspace.InheritanceLinkType, []atomspace.Atom{dog, mammal})
	mammalAnimal, _ := space.AddLink(atomspace.InheritanceLinkType, []atomspace.Atom{mammal, animal})

	// Set truth values (PLN - Probabilistic Logic Networks)
	catMammal.SetTruthValue(atomspace.NewTruthValue(0.99, 0.95))
	dogMammal.SetTruthValue(atomspace.NewTruthValue(0.99, 0.95))
	mammalAnimal.SetTruthValue(atomspace.NewTruthValue(1.0, 0.99))

	fmt.Printf("  %s [TV: %.2f, %.2f]\n", catMammal.String(), 
		catMammal.GetTruthValue().Strength, catMammal.GetTruthValue().Confidence)
	fmt.Printf("  %s [TV: %.2f, %.2f]\n", dogMammal.String(),
		dogMammal.GetTruthValue().Strength, dogMammal.GetTruthValue().Confidence)
	fmt.Printf("  %s [TV: %.2f, %.2f]\n\n", mammalAnimal.String(),
		mammalAnimal.GetTruthValue().Strength, mammalAnimal.GetTruthValue().Confidence)

	// Example 2: Similarity relationships
	fmt.Println("2. Adding similarity relationships:")
	similarity, _ := space.AddLink(atomspace.SimilarityLinkType, []atomspace.Atom{cat, dog})
	similarity.SetTruthValue(atomspace.NewTruthValue(0.75, 0.8))
	fmt.Printf("  %s [TV: %.2f, %.2f]\n\n", similarity.String(),
		similarity.GetTruthValue().Strength, similarity.GetTruthValue().Confidence)

	// Example 3: Attention allocation (ECAN)
	fmt.Println("3. Setting attention values (ECAN):")
	cat.SetAttentionValue(atomspace.NewAttentionValue(150, 75, 20))
	dog.SetAttentionValue(atomspace.NewAttentionValue(120, 60, 15))
	mammal.SetAttentionValue(atomspace.NewAttentionValue(100, 80, 30))

	fmt.Printf("  cat:    STI=%d, LTI=%d, VLTI=%d\n",
		cat.GetAttentionValue().STI, cat.GetAttentionValue().LTI, cat.GetAttentionValue().VLTI)
	fmt.Printf("  dog:    STI=%d, LTI=%d, VLTI=%d\n",
		dog.GetAttentionValue().STI, dog.GetAttentionValue().LTI, dog.GetAttentionValue().VLTI)
	fmt.Printf("  mammal: STI=%d, LTI=%d, VLTI=%d\n\n",
		mammal.GetAttentionValue().STI, mammal.GetAttentionValue().LTI, mammal.GetAttentionValue().VLTI)

	// Example 4: Querying the knowledge base
	fmt.Println("4. Querying the knowledge base:")
	
	size, _ := space.Size()
	fmt.Printf("  Total atoms in space: %d\n", size)

	concepts, _ := space.GetAllAtoms(atomspace.ConceptNodeType)
	fmt.Printf("  Number of concepts: %d\n", len(concepts))

	inheritanceLinks, _ := space.GetAllAtoms(atomspace.InheritanceLinkType)
	fmt.Printf("  Number of inheritance links: %d\n", len(inheritanceLinks))

	// Example 5: Getting incoming links (backlinks)
	fmt.Println("\n5. Getting incoming links:")
	
	mammalIncoming, _ := space.GetIncoming(mammal)
	fmt.Printf("  'mammal' has %d incoming links:\n", len(mammalIncoming))
	for _, link := range mammalIncoming {
		fmt.Printf("    %s\n", link.String())
	}

	animalIncoming, _ := space.GetIncoming(animal)
	fmt.Printf("  'animal' has %d incoming links:\n", len(animalIncoming))
	for _, link := range animalIncoming {
		fmt.Printf("    %s\n", link.String())
	}

	// Example 6: Creating predicates and evaluations
	fmt.Println("\n6. Creating predicates and evaluations:")
	
	eats, _ := space.AddNode(atomspace.PredicateNodeType, "eats")
	meat, _ := space.AddNode(atomspace.ConceptNodeType, "meat")
	
	// (EvaluationLink (PredicateNode "eats") (ListLink (ConceptNode "cat") (ConceptNode "meat")))
	// Simplified version without ListLink for this demo
	catEatsMeat, _ := space.AddLink(atomspace.EvaluationLinkType, []atomspace.Atom{eats, cat, meat})
	catEatsMeat.SetTruthValue(atomspace.NewTruthValue(0.9, 0.85))
	
	fmt.Printf("  %s [TV: %.2f, %.2f]\n", catEatsMeat.String(),
		catEatsMeat.GetTruthValue().Strength, catEatsMeat.GetTruthValue().Confidence)

	// Example 7: JSON export
	fmt.Println("\n7. Exporting to JSON:")
	jsonData, err := space.MarshalJSON()
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	var prettyJSON map[string]interface{}
	json.Unmarshal(jsonData, &prettyJSON)
	prettyData, _ := json.MarshalIndent(prettyJSON, "  ", "  ")
	fmt.Printf("%s\n", prettyData)

	// Final statistics
	fmt.Println("\n=== Final Statistics ===")
	finalSize, _ := space.Size()
	fmt.Printf("Total atoms: %d\n", finalSize)
	
	allConcepts, _ := space.GetAllAtoms(atomspace.ConceptNodeType)
	allPredicates, _ := space.GetAllAtoms(atomspace.PredicateNodeType)
	allInheritance, _ := space.GetAllAtoms(atomspace.InheritanceLinkType)
	allSimilarity, _ := space.GetAllAtoms(atomspace.SimilarityLinkType)
	allEvaluation, _ := space.GetAllAtoms(atomspace.EvaluationLinkType)
	
	fmt.Printf("  ConceptNodes: %d\n", len(allConcepts))
	fmt.Printf("  PredicateNodes: %d\n", len(allPredicates))
	fmt.Printf("  InheritanceLinks: %d\n", len(allInheritance))
	fmt.Printf("  SimilarityLinks: %d\n", len(allSimilarity))
	fmt.Printf("  EvaluationLinks: %d\n", len(allEvaluation))
}
