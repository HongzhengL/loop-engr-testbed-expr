---
id: 001-boundary-value-test-cases-for-comparison-operato
ticket: T10
created: "2026-08-01"
scope:
    pipeline:
        - direct_l2
    role:
        - ut_writer
    ticket_type:
        - refactor
---

# Boundary-value test cases for comparison operators are critical for mutation coverage

When generating test cases for functions that implement comparison operators (<=, >=, <, >, ==), include test cases where operands are adjacent in value—differing by exactly 1. These boundary cases exercise the precise switch point where the operator's return value changes from true to false or vice versa.

For example, when testing a function `LessOrEqual(a, b)`:
- Test cases like `(4, 5)`, `(5, 5)`, `(6, 5)` reveal the exact semantics
- Test cases like `(1, 5)`, `(1, 10)` (difference ≥ 2) do not exercise the critical boundary

Without these adjacent-value cases, mutations to the boundary logic (e.g., `<=` changed to `<`, or boundary constants off-by-one) can survive testing undetected. Mutation testing will mark them as "code not covered" because the boundary conditions were never exercised, even if statement coverage appears adequate.

When expanding test coverage for a comparison operator:
1. For each interesting operand value, include cases at n-1, n, and n+1
2. Avoid focusing only on wide-gap differences (e.g., 1 vs 9)
3. If expanding the operand range, maintain boundary coverage across the new range

This pattern applies both to unit tests generated from contract specs and to tests written via iterative improvement after mutation analysis.
