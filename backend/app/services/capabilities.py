from app.schemas.system import CapabilitySet

CORE_CAPABILITIES = [
    CapabilitySet(name="multi-agent", description="Coordinator + specialized AI agents"),
    CapabilitySet(
        name="codebase-understanding",
        description="Semantic project analysis and indexing",
    ),
    CapabilitySet(name="architecture-visualization", description="Service and dependency graphing"),
    CapabilitySet(
        name="requirement-analysis",
        description="Requirement decomposition and planning",
    ),
    CapabilitySet(name="code-generation", description="Context-aware implementation support"),
    CapabilitySet(name="code-review", description="Automated review with policy checks"),
    CapabilitySet(name="documentation-generation", description="Living docs from source context"),
    CapabilitySet(
        name="test-case-generation",
        description="Targeted test generation by change set",
    ),
    CapabilitySet(
        name="change-impact-analysis",
        description="Predict downstream impact from edits",
    ),
    CapabilitySet(
        name="rag-knowledge-retrieval",
        description="Repository and docs retrieval via embeddings",
    ),
    CapabilitySet(
        name="mcp-integration",
        description="Model Context Protocol tool interoperability",
    ),
    CapabilitySet(name="plugin-system", description="Extensible plugins for custom workflows"),
    CapabilitySet(name="rest-apis", description="Versioned API surface for platform features"),
]
