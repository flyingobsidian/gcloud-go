# Golang Patterns and Anti-patterns

## Patterns to Follow

**Factory Pattern**
A creational design approach that enables object instantiation without requiring explicit class specification. Defined via a factory function that takes parameters and returns an interface, using a switch statement to return appropriate implementations based on input. Valuable when you need to create objects of different types based on a given input while maintaining code flexibility and separation of concerns.[^1]

**Singleton Pattern**
An architectural technique ensuring only one instance of a type exists throughout application execution. Implemented using a global variable for the instance and a sync.Once structure to provide thread-safe, lazy initialization. Helpful for providing a single source of truth or global state for an application.[^1]

**Decorator Pattern**
A structural approach permitting dynamic behavior augmentation without modifying underlying structures. Achieved via interface embedding and composition, enabling wrapping of functionality such as logging or caching. Adheres to the single responsibility principle and provides flexibility for future changes.[^1]

**Small Interfaces**
Keep interfaces minimal, ideally with one or two methods. Go's standard library follows this extensively—io.Reader, io.Writer, and error interfaces each contain a single method. Small interfaces promote composability, enable easier testing through focused mocks, and allow code to depend only on functionality it actually requires.[^2]

**Define Interfaces at the Point of Use**
Establish interfaces in the package where they are consumed, not where implementations reside. This inverts the dependency graph, ensuring business logic packages do not depend on infrastructure packages. Infrastructure adapters conform to contracts defined by core logic, enabling implementation swaps without modifying core business logic.[^2]

**Accept Interfaces, Return Structs**
Function signatures should accept interface types as parameters while returning concrete struct types. Accepting interfaces maintains flexibility and loose coupling, while returning structs provides API clarity and allows callers to access the full method set. Factory functions may legitimately return interfaces when supporting multiple implementation types.[^2]

**Composition Over Interface Inheritance**
Combine smaller interfaces to create more complex contracts rather than creating large, monolithic interfaces. By embedding interfaces, you build flexible hierarchies where consumers can depend on exactly what they need. Functions accepting an io.Reader work with anything satisfying that contract, maximizing reusability.[^2]

**Interface Versioning for Backward Compatibility**
When extending interfaces, create new interface types (e.g., StorageV2) that embed the original rather than modifying existing interfaces. At runtime, code checks whether implementations support enhanced capabilities via type assertions. Existing implementations continue working without modification while new ones can opt into extended functionality.[^2]

**Minimal Mock Interfaces for Testing**
Leverage Go's implicit interface satisfaction to create simple test doubles with only the methods necessary for testing. No mocking frameworks required—just define minimal structs implementing the needed methods. Keeps tests focused, simple, and readable.[^2]

**Generator Pattern**
A concurrency design approach that creates a function producing a series of values through a channel, yielding values one at a time rather than returning a complete slice. The generator function returns a receive-only channel; inside, a goroutine sends values until exhausted. Enables lazy evaluation—values are produced only when requested—preventing memory exhaustion with large datasets.[^5]

**Worker Pool Pattern**
A concurrency model employing a fixed number of goroutines (workers) that process tasks from a shared jobs channel. Each worker independently receives assignments, executes work, and sends results through a results channel. Limits concurrent goroutine creation and prevents resource exhaustion, improving performance compared to spawning a new goroutine per task. A WaitGroup synchronizes completion.[^5]

**Pipeline Pattern**
A decomposition technique breaking complex processing into sequential stages, each executing independently via goroutines, where output from one stage becomes input for the next through channel connections. Multiple data items progress simultaneously through different stages, enabling concurrent execution and substantially improving throughput.[^5]

**Fan-In Pattern**
A consolidation technique that merges multiple input channels into a single output channel. Goroutines listen to different input channels and forward received values to a shared output channel. Improved variants use a WaitGroup to properly close the output channel only after all inputs are exhausted, avoiding goroutine leaks.[^5]

**Semaphore Pattern**
A resource-access control mechanism limiting simultaneous goroutine access to a protected resource, allowing a configurable number of concurrent accessors (unlike sync.Mutex which allows one). Implemented with a buffered channel whose capacity equals the maximum concurrent accessors. Can be wrapped in a struct with Acquire() and Release() methods for a cleaner API.[^5]

**Timeout Pattern**
A mechanism ensuring code sections terminate within specified durations, preventing indefinite hangs. Integrates context cancellation with time.NewTimer or time.NewTicker; a select statement listens for context cancellation or timer expiration, enabling graceful termination of long-running or blocking operations.[^5]

**Explicit Error Return Values**
Go's foundational error pattern treats errors as ordinary values returned from functions alongside normal results. Functions return an error as their final return value, keeping error handling visible throughout the call chain and ensuring developers consciously address failure modes at each level.[^6][^7]

**Error Wrapping with Context**
Use fmt.Errorf with the %w verb (Go 1.13+) to preserve the original error while adding contextual information about what operation failed. As errors propagate upward through function calls, wrapping accumulates details creating an auditable trail of failure. Combined with errors.Unwrap, enables tracing errors to their source.[^6][^7]

**Sentinel Error Variables**
Define package-level error variables for predictable failure conditions. These immutable sentinel errors enable callers to use errors.Is() for reliable matching, avoiding fragile string-based comparisons and supporting programmatic handling.[^7]

**Custom Error Types**
Define types that implement the error interface to provide richer failure information. Custom types can encode additional state, represent distinct error categories, and give callers programmatic ways to distinguish between failure modes. Use errors.As() for type-safe extraction from error chains.[^6][^7]

**Structured Error Categories**
Extend custom error types with enumerated cause fields rather than creating a unique type per failure condition. A single error type with a Cause field holding enumerated values reduces type proliferation while maintaining semantic clarity and scalability.[^6]

**Error Chain Inspection with errors.Is and errors.As**
Use errors.Is to check whether an error or any error in its wrapped chain matches a specific sentinel value. Use errors.As to extract custom error types from error chains for access to structured information. Both functions support wrapped error chains, replacing fragile string matching and direct type assertions.[^6][^7]

**Combining Multiple Errors with errors.Join**
errors.Join (Go 1.20+) creates a single error value representing multiple simultaneous failures. Enables expressing scenarios where independent operations all fail, allowing callers to check for specific errors within the combined set using errors.Is.[^6]

**Panic Recovery with Deferred Functions**
Use defer with recover() to gracefully handle exceptional runtime conditions such as nil pointer dereferences or out-of-bounds access. Recovering panics within deferred functions keeps applications operational rather than crashing, while recording the error for observability.[^6]

**Defer for Resource Cleanup**
Use defer statements to guarantee cleanup code runs regardless of error paths. Ensures resources like file handles, database connections, and locks are released even when errors occur, preventing resource leaks.[^7]

**Behavior-Based Error Handling**
Design error types around what callers need to do rather than error categorization alone. For example, a TemporaryError interface allows callers to determine whether retry logic should apply, decoupling the caller's recovery strategy from the error's internal classification.[^7]

**Early Return Pattern**
Check error conditions at function entry and return immediately upon failure. Keeps successful execution paths left-aligned and improves readability by avoiding deeply nested conditionals. Also known as the guard clause pattern.[^7]

**Error Handling at Boundaries**
Handle errors once at appropriate abstraction levels rather than repeatedly throughout call chains. Either handle an error (log, recover) or return it for upstream handling—not both. This avoids duplicated logging and establishes clear ownership of error responses.[^7]

## Anti-patterns to Avoid

**Ignoring Errors**
Failing to handle or acknowledge error conditions, whether by using the blank identifier (_) to discard error return values or by not inspecting returned errors. Silently swallows failures, hides problems, makes debugging difficult, and can lead to subtle bugs, security vulnerabilities, or crashes.[^1][^6][^7]

**Nil Return Instead of Error**
Returning nil from a function when an error occurred, rather than returning an explicit error value with information about what went wrong. Prevents the caller from making informed decisions about handling the failure.[^1]

**Not Using Synchronization Primitives**
Neglecting concurrency tools such as sync.Mutex, sync.RWMutex, and sync.WaitGroup when developing concurrent applications, leading to race conditions, data corruption, or deadlocks. Concurrent access to shared variables without synchronization causes data races.[^1]

**Reinventing the Wheel / Not Using Helper Functions**
Implementing custom solutions or manually performing operations that established libraries, the standard library, or dedicated helper functions already handle. For example, using wg.Add(-1) instead of wg.Done(), or implementing custom serialization when the standard library suffices. Wastes development time and introduces unnecessary bugs.[^1][^3][^4]

**Returning Unexported Types from Exported Functions**
An exported (capitalized) function that returns a value of an unexported (lowercase) type creates friction for API consumers. Callers from other packages cannot directly use the return type without redefining it themselves, defeating the purpose of exporting the function.[^3][^4]

**Overuse of the Blank Identifier**
Excessive or unnecessary use of the underscore (_) to ignore values where it serves no purpose. Examples: "for _ = range" instead of "for range", or ignoring a second return value unnecessarily. Creates visual clutter and obscures code intent.[^3][^4]

**Looping to Concatenate Slices**
Merging slices by iterating and appending elements individually rather than using Go's variadic append. Verbose, inefficient, and performs unnecessary allocations. Use append(sliceOne, sliceTwo...) instead.[^3][^4]

**Redundant Arguments in Make Calls**
Passing explicit default arguments to the make function when unnecessary. For instance, make(chan int, 0) specifies zero buffer capacity when this is already the default for unbuffered channels, adding verbosity without value.[^3][^4]

**Unnecessary Return Statements in Void Functions**
Including a return statement at the end of functions that do not return any value. Adds no functional value and clutters the code. Named return values with explicit return statements are a separate, legitimate construct.[^3][^4]

**Redundant Break in Switch Cases**
Adding break statements after switch cases. Unlike C-style languages, Go switch cases do not fall through to the next case automatically, making explicit breaks redundant and confusing to readers.[^3][^4]

**Unnecessary Nil Checks on Slices**
Checking whether a slice is nil before examining its length, such as "if x != nil && len(x) != 0". Since nil slices have zero length in Go, the nil check is redundant; "if len(x) != 0" alone suffices.[^3][^4]

**Overly Complex Function Literals**
Creating function literals that merely wrap another function call without adding behavior. For instance, "fn := func(x, y int) int { return add(x, y) }" when a direct assignment "fn := add" would suffice. Adds indirection without value.[^3][^4]

**Single-Case Select Statements**
Employing the select statement for operations involving only one channel. The select construct is designed for managing multiple concurrent channel operations. For single channels, use direct send/receive; add a default case for non-blocking behavior.[^3][^4]

**context.Context Parameter Ordering**
Placing context.Context anywhere other than as the first parameter in function signatures. Go convention dictates context should be the initial parameter to ensure consistency. Go's variadic parameters must appear last, making first position the logical standard for context.[^3][^4]

**Interface Pollution**
Creating an interface for every struct "just in case" future implementations emerge, even when only one implementation exists and no concrete abstraction need is evident. Interfaces should emerge from actual abstraction needs, not hypothetical future flexibility.[^2]

**The Empty Interface Trap**
Overusing interface{} (or any in Go 1.18+) as a generic catch-all type. This sacrifices compile-time type safety for runtime flexibility, requiring extensive type assertions and error handling throughout code. Limit to scenarios where flexibility is genuinely necessary, such as JSON unmarshaling.[^2]

**Interface Abstraction for Its Own Sake**
Wrapping concrete types that do not benefit from abstraction—configuration structs, DTOs, data containers—in unnecessary interfaces. These types are fundamentally concrete and rarely need polymorphic behavior. A simple Config struct is better than a ConfigInterface with getter methods.[^2]

**Unnecessary Large Interfaces**
Combining unrelated functionality into single interfaces that try to serve multiple purposes. Forces implementations to provide methods they might not need and consumers to depend on functionality they do not use. Prefer focused, single-purpose interfaces.[^2]

**String-Based Error Comparison**
Comparing error messages using string operations instead of type-safe mechanisms. Message wording can change, translations complicate matching, and there is no compile-time safety. Use errors.Is, errors.As, or sentinel variable comparisons instead.[^6][^7]

**Excessive Custom Error Types**
Creating a unique error type for every distinct failure condition leads to type explosion and reduced readability. Forces callers to master an ever-growing type hierarchy when a smaller set of categorized types with enumerated causes would suffice.[^6]

**Panic for Control Flow**
Using panic-and-recover mechanisms to implement normal application logic rather than to handle unexpected runtime conditions. Panics should signal genuinely exceptional situations, not expected business logic branches.[^6]

**Losing Error Context**
Returning errors without additional information about where or why failure occurred, or repeatedly wrapping errors without adding meaningful new context. Wrap errors at decision points, adding useful information without indiscriminate nesting at every level.[^6][^7]

**Dual Error Handling**
Both logging an error and returning it to the caller. Creates redundant logging and violates the single-responsibility principle for error handling. Either handle the error at the current level (log, recover) or return it for upstream handling, not both.[^7]

**Goroutine Leaks**
Failing to provide a mechanism to stop goroutines after their work is complete or after input channels close, resulting in perpetual background goroutine execution and resource exhaustion. Always ensure goroutines can terminate—use context cancellation, done channels, or WaitGroups to manage goroutine lifecycles.[^5]

[^1]: See [Go Patterns and Anti-Patterns](https://appmaster.io/blog/go-patterns-anti-patterns).
[^2]: See [Go Interfaces: Design Patterns and Anti-Patterns](https://reintech.ai/blog/go-interfaces-design-patterns-and-anti-patterns).
[^3]: See [Go by Common: Avoiding Anti-Patterns in Go Programming](https://www.devzery.com/post/go-by-common).
[^4]: See [Common anti-patterns in Go](https://deepsource.com/blog/common-antipatterns-in-go).
[^5]: See [Mastering 6 Golang Concurrency Patterns](https://reliasoftware.com/blog/golang-concurrency-patterns).
[^6]: See [A practical guide to error handling in Go](https://www.datadoghq.com/blog/go-error-handling/).
[^7]: See [Error Handling Best Practices in Go](https://www.bytesizego.com/blog/error-handling-golang).
