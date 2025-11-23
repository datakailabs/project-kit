# .project.toml Schema Design

## Design Goals

1. **Core schema** - Universal, useful for any developer
2. **Consultant extension** - Optional, for consulting/client work
3. **DataKai extension** - Optional, for DKOS ecosystem integration
4. **Backward compatible** - Migration path from current schema

## New Schema Structure

### Core (Required)

```toml
[project]
name = "My Project"
id = "my-project"
status = "active"  # active | archived | completed | experimental
type = "product"   # product | tool | library | experiment

[tech]
stack = ["python", "fastapi", "postgresql"]
domain = ["web", "api", "backend"]

[dates]
started = "2025-11-01"
completed = ""  # Empty if ongoing

[links]
repository = "https://github.com/user/repo"
documentation = "README.md"

[notes]
description = "Brief description of the project"

# Optional but generic
[tmux]
layout = "main-vertical"
windows = [
    {name = "editor", command = "nvim"},
    {name = "server", command = "npm run dev"}
]

[context]
aws_profile = "personal"
azure_subscription = "dev"
gcloud_project = "my-project"
databricks_profile = "personal"
snowflake_account = "my-account"
git_identity = "personal"
```

### Consultant Extension (Optional)

For consultants, freelancers, agencies tracking client work.

```toml
[consultant]
# Who owns the intellectual property
ownership = "datakai" | "client" | "shared" | "open-source"

# Client information
client_name = "Acme Corp"
client_type = "direct" | "partner" | "internal"
partner = "West Monroe"  # If delivered through partner

# Role and delivery model
my_role = "lead" | "contributor" | "advisor"
deliverable_type = "product" | "consulting" | "support"
license_model = "proprietary" | "client-owned" | "open-source"

# Billing/business metadata
billable = true
rate_type = "fixed" | "hourly" | "retainer"
```

### DataKai Extension (Optional)

For DataKai ecosystem projects using DKOS, dkproto, etc.

```toml
[datakai]
# CRITICAL: Protocol variant selection (used by dkproto)
# This field determines which commit message standards, documentation rules,
# and development protocols apply to the project.
# - "private": Internal DataKai projects (lenient, supports CCC files, conventional commits with body required)
# - "public": Open source DataKai projects (imperative style, no conventional prefix)
# - "client-confidential": Client deliverables (strict, professional, no internal references)
visibility = "private" | "public" | "client-confidential"

# Knowledge/documentation integration
scriptorium_project = "acme-data-platform"  # Link to Scriptoria project notes
conduit_graph = "acme-kg"                   # Link to Conduit knowledge graph

# Protocol compliance (enforced by dkproto)
protocols = ["semantic-self-containment", "no-dry-runs"]
dkos_version = "1.0"

# Strategic product categorization
product_category = "infrastructure" | "client-deliverable" | "internal-tool"
revenue_model = "saas" | "consulting" | "open-source" | "internal"

# Development lifecycle
maturity = "experimental" | "mvp" | "production" | "deprecated"
```

## Field Definitions

### Core Fields

#### [project]
- `name` (string, required): Human-readable project name
- `id` (string, required): Machine-friendly identifier (lowercase, hyphens)
- `status` (enum, required): active | archived | completed | experimental
- `type` (enum, required): product | tool | library | experiment

#### [tech]
- `stack` (array[string], required): Technology stack (e.g., ["python", "fastapi"])
- `domain` (array[string], required): Domain categories (e.g., ["web", "api"])

#### [dates]
- `started` (date, required): ISO 8601 date (YYYY-MM-DD)
- `completed` (date, optional): ISO 8601 date, empty if ongoing

#### [links]
- `repository` (url, optional): Git repository URL
- `documentation` (path, optional): Path to documentation file

#### [notes]
- `description` (string, required): Brief project description

### Consultant Extension Fields

#### [consultant]
- `ownership` (enum, required if section present): datakai | client | shared | open-source
- `client_name` (string, optional): Client organization name
- `client_type` (enum, optional): direct | partner | internal
- `partner` (string, optional): Partner organization (e.g., "West Monroe")
- `my_role` (enum, optional): lead | contributor | advisor
- `deliverable_type` (enum, optional): product | consulting | support
- `license_model` (enum, optional): proprietary | client-owned | open-source
- `billable` (boolean, optional): Whether project is billable
- `rate_type` (enum, optional): fixed | hourly | retainer

### DataKai Extension Fields

#### [datakai]
- `visibility` (enum, **CRITICAL**): private | public | client-confidential
  - **private**: Internal DataKai projects (conventional commits with body, CCC files allowed)
  - **public**: Open source projects (imperative style, no conventional prefix)
  - **client-confidential**: Client work (strict professional standards, no internal refs)
  - **Used by dkproto for protocol variant selection**
- `scriptorium_project` (string, optional): Link to Scriptoria project notes
- `conduit_graph` (string, optional): Link to Conduit knowledge graph
- `protocols` (array[string], optional): dkproto protocols to enforce
- `dkos_version` (string, optional): DKOS version compatibility
- `product_category` (enum, optional): infrastructure | client-deliverable | internal-tool
- `revenue_model` (enum, optional): saas | consulting | open-source | internal
- `maturity` (enum, optional): experimental | mvp | production | deprecated

## Migration Strategy

### Old → New Mapping

**CRITICAL: visibility field migration**

Old location (WRONG):
```toml
[ownership]
primary = "datakai"
visibility = "private"  # ❌ Wrong section!
```

New location (CORRECT):
```toml
[consultant]
ownership = "datakai"  # Moved from ownership.primary

[datakai]
visibility = "private"  # ✅ Correct section!
```

**Current `[ownership]` section:**
```toml
[ownership]
primary = "datakai"
partners = ["westmonroe"]
license_model = "proprietary"
visibility = "private"  # ❌ Remove from here
```

**Migrates to:**
```toml
[consultant]
ownership = "datakai"  # Renamed from "primary"
partner = "westmonroe"  # Singular, first item from array
license_model = "proprietary"

[datakai]
visibility = "private"  # ✅ Moved to DataKai section
```

**Current `[client]` section:**
```toml
[client]
end_client = "Acme Corp"
intermediary = "West Monroe"
my_role = "lead"
```

**Migrates to:**
```toml
[consultant]
client_name = "Acme Corp"
client_type = "partner"
partner = "West Monroe"
my_role = "lead"
```

**Current `[links]` with DataKai fields:**
```toml
[links]
scriptorium_project = "acme"
conduit_graph = "acme-kg"
repository = "https://..."
```

**Migrates to:**
```toml
[links]
repository = "https://..."

[datakai]
scriptorium_project = "acme"
conduit_graph = "acme-kg"
```

## dkproto Protocol Variant Selection

### How visibility Field Controls Protocols

The `[datakai].visibility` field is **critical** for protocol variant selection in dkproto.

**Variant Matching Logic:**

```toml
# dkproto reads from event data:
{
  "message": "commit message",
  "datakai": {
    "visibility": "private"  # From [datakai] section
  },
  "consultant": {
    "ownership": "datakai"   # From [consultant] section
  }
}

# Protocol variant conditions:
[variants.datakai_internal]
applies_when = 'datakai.visibility == "private" && consultant.ownership == "datakai"'
# Enforces: conventional commits, body required, CCC files allowed

[variants.datakai_public]
applies_when = 'datakai.visibility == "public" && consultant.ownership == "datakai"'
# Enforces: imperative style, no conventional prefix

[variants.client_project]
applies_when = 'consultant.ownership != "datakai"'
# Enforces: professional single-line, no internal references
```

### DKOS Integration

**dk_check_commit tool reads .project.toml:**

```python
# Load project metadata
project_toml = load_toml(".project.toml")

# Extract for dkproto
project_context = {
    "datakai": {
        "visibility": project_toml.get("datakai", {}).get("visibility", "private")
    },
    "consultant": {
        "ownership": project_toml.get("consultant", {}).get("ownership", "datakai")
    }
}

# Pass to dkproto
dkproto.validate_commit(message, project_context)
```

### Why This Structure Matters

1. **Consultant ownership** (`[consultant].ownership`) determines **WHO owns the code**
   - `"datakai"`: DataKai internal projects
   - `"client"`: Client deliverables
   - Affects: commit message style, file restrictions

2. **DataKai visibility** (`[datakai].visibility`) determines **PUBLIC vs PRIVATE**
   - `"private"`: Internal projects (not publicly visible)
   - `"public"`: Open source projects (GitHub public repos)
   - `"client-confidential"`: Client projects (special handling)
   - Affects: commit style, documentation requirements

3. **Protocol variants combine both:**
   - `ownership="datakai" + visibility="private"` → Internal dev protocols
   - `ownership="datakai" + visibility="public"` → Open source protocols
   - `ownership="client"` → Client delivery protocols

## Validation Rules

1. **Core schema** - Always validated, always present
2. **[consultant]** - Validated only if section exists
   - If present, `ownership` field is required
3. **[datakai]** - Validated only if:
   - Section exists, AND
   - dkproto binary found, AND
   - Project in DKOS environment
   - If present, `visibility` field is required

## Benefits

1. **Clear separation** - Core vs extensions
2. **Universal utility** - Consultant extension useful beyond DataKai
3. **Optional complexity** - Open source users never see extensions
4. **Protocol enforcement** - dkproto validates `[datakai]` section
5. **Smooth migration** - Auto-convert on load, write new format
6. **Semantic clarity** - visibility in DataKai section (where it belongs)
