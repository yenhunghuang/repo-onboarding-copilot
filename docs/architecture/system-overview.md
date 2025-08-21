# System Overview

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              Repo Onboarding Copilot                           │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌─────────────────┐    ┌──────────────────┐    ┌─────────────────────────────┐ │
│  │   CLI Interface │    │  Web API Gateway │    │    Future Web Dashboard     │ │
│  │                 │    │                  │    │                             │ │
│  │ • Commands      │    │ • REST Endpoints │    │ • Analysis Visualization    │ │
│  │ • Progress      │    │ • Authentication │    │ • Interactive Reports       │ │
│  │ • Results       │    │ • Rate Limiting  │    │ • Team Collaboration       │ │
│  └─────────────────┘    └──────────────────┘    └─────────────────────────────┘ │
│           │                       │                            │                 │
│           └───────────────────────┼────────────────────────────┘                 │
│                                   │                                              │
│  ┌─────────────────────────────────┼─────────────────────────────────────────────┤
│  │                      Core Application Layer                                  │
│  ├─────────────────────────────────┼─────────────────────────────────────────────┤
│  │                                 │                                              │
│  │  ┌────────────────┐    ┌─────────────────┐    ┌─────────────────────────────┐ │
│  │  │ Analysis       │    │ Security        │    │ Output Generation           │ │
│  │  │ Orchestrator   │    │ Sandbox         │    │ Engine                      │ │
│  │  │                │    │                 │    │                             │ │
│  │  │ • Job Queue    │    │ • Docker        │    │ • Documentation Generator   │ │
│  │  │ • Progress     │    │ • Isolation     │    │ • Visualization Engine      │ │
│  │  │ • Coordination │    │ • Resource      │    │ • Multi-format Export       │ │
│  │  │ • Caching      │    │   Limits        │    │ • Template System           │ │
│  │  └────────────────┘    └─────────────────┘    └─────────────────────────────┘ │
│  │           │                      │                            │               │
│  │           └──────────────────────┼────────────────────────────┘               │
│  │                                  │                                            │
│  │  ┌────────────────────────────────┼────────────────────────────────────────────┤
│  │  │                    Analysis Engine Layer                                  │
│  │  ├────────────────────────────────┼────────────────────────────────────────────┤
│  │  │                                │                                            │
│  │  │  ┌──────────────┐  ┌─────────────────┐  ┌─────────────┐  ┌──────────────┐ │
│  │  │  │ Git Handler  │  │ AST Parser      │  │ Dependency  │  │ Security     │ │
│  │  │  │              │  │                 │  │ Analyzer    │  │ Scanner      │ │
│  │  │  │ • Clone      │  │ • JavaScript    │  │             │  │              │ │
│  │  │  │ • Validate   │  │ • TypeScript    │  │ • Tree      │  │ • Vulns      │ │
│  │  │  │ • Metadata   │  │ • Tree-sitter   │  │ • Audit     │  │ • Patterns   │ │
│  │  │  │ • Cleanup    │  │ • Babylon       │  │ • Licenses  │  │ • Config     │ │
│  │  │  └──────────────┘  └─────────────────┘  └─────────────┘  └──────────────┘ │
│  │  │                                │                                            │
│  │  └────────────────────────────────┼────────────────────────────────────────────┘
│  │                                   │                                              │
│  │  ┌────────────────────────────────┼────────────────────────────────────────────┐ │
│  │  │                     Data & Storage Layer                                   │ │
│  │  ├────────────────────────────────┼────────────────────────────────────────────┤ │
│  │  │                                │                                            │ │
│  │  │  ┌──────────────┐  ┌─────────────────┐  ┌─────────────┐  ┌──────────────┐ │ │
│  │  │  │ Analysis     │  │ Vulnerability   │  │ Template    │  │ Cache        │ │ │
│  │  │  │ Results      │  │ Database        │  │ Store       │  │ Layer        │ │ │
│  │  │  │              │  │                 │  │             │  │              │ │ │
│  │  │  │ • Metadata   │  │ • CVE Data      │  │ • Runbooks  │  │ • Parsed     │ │ │
│  │  │  │ • AST Data   │  │ • OWASP         │  │ • Arch Diagrams  │ • Results    │ │ │
│  │  │  │ • Metrics    │  │ • Updates       │  │ • Roadmaps  │  │ • Vulns      │ │ │
│  │  │  │ • Artifacts  │  │ • Feeds         │  │ • Styles    │  │ • Deps       │ │ │
│  │  │  └──────────────┘  └─────────────────┘  └─────────────┘  └──────────────┘ │ │
│  │  │                                                                            │ │
│  │  └────────────────────────────────────────────────────────────────────────────┘ │
│  │                                                                                  │
│  └──────────────────────────────────────────────────────────────────────────────────┘
│                                                                                    │
└────────────────────────────────────────────────────────────────────────────────────┘
```

## Core Components

1. **CLI Interface**: Primary user interaction point with command processing and result display
2. **Analysis Orchestrator**: Central coordinator managing analysis workflows and job execution
3. **Security Sandbox**: Container-based isolation ensuring safe repository analysis
4. **Analysis Engine**: Modular analyzers for code parsing, dependency analysis, and security scanning
5. **Output Generator**: Template-driven documentation and visualization engine
6. **Data Layer**: Persistent storage for results, templates, and vulnerability data
