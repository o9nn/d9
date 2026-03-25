/*
 * SPDX-FileCopyrightText: © Hypermode Inc. <hello@hypermode.com>
 * SPDX-License-Identifier: Apache-2.0
 */

package atomspace

import "math"

// PLN (Probabilistic Logic Networks) inference helpers.
//
// Each function operates on TruthValue pairs and returns a new TruthValue
// following the standard PLN formulas described in:
//   Goertzel et al., "Probabilistic Logic Networks" (2008)

// PLNRevision combines two independent truth values about the same statement
// using the PLN revision rule.  The resulting confidence is the sum of the
// input confidences (clamped to 1), and the resulting strength is the
// confidence-weighted average of the input strengths.
func PLNRevision(tv1, tv2 *TruthValue) *TruthValue {
	c1 := tv1.Confidence
	c2 := tv2.Confidence
	totalConf := clamp(c1+c2, 0.0, 1.0)
	if totalConf == 0.0 {
		return NewTruthValue(0.5, 0.0)
	}
	strength := (tv1.Strength*c1 + tv2.Strength*c2) / (c1 + c2)
	return NewTruthValue(strength, totalConf)
}

// PLNDeduction applies the PLN deduction rule:
//
//	A→B with tv(sAB, cAB), B→C with tv(sBC, cBC)
//	yields A→C with the approximated truth value.
//
// The simplified formula used here is:
//
//	strength = sAB * sBC + (1-sAB) * (sB/(1-sB)) * (1-sBC)  [truncated to [0,1]]
//	confidence = min(cAB, cBC)
//
// where sB is the strength of the premise node B (pass DefaultTruthValue when
// the base rate is unknown).
func PLNDeduction(tvAB, tvBC, tvB *TruthValue) *TruthValue {
	sAB := tvAB.Strength
	sBC := tvBC.Strength
	sB := tvB.Strength

	var strength float64
	if sB >= 1.0 {
		strength = sAB * sBC
	} else {
		strength = sAB*sBC + (1-sAB)*(sB/(1-sB))*(1-sBC)
	}
	strength = clamp(strength, 0.0, 1.0)
	confidence := math.Min(tvAB.Confidence, tvBC.Confidence)
	return NewTruthValue(strength, confidence)
}

// PLNAbduction applies the PLN abduction rule:
//
//	A→B, A→C  yields  B→C
//
// Simplified approximation: s(B→C) ≈ s(A→C) / s(A→B) when s(A→B) > 0.
// tvA and tvB are reserved for a full second-order abduction formula and are
// not used in this simplified implementation.
func PLNAbduction(tvAB, tvAC, tvA, tvB *TruthValue) *TruthValue {
	sAB := tvAB.Strength
	if sAB == 0.0 {
		return NewTruthValue(0.5, 0.0)
	}
	strength := clamp(tvAC.Strength/sAB, 0.0, 1.0)
	confidence := math.Min(tvAB.Confidence, tvAC.Confidence)
	_ = tvA
	_ = tvB
	return NewTruthValue(strength, confidence)
}

// PLNInversion (Bayes' rule) derives P(A|B) from P(B|A) and base rates tvA, tvB.
//
//	s(A→B) * s(A) = s(B→A) * s(B)
//	⟹  s(B→A) = s(A→B) * s(A) / s(B)
func PLNInversion(tvAB, tvA, tvB *TruthValue) *TruthValue {
	sB := tvB.Strength
	if sB == 0.0 {
		return NewTruthValue(0.5, 0.0)
	}
	strength := clamp(tvAB.Strength*tvA.Strength/sB, 0.0, 1.0)
	confidence := math.Min(tvAB.Confidence, math.Min(tvA.Confidence, tvB.Confidence))
	return NewTruthValue(strength, confidence)
}

// PLNAnd computes the truth value of (A AND B) under the assumption that A and
// B are independent.
//
//	s(A∧B) = s(A) * s(B)
//	c(A∧B) = min(c(A), c(B))
func PLNAnd(tv1, tv2 *TruthValue) *TruthValue {
	return NewTruthValue(tv1.Strength*tv2.Strength,
		math.Min(tv1.Confidence, tv2.Confidence))
}

// PLNOr computes the truth value of (A OR B) under the assumption that A and
// B are independent.
//
//	s(A∨B) = 1 - (1-s(A))*(1-s(B))
//	c(A∨B) = min(c(A), c(B))
func PLNOr(tv1, tv2 *TruthValue) *TruthValue {
	strength := 1 - (1-tv1.Strength)*(1-tv2.Strength)
	return NewTruthValue(strength,
		math.Min(tv1.Confidence, tv2.Confidence))
}

// PLNNot computes the truth value of (NOT A).
//
//	s(¬A) = 1 - s(A)
//	c(¬A) = c(A)
func PLNNot(tv *TruthValue) *TruthValue {
	return NewTruthValue(1-tv.Strength, tv.Confidence)
}
