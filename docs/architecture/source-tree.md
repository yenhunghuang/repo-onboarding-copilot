# Source Tree

## Project Structure

```
repo-onboarding-copilot/
├── cmd/                          # CLI entry points
│   ├── analyze/                  # Analysis command implementation
│   ├── generate/                 # Generation command implementation
│   └── server/                   # API server entry point (future)
├── internal/                     # Private application code
│   ├── analysis/                 # Core analysis engine
│   │   ├── ast/                 # AST parsing and analysis
│   │   ├── security/            # Security scanning components
│   │   ├── metrics/             # Code metrics and quality analysis
│   │   └── orchestrator/        # Analysis workflow coordination
│   ├── generator/               # Documentation generation engine
│   │   ├── templates/           # Output templates (markdown, html)
│   │   ├── visualizer/          # Architecture diagram generation
│   │   └── formatter/           # Output formatting and styling
│   ├── security/                # Security and sandboxing
│   │   ├── sandbox/             # Container isolation management
│   │   ├── validator/           # Input validation and sanitization
│   │   └── scanner/             # Vulnerability scanning integration
│   ├── storage/                 # Data persistence layer
│   │   ├── cache/               # Caching implementation
│   │   ├── database/            # Database abstraction
│   │   └── filesystem/          # File system operations
│   └── api/                     # API layer (future web interface)
│       ├── handlers/            # HTTP request handlers
│       ├── middleware/          # Authentication, logging, etc.
│       └── routes/              # Route definitions
├── pkg/                         # Public library code
│   ├── config/                  # Configuration management
│   ├── logger/                  # Structured logging
│   ├── utils/                   # Utility functions
│   └── types/                   # Shared type definitions
├── web/                         # Frontend code (future)
│   ├── src/                     # React application source
│   ├── public/                  # Static assets
│   └── dist/                    # Built frontend assets
├── deployments/                 # Deployment configurations
│   ├── docker/                  # Docker configurations
│   ├── kubernetes/              # K8s manifests (future)
│   └── compose/                 # Docker Compose files
├── scripts/                     # Build and deployment scripts
│   ├── build.sh                 # Build automation
│   ├── test.sh                  # Testing automation
│   └── deploy.sh                # Deployment automation
├── docs/                        # Project documentation
│   ├── architecture/            # Architecture documentation
│   ├── api/                     # API documentation (future)
│   └── user-guide/              # User documentation
├── test/                        # Test files and test data
│   ├── integration/             # Integration test suites
│   ├── fixtures/                # Test data and fixtures
│   └── mocks/                   # Test doubles and mocks
├── configs/                     # Configuration files
│   ├── development.yaml         # Development configuration
│   ├── production.yaml          # Production configuration
│   └── testing.yaml             # Testing configuration
├── .github/                     # GitHub workflows and templates
│   ├── workflows/               # CI/CD pipeline definitions
│   └── templates/               # Issue and PR templates
├── .docker/                     # Docker-related files
│   ├── Dockerfile               # Main application image
│   ├── Dockerfile.dev           # Development image
│   └── docker-compose.yml       # Local development setup
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
├── Makefile                     # Build automation
├── README.md                    # Project overview
└── LICENSE                      # License file
```

## Key Design Principles

```yaml
Package Organization:
    - Domain-driven design with clear bounded contexts
    - Internal packages for implementation details
    - Public packages (pkg/) for reusable components
    - Clear separation between CLI, core logic, and future API

Dependency Management:
    - Import restrictions: internal packages cannot import cmd/
    - Clean architecture: dependencies point inward
    - Interface-based design for testability
    - Minimal external dependencies

Scalability Considerations:
    - Modular design enables microservices migration
    - Clear API boundaries for service extraction
    - Stateless design for horizontal scaling
    - Configuration-driven behavior
```
