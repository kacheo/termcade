// Package testharness documents the convention for game-level scenario
// tests in tmvgs.
//
// Why most helpers stay local
//
// The devlog reference project shares helpers via tests/testharness because
// its surface is a stable CLI: every helper can drive the public subprocess
// interface. tmvgs is a Bubble Tea TU where each game has its own internal
// state machine — phase enums, AI delays, board cells — that scenario tests
// legitimately need to peek at and mutate.
//
// Lifting those helpers out of their game package would require either
// exporting dedicated test hooks on every game type (polluting the public
// API) or duplicating state machines. Both are worse than the current
// arrangement, where each game owns a small set of local helpers:
//
//	games/poker/scenario_test.go     driveToHumanTurn, driveToCompletion,
//	                                 newTestPoker
//	games/blackjack/scenario_test.go advancePastDealerTurn, makeDeck
//	games/sudoku/scenario_test.go    findEmptyCell
//	games/sudoku/render_test.go      generateSeeded
//
// The total footprint is ~50 lines of clearly local helpers, each used
// only inside its own package.
//
// What this package is for
//
// This package is reserved for cross-cutting test utilities that do NOT
// depend on unexported game state — for example, future helpers that
// scrub the environment, build fixtures shared across packages, or assert
// on common output shapes. Add new helpers here as concrete needs arise
// rather than pre-emptively abstracting game-local logic.
//
// Determinism
//
// Several games (tetris, blackjack, poker) call rand.Shuffle / rand.Intn
// against the global math/rand source during New*() and during gameplay.
// In Go 1.20+ the global source is no longer reliably re-seedable via
// rand.Seed, so the only portable way to get a deterministic initial
// state is to set the per-game rng field directly from within the game
// package — exactly what each game's render_test.go and scenario_test.go
// already do. That pattern is the canonical answer; do not try to wrap
// it in a global helper.
package testharness
