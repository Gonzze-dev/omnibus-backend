# Golang Backend Standards & Skills

This repository defines the architectural standards, coding patterns, and best practices for building scalable, maintainable, and idiomatic Go backend services. It serves as the source of truth for both human developers and AI agents working on the project.

## Core Philosophy

*   **Standard Library First**: Minimize external dependencies. Use the robust Go standard library whenever possible.
*   **Layered Architecture**: Strict separation of concerns between Transports (HTTP), Business Logic (Service), and Data Access (Repository).
*   **Interface Design**: Return structs, accept interfaces. This provides maximum flexibility for consumers while keeping implementations concrete.
*   **Type Safety**: Leverage Go's type system to enforce contracts, especially at the API boundaries using the `Handler` wrapper.
*   **Simplicity (KISS)**: Prioritize readable, simple code over complex abstractions.

## Documentation Map

The `references/` directory contains detailed guides on specific topics.

### 🏗 Architecture & Structure
*   [Project Organization](references/PROJECT_ORGANIZATION.md): Standard directory layout (`cmd`, `internal`, `pkg`).
*   [Core Architectural Patterns](references/CORE_ARCHITECTURAL_PATTERNS.md): High-level overview of the system design.
*   [Layered Responsibility](references/LAYERED_RESPONSIBILITY.md): Rules for what goes into Handlers vs. Services vs. Repositories.
*   [Dependency Injection](references/DEPENDENCY_INJECTION.md): How to manage dependencies using constructors.

### 🔌 HTTP & API
*   [The Handler Wrapper](references/THE_HANDLER_WRAPPER.md): The generic `Handler[REQ, RESP]` pattern for standardized API endpoints.
*   [Middleware](references/MIDDLEWARE.md): Handling cross-cutting concerns like auth and logging.
*   [Request Validation](references/VALIDATOR_INTERFACE.md): Patterns for validating request parameters using composable interfaces.
*   [Context Usage](references/CONTEXT_USAGE.md): Proper propagation of `context.Context`.

### 💾 Data & Logic
*   [Database Access](references/DATABASE_ACCESS.md): Patterns for using raw SQL with `pgx` or `sqlx`.
*   [Error Handling](references/ERROR_HANDLING.md): Typed errors and return value conventions.
*   [Extensibility](references/EXTENSIBILITY.md): Designing for future growth without breaking changes.

### 🎨 Coding Standards
*   [Design Principles](references/DESIGN_PRINCIPLES.md): SOLID, KISS, and other guiding principles.
*   [Standard Library First](references/STANDARD_LIBRARY_FIRST.md): Guidelines on avoiding unnecessary libraries.
*   [Function Parameter Design](references/FUNCTION_PARAMETER_DESIGN.md): Best practices for function signatures and options patterns.
*   [Parameter Mutation](references/AVOID_MUTATING_PARAMETERS.md): Why and how to avoid modifying input parameters.
*   [Testing](references/TESTING.md): Strategies for unit and integration testing.

### ⚠️ Error & Panic Safety
*   [Constructor Error Handling](references/CONSTRUCTOR_ERROR_HANDLING.md): How constructors should handle errors instead of panicking.
*   [Must-Prefix Functions](references/MUST_FUNCTIONS.md): Safe usage patterns for Must-functions and panic helpers.

### 🔄 Concurrency
*   [Goroutine Pools](references/GOROUTINE_POOLS.md): Managing goroutines to avoid memory leaks and unbounded concurrency.
*   [Mutex Locking](references/MUTEX_LOCKING.md): Best practices for minimizing lock contention using fine-grained locking.

### 🚀 Performance
*   [Performance Optimization](references/PERFORMANCE_OPTIMIZATION.md): Techniques for object pooling and stack allocation.

## For AI Agents

*   **[SKILL.md](SKILL.md)**: This is the primary instruction file for AI agents. It summarizes the critical rules from the references above to ensure the AI generates code that adheres to these standards.