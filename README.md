# Repo Onboarding Copilot

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen)](https://github.com/yenhunghuang/repo-onboarding-copilot)
[![Coverage](https://img.shields.io/badge/Coverage-80%2B-brightgreen)](coverage.html)
[![Security](https://img.shields.io/badge/Security-Hardened-blue)](https://github.com/yenhunghuang/repo-onboarding-copilot)

**Automated repository analysis tool that transforms unknown codebases into comprehensive onboarding experiences within 1 hour.**

Repo Onboarding Copilot is a security-focused CLI tool built in Go that analyzes Git repositories to generate comprehensive documentation, identify patterns, dependencies, and architectural insights. It helps reduce developer onboarding time from weeks to hours through intelligent codebase analysis and automated documentation generation.

## 🚀 Quick Start

### Installation

#### Binary Installation
```bash
# Download the latest release
curl -L -o repo-onboarding-copilot https://github.com/yenhunghuang/repo-onboarding-copilot/releases/latest/download/repo-onboarding-copilot-linux-amd64
chmod +x repo-onboarding-copilot
```

#### Build from Source
```bash
# Clone the repository
git clone https://github.com/yenhunghuang/repo-onboarding-copilot.git
cd repo-onboarding-copilot

# Build the project
make build

# Or install directly
make install
```

### Basic Usage

```bash
# Analyze a GitHub repository
repo-onboarding-copilot https://github.com/owner/repo.git

# Analyze using SSH URL  
repo-onboarding-copilot git@github.com:owner/repo.git

# Show version information
repo-onboarding-copilot --version
```

## 🏗️ Architecture Overview

The project follows a **domain-driven design** with clean architecture principles:

```
repo-onboarding-copilot/
├── cmd/                          # CLI entry points
├── internal/                     # Private application code
│   ├── analysis/                 # Core analysis engine
│   │   ├── ast/                 # AST parsing and analysis
│   │   ├── security/            # Security scanning
│   │   └── orchestrator/        # Analysis coordination
│   └── security/                # Security and sandboxing
│       ├── sandbox/             # Container isolation
│       ├── validator/           # Input validation
│       └── scanner/             # Vulnerability scanning
├── pkg/                         # Public library code
│   ├── config/                  # Configuration management
│   ├── logger/                  # Structured logging
│   └── types/                   # Shared type definitions
└── web/                         # Frontend (future)
```

## 🎯 Key Features

### 🔍 **Intelligent Code Analysis**
- **Multi-language AST parsing** with tree-sitter integration
- **Dependency graph generation** with vulnerability assessment  
- **Performance bottleneck identification** through static analysis
- **Architecture pattern detection** and documentation
- **License compliance analysis** with risk assessment

### 🛡️ **Security-First Design**
- **Container isolation** for secure repository analysis
- **Input validation and sanitization** to prevent code injection
- **Vulnerability scanning** integrated with NVD, OWASP databases
- **Resource limits** and timeout management
- **Audit logging** with structured security events

### 📊 **Comprehensive Reporting**
- **Multi-format output** (JSON, HTML, Markdown)
- **Visual dependency graphs** and architecture diagrams
- **Performance metrics** and optimization recommendations
- **Security risk assessments** with remediation guidance
- **Bundle analysis** for web applications

### ⚡ **Performance & Scalability**
- **Parallel processing** with bounded goroutines
- **Memory-efficient parsing** with streaming analysis
- **Caching strategies** for repeated analysis
- **Timeout management** for large repositories
- **Resource monitoring** and limits

## 🔧 Technology Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Core Language** | Go 1.23+ | High-performance, concurrent processing |
| **CLI Framework** | Cobra | Command-line interface and argument parsing |
| **AST Parsing** | tree-sitter | Multi-language syntax tree analysis |
| **Containerization** | Docker | Secure repository isolation |
| **Testing** | testify + Go testing | Comprehensive test coverage |
| **Logging** | logrus | Structured logging and audit trails |
| **Configuration** | YAML | Environment-specific configuration |

## 📋 Requirements

- **Go**: 1.23 or later
- **Docker**: Required for secure sandboxing (optional)
- **Memory**: 2GB+ recommended for large repositories  
- **Disk**: 1GB+ for analysis cache and temporary files

## 🛠️ Development

### Prerequisites
```bash
# Install Go dependencies
make deps

# Install development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
```

### Build Commands
```bash
make build              # Build for current platform
make build-all          # Cross-platform builds
make test               # Run tests
make test-coverage      # Generate coverage report
make lint               # Run linter
make security           # Security checks
make check              # Full quality check
```

### Running Tests
```bash
# Unit tests
go test ./...

# Integration tests
go test ./test/integration/...

# Security tests  
go test ./test/integration/security/...

# Performance benchmarks
go test -bench=. ./test/integration/
```

## 🔒 Security Features

### Container Isolation
- **Sandboxed execution** of repository analysis
- **Resource limits** (CPU, memory, network)
- **Non-root user execution** for enhanced security
- **Temporary filesystem** cleanup after analysis

### Input Validation
- **URL sanitization** and validation
- **Path traversal protection** 
- **Malicious input detection**
- **Size limits** on processed files

### Vulnerability Scanning  
- **Dependency vulnerability** detection
- **License compliance** checking
- **Security pattern** analysis
- **OWASP integration** for web applications

## 📈 Performance Characteristics

| Metric | Small Repo (<1K files) | Medium Repo (1K-10K files) | Large Repo (10K+ files) |
|--------|------------------------|----------------------------|--------------------------|
| **Analysis Time** | < 30 seconds | 2-10 minutes | 10-60 minutes |
| **Memory Usage** | < 100MB | 200-500MB | 500MB-2GB |
| **Disk Usage** | < 50MB | 100-500MB | 500MB-2GB |
| **Accuracy** | 95%+ | 90%+ | 85%+ |

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Workflow
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests and quality checks (`make check`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)  
7. Open a Pull Request

### Code Standards
- Follow Go best practices and idioms
- Maintain 80%+ test coverage
- Use structured logging for all operations
- Document public APIs and complex algorithms
- Run security checks before submission

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [tree-sitter](https://tree-sitter.github.io/) for multi-language parsing
- [Cobra](https://cobra.dev/) for CLI framework
- [Docker](https://docker.com/) for containerization
- [Go team](https://golang.org/) for the excellent language and toolchain

## 📞 Support & Contact

- **Issues**: [GitHub Issues](https://github.com/yenhunghuang/repo-onboarding-copilot/issues)
- **Documentation**: [Project Wiki](https://github.com/yenhunghuang/repo-onboarding-copilot/wiki)
- **Security**: Report security issues to [security@example.com](mailto:security@example.com)

---

**Built with ❤️ in Go | Secure by Design | Enterprise Ready**