// Package protocol defines the bounded, deterministic Sovereign V4 edge wire
// contracts. It is an unwired V4-0 library: importing it does not enable a
// transport, wallet, application runtime, chain transaction, or deployment.
//
// MembershipFact is intentionally a caller-provided fact. This package binds
// that fact to certificates and envelopes, but it does not verify a TRChain
// light-client proof. Callers must perform that verification before describing
// a fact as verified.
//
// PublisherFact follows the same boundary for the domain-app publisher
// registry. ReplayStore must receive only AuthValid envelopes and intentionally
// returns a hard capacity error until verified history is checkpointed and a
// new bounded store is built; it never silently forgets replay evidence.
package protocol
