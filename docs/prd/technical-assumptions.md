# Technical Assumptions

## Repository Structure: Monorepo

Single repository containing CLI tool, analysis engine, and output generation components. This approach simplifies MVP development, testing, and deployment while enabling shared utilities and consistent versioning across components. Future scaling may require polyrepo transition for microservices architecture.

## Service Architecture

**Monolith with Modular Design** - MVP implementation uses single deployable application with clearly separated modules: Git Interface, Security Sandbox, Analysis Engine, and Output Generator. This approach minimizes operational complexity while establishing clean architectural boundaries for future microservices extraction. Container-based isolation provides security without distributed system complexity.

## Testing Requirements

**Full Testing Pyramid** - Comprehensive testing strategy including:
- Unit tests for analysis algorithms and output generation (>90% coverage)
- Integration tests for Git operations and sandbox isolation
- End-to-end tests validating complete repository analysis workflows
- Security testing for sandbox containment and vulnerability detection
- Performance testing to validate 1-hour analysis targets
- Manual testing convenience methods for rapid developer feedback during iteration

## Additional Technical Assumptions and Requests

**Languages and Frameworks:**
- **Core Analysis Engine**: Go for performance-critical Git operations and file system traversal, with Rust consideration for security-sensitive parsing components
- **AI/ML Components**: Python with established AST parsing libraries (ast, tree-sitter) and security analysis frameworks
- **CLI Interface**: Go for cross-platform compatibility and single-binary distribution
- **Container Management**: Docker for sandbox isolation with Kubernetes readiness for cloud scaling

**Security Architecture:**
- **Sandbox Technology**: Docker containers with resource limits, network isolation, and read-only file system mounts
- **Code Analysis Strategy**: Static analysis only for MVP - no code execution within analyzed repositories
- **Audit Requirements**: Comprehensive logging of all repository access, analysis operations, and output generation
- **Threat Model**: Assume all analyzed repositories are potentially malicious with defense-in-depth approach

**Development Infrastructure:**
- **CI/CD Pipeline**: GitHub Actions for automated testing, security scanning, and release management
- **Dependency Management**: Go modules with vulnerability scanning integration
- **Development Environment**: Docker Compose for consistent local development with sandbox testing
- **Version Control**: Semantic versioning with automated changelog generation

**Performance and Scalability:**
- **Resource Constraints**: 2GB RAM, 4 CPU cores per analysis session with automatic cleanup
- **Concurrency Model**: Bounded parallelism for repository analysis with queue management
- **Caching Strategy**: Aggressive caching of dependency analysis and vulnerability data
- **Monitoring**: Prometheus metrics for analysis performance and resource utilization

**Data Management:**
- **Storage Requirements**: Ephemeral analysis storage with optional result persistence
- **Output Formats**: Structured Markdown with HTML export capability for web viewing
- **Configuration**: YAML-based configuration with environment variable overrides
- **Logging**: Structured JSON logging with configurable verbosity levels
