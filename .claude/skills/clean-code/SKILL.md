---
name: clean-code
description: Clean code guidelines for Go (Gin/REST API project) — naming, function size, error handling, comments, and readability — focused on pragmatic application and Go's own idioms, avoiding generic rules from other languages that don't fit here. Use whenever writing, reviewing, or refactoring Go code, naming variables/functions/packages, handling errors, or when the user asks for "clean code", "readability", "best practices", or a code review.
---

# Clean Code in Go

A clean code guide adapted to Go/Gin idioms. Clean Code (Robert Martin) was written with Java in mind — many of its rules don't translate 1:1 to Go, and applying them without adapting produces code that's *less* idiomatic, not cleaner. This skill filters what actually holds up in Go and drops what doesn't.

## Golden rule: readability > rule

Every "cleanup" suggestion must answer: **does this make the code easier to understand for whoever reads it later, or does it just look "nicer" on paper?**

Signs a suggestion is overengineering disguised as clean code:
- Splitting a simple, linear 15-line function into 4 functions of 3 lines each just to "follow the small-function rule" — if this forces the reader to jump between 4 places to follow a simple flow, it made things worse.
- Adding layers of indirection (wrapper over wrapper) just to "hide complexity" that wasn't actually complex.
- Renaming something that was already clear just to follow a naming convention to the letter.
- Adding comments that explain the obvious (`// increment i` above `i++`).
- Extracting a constant or config struct for a value used exactly once, with no real clarity gain.

If a change doesn't reduce real reading/maintenance effort, don't make it. Favor pragmatism: correct, readable, direct code beats "canonically clean" but fragmented code.

## Naming

- **Prefer descriptive names over short ones, even in tight scopes.** Go idiom leans toward short names (`i`, `err`, `r`/`w`, `ctx`) for very short-lived loop/local variables, but default to a clear, descriptive name whenever there's any ambiguity about what a variable holds or does — don't shorten just to "look more Go-like." `userID` over `u`, `paymentGateway` over `pg`, `requestBody` over `body` when there's more than one body-like value in scope. The only names that stay short are truly conventional, unambiguous ones: `err` for an error, `ctx` for `context.Context`, `i`/`j` for a plain integer loop index, `w`/`r` for `http.ResponseWriter`/`*http.Request` when there's exactly one of each in scope.
- For anything with a wider scope (struct fields, exported functions, service/repository variables), always use descriptive names: `userRepository`, `ErrInvalidPayload`, `paymentProcessor`.
- Package names: short, lowercase, no underscores, and don't repeat what they export (avoid `user.UserService` — prefer `user.Service`). Short package name is fine; short *variable* names inside it are not the default.
- Getters don't use a `Get` prefix in idiomatic Go: `user.Name()`, not `user.GetName()`.
- Exported errors start with `Err`: `var ErrNotFound = errors.New("not found")`.
- Single-method interfaces usually end in `-er`: `Reader`, `Validator`.

## Functions

- Prefer small functions with a clear purpose, but **don't fragment artificially**: a function that runs a linear, sequential flow (validate → fetch → transform → return) can stay as one function if each step is simple and the function name makes clear what it does.
- Extract a sub-function when: the block has a clearly separate responsibility, it's reused in more than one place, or the sub-function's name reduces the reader's cognitive load (the code "tells a better story" with it extracted).
- Short signatures: if a Go function takes 5+ parameters, consider grouping them into an options/config struct — this is idiomatic and improves readability at call sites.
- Returning the error as the last value is the Go standard (`func Foo() (Result, error)`) — always follow this convention, never invert the order.

## Error handling

- Handle the error explicitly where it happens; don't discard it with `_` unless it's genuinely irrelevant (and if so, consider a comment justifying why).
- Errors should carry context about where they came from: use `fmt.Errorf("fetching user %s: %w", id, err)` (with `%w` to allow `errors.Is`/`errors.As`), rather than returning the raw error with no context as it bubbles up through layers.
- Don't use `panic` for expected errors (input validation, resource not found, etc). `panic` is for genuinely unrecoverable states (programming bugs, broken invariants). In Gin handlers, an expected error means an HTTP response with the right status code, not panic/recover as normal control flow.
- Avoid repeated `if err != nil { return err }` with zero added context when something more direct would do — but this pattern itself is standard Go idiom, so don't try to "hide" it behind generic helpers just to cut down line count.

## Comments

- A good comment explains **why**, not **what** — if the comment just repeats what the code already says clearly, remove the comment (or improve the variable/function name so it isn't needed).
- Comments on exported functions (Go doc convention): start with the function's name — `// FindUserByID looks up a user by ID in the database.` — this is a real Go convention (`go doc`/`godoc`), worth following.
- TODO/FIXME are fine when they describe a real, known limitation, not as an excuse to leave sloppy code in place.

## Formatting and structure

- `gofmt`/`goimports` handle formatting — don't debate indentation style, brace placement, etc.: it's automatic and non-negotiable in Go.
- Package organization: group by domain/feature (e.g. `internal/order/`, `internal/user/`), not by technical type (loose `controllers/`, `services/`, `models/` at the root) — this is more idiomatic in Go than the horizontal-layer pattern common in Java/C#.
- Avoid generic "utils" or "helpers" packages that become a dumping ground — if a function is only used in one place, consider keeping it near that use site instead of "organizing it preemptively."

## How to apply this in the workflow

1. Write for correctness and direct clarity first — don't optimize for "looking clean" before it works.
2. When reviewing, articulate in one sentence the real readability or maintenance gain for every suggested change. If you can't articulate a concrete gain, don't suggest the change.
3. If the user's code is already over-fragmented (functions too small, forcing jumps between places, unnecessary abstractions), it's valid to suggest **consolidating**, not just "cleaning" further.
4. Run/suggest `gofmt` and `go vet` for anything mechanical — don't spend manual review time on that.