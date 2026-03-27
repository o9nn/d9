/*
 * SPDX-FileCopyrightText: © Hypermode Inc. <hello@hypermode.com>
 * SPDX-License-Identifier: Apache-2.0
 */

package atomspace

// ECAN (Economic Attention Network) helpers.
//
// These functions implement simplified versions of the ECAN attention-
// spreading dynamics described in:
//   Goertzel et al., "OpenCog Prime" (2012)

// ECANConfig holds tunable parameters for attention spreading.
type ECANConfig struct {
	// STIDecayRate is the fraction of STI lost per decay step (0–1).
	STIDecayRate float64
	// LTIDecayRate is the fraction of LTI lost per decay step (0–1).
	LTIDecayRate float64
	// SpreadFraction is the fraction of an atom's STI spread to its
	// neighbours on each spreading step.
	SpreadFraction float64
	// MaxSTI is the upper bound for STI values.
	MaxSTI int64
	// MinSTI is the lower bound for STI values.
	MinSTI int64
}

// DefaultECANConfig returns a sensible set of default parameters.
func DefaultECANConfig() ECANConfig {
	return ECANConfig{
		STIDecayRate:   0.1,
		LTIDecayRate:   0.01,
		SpreadFraction: 0.2,
		MaxSTI:         1000,
		MinSTI:         -1000,
	}
}

// clampSTI restricts an STI value to [cfg.MinSTI, cfg.MaxSTI].
func clampSTI(v int64, cfg ECANConfig) int64 {
	if v > cfg.MaxSTI {
		return cfg.MaxSTI
	}
	if v < cfg.MinSTI {
		return cfg.MinSTI
	}
	return v
}

// ECANDecay applies one step of attention decay to every atom in the space.
// STI is reduced by STIDecayRate and LTI by LTIDecayRate.
func (ds *DgraphSpace) ECANDecay(cfg ECANConfig) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	for _, atom := range ds.atoms {
		av := atom.GetAttentionValue()
		newSTI := int64(float64(av.STI) * (1 - cfg.STIDecayRate))
		newLTI := int64(float64(av.LTI) * (1 - cfg.LTIDecayRate))
		atom.SetAttentionValue(NewAttentionValue(
			clampSTI(newSTI, cfg),
			newLTI,
			av.VLTI,
		))
	}
}

// ECANSpread performs one step of STI spreading from each atom to its
// immediate neighbours (atoms in the outgoing set of links that point to it).
// Each atom donates SpreadFraction of its STI, shared equally among its
// neighbours.
func (ds *DgraphSpace) ECANSpread(cfg ECANConfig) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// Collect donations first to avoid order-dependent updates.
	donations := make(map[string]int64, len(ds.atoms))

	for id, atom := range ds.atoms {
		link, ok := atom.(*Link)
		if !ok {
			continue
		}
		neighbours := link.GetOutgoing()
		if len(neighbours) == 0 {
			continue
		}
		linkAV := link.GetAttentionValue()
		totalDonation := int64(float64(linkAV.STI) * cfg.SpreadFraction)
		if totalDonation <= 0 {
			continue
		}
		perNeighbour := totalDonation / int64(len(neighbours))
		if perNeighbour == 0 {
			continue
		}
		// Deduct from the link
		donations[id] -= totalDonation
		for _, nb := range neighbours {
			donations[nb.GetID()] += perNeighbour
		}
	}

	// Apply accumulated donations
	for id, delta := range donations {
		if atom, ok := ds.atoms[id]; ok {
			av := atom.GetAttentionValue()
			atom.SetAttentionValue(NewAttentionValue(
				clampSTI(av.STI+delta, cfg),
				av.LTI,
				av.VLTI,
			))
		}
	}
}

// ECANTopN returns up to n atoms with the highest STI values.
// The returned slice is sorted descending by STI.
// Uses a min-heap of size n for O(|atoms| * log(n)) performance.
func (ds *DgraphSpace) ECANTopN(n int) []Atom {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if n <= 0 {
		return nil
	}

	h := &stiMinHeap{}
	for _, a := range ds.atoms {
		if h.Len() < n {
			h.push(a)
			if h.Len() == n {
				// Build the heap only once it is full.
				h.init()
			}
		} else if a.GetAttentionValue().STI > (*h)[0].GetAttentionValue().STI {
			(*h)[0] = a
			h.down(0)
		}
	}

	// Drain the heap into a slice sorted descending by STI.
	result := make([]Atom, h.Len())
	for i := len(result) - 1; i >= 0; i-- {
		result[i] = h.pop()
	}
	return result
}

// stiMinHeap is a simple min-heap keyed on STI, used internally by ECANTopN.
type stiMinHeap []Atom

func (h *stiMinHeap) Len() int { return len(*h) }

func (h *stiMinHeap) push(a Atom) { *h = append(*h, a) }

// pop removes and returns the minimum-STI element.
func (h *stiMinHeap) pop() Atom {
	n := len(*h) - 1
	(*h)[0], (*h)[n] = (*h)[n], (*h)[0]
	min := (*h)[n]
	*h = (*h)[:n]
	h.down(0)
	return min
}

// init builds the heap in-place (Floyd's algorithm).
func (h *stiMinHeap) init() {
	for i := h.Len()/2 - 1; i >= 0; i-- {
		h.down(i)
	}
}

// down sifts element at index i downwards to restore the min-heap property.
func (h *stiMinHeap) down(i int) {
	n := h.Len()
	for {
		left := 2*i + 1
		if left >= n {
			break
		}
		smallest := left
		if right := left + 1; right < n && (*h)[right].GetAttentionValue().STI < (*h)[left].GetAttentionValue().STI {
			smallest = right
		}
		if (*h)[i].GetAttentionValue().STI <= (*h)[smallest].GetAttentionValue().STI {
			break
		}
		(*h)[i], (*h)[smallest] = (*h)[smallest], (*h)[i]
		i = smallest
	}
}

// ECANStimulate increases the STI of the given atom by amount, clamped by the
// configured maximum.
func ECANStimulate(atom Atom, amount int64, cfg ECANConfig) {
	av := atom.GetAttentionValue()
	atom.SetAttentionValue(NewAttentionValue(
		clampSTI(av.STI+amount, cfg),
		av.LTI,
		av.VLTI,
	))
}
