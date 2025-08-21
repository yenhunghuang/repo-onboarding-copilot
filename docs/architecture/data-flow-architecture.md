# Data Flow Architecture

## Analysis Pipeline

```
Repository URL Input
        ↓
    Validation & Sanitization
        ↓
    Secure Container Creation
        ↓
    Git Clone & Repository Extraction
        ↓
    ┌─────────────────────────────────────────────────────────┐
    │              Parallel Analysis Phase                    │
    ├─────────────────────────────────────────────────────────┤
    │                                                         │
    │  ┌──────────────┐  ┌─────────────┐  ┌─────────────────┐ │
    │  │ AST Parsing  │  │ Dependency  │  │ Security        │ │
    │  │              │  │ Analysis    │  │ Scanning        │ │
    │  │ • Functions  │  │             │  │                 │ │
    │  │ • Classes    │  │ • Tree      │  │ • Vulnerabilities│ │
    │  │ • Imports    │  │ • Vulns     │  │ • Patterns      │ │
    │  │ • Exports    │  │ • Licenses  │  │ • Configuration │ │
    │  └──────────────┘  └─────────────┘  └─────────────────┘ │
    │          │               │                    │         │
    │          └───────────────┼────────────────────┘         │
    │                          │                              │
    └──────────────────────────┼──────────────────────────────┘
                               ↓
                    Results Aggregation
                               ↓
                    Quality Analysis & Scoring
                               ↓
    ┌─────────────────────────────────────────────────────────┐
    │              Documentation Generation                   │
    ├─────────────────────────────────────────────────────────┤
    │                                                         │
    │  ┌──────────────┐  ┌─────────────┐  ┌─────────────────┐ │
    │  │ Runbook      │  │ Architecture│  │ Learning        │ │
    │  │ Generation   │  │ Diagrams    │  │ Roadmap         │ │
    │  │              │  │             │  │                 │ │
    │  │ • Setup      │  │ • Component │  │ • 30-day plan   │ │
    │  │ • Scripts    │  │   Map       │  │ • 60-day goals  │ │
    │  │ • Validation │  │ • Data flow │  │ • 90-day target │ │
    │  │ • Troubleshoot│ │ • Dependencies│ │ • Milestones   │ │
    │  └──────────────┘  └─────────────┘  └─────────────────┘ │
    │                                                         │
    └─────────────────────────────────────────────────────────┘
                               ↓
                    Multi-format Export
                               ↓
                    Container Cleanup & Audit
```

## Data Models

### Core Analysis Result

```json
{
    "analysis_id": "uuid",
    "repository": {
        "url": "string",
        "commit_hash": "string",
        "size_bytes": "number",
        "languages": ["string"],
        "frameworks": ["string"]
    },
    "analysis_metadata": {
        "started_at": "timestamp",
        "completed_at": "timestamp",
        "duration_seconds": "number",
        "version": "string"
    },
    "code_analysis": {
        "ast_data": "object",
        "component_map": "object",
        "complexity_metrics": "object",
        "quality_score": "number"
    },
    "dependencies": {
        "direct": ["object"],
        "transitive": ["object"],
        "vulnerabilities": ["object"],
        "licenses": ["object"]
    },
    "security_findings": {
        "vulnerabilities": ["object"],
        "risk_score": "number",
        "compliance_status": "object"
    },
    "documentation": {
        "runbook": "string",
        "architecture_diagrams": ["string"],
        "learning_roadmap": "object"
    }
}
```
