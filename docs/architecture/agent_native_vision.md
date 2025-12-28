# Agent-Native Architecture: Layered Design & Governance (v3)

## Core Philosophy
NexusCRM is a **Fully Dynamic Platform**. Objects, fields, and workflows are created at runtime. Therefore, Agents cannot rely on hardcoded tool definitions. They must **Dynamic Discovery** to navigate the ever-changing environment.

---

## 1. The Layered Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Layer 4: Interaction Plane (Chat, Voice, Background Flows)│
├─────────────────────────────────────────────────────────────┤
│  Layer 3: Agent Runtime (Discovery, Reasoning, Skills)     │
├─────────────────────────────────────────────────────────────┤
│  Layer 2: DYNAMIC Tool Bus (The "Live" Bridge)             │
├─────────────────────────────────────────────────────────────┤
│  Layer 1: Foundation (Persistence, Flow, Metadata)         │
└─────────────────────────────────────────────────────────────┘
```

### Layer 1: Foundation (The Dynamic Core)
- **Role**: Truth & Execution.
- **Dynamic Nature**: Users define new Objects (`Car`), Fields (`VIN`), and Logic (`CarRepairFlow`) at runtime.
- **Agent Constraint**: Agents MUST use standard APIs.

### Layer 2: The Dynamic Tool Bus (The "Magic" Layer)
- **Role**: **Live Projection** of system capabilities.
- **Mechanism**: This layer is NOT a static config file. It is a real-time translator.
- **tool_definitions.json** is generated **On-Demand**:
    1.  Agent asks: "What can I do?"
    2.  Tool Bus queries `MetadataService`.
    3.  Tool Bus sees `Car` object exists.
    4.  Tool Bus *instantly* generates `create_car`, `search_car`, `update_car` tool definitions.
    5.  Agent receives full toolset, including objects created 5 seconds ago.

### Layer 3: Agent Runtime (The Explorer)
- **Role**: Discovery & Execution.
- **The "Context Limit" Problem**: A full CRM has 500+ objects. We cannot feed *all* tools to the LLM at once.
- **Solution: Introspection First**:
    - Agents are equipped with **Meta-Tools**: `list_all_objects()`, `describe_object(name)`, `search_tools(query)`.
    - **Discovery Loop**:
        1.  **Goal**: "Register a new vehicle."
        2.  **Introspection**: "I don't see a `create_vehicle` tool in my main set. I will call `search_tools('vehicle')`."
        3.  **Result**: "Found tool `create_Car__c` and object `Car__c`."
        4.  **Loading**: Agent now understands how to use `create_Car__c`.

---

## 2. How Agents Know About "New" Things

This is the critical "Agent-Native" differentiator.

### Scenario: The "Day 1" Problem
*Admin creates a completely new `Conference` object with `Ticket_Price` field and a `BookTicket` flow.*

**Legacy Way**: Developer writes a new plugin/connector.
**Agent-Native Way**:
1.  **Metadata Event**: Admin saves `Conference` object.
2.  **Tool Bus Update**: The Tool Bus *immediately* reflects `create_Conference` and `run_flow_BookTicket` as available capabilities.
3.  **Agent Discovery**:
    *   **Option A (Active Search)**: User asks "Book a conference." Agent searches tools for "conference", finds `create_Conference`, and asks user for `Ticket_Price`.
    *   **Option B (Skill Trigger)**: Admin adds a new Skill "Manage Events" linked to `Conference`. Agent loads this skill and *automatically* knows how to handle the new object.

---

## 3. Role Relationships & Governance

| Role | Responsibilities | Permissions |
| :--- | :--- | :--- |
| **Admin** | Defines the Schema (The "Physics" of the world). | Full System Access. |
| **User** | Sets Goals. | Standard Access. |
| **Agent** | **Explores** the Schema to fulfill Goals. | **Strictly Bound** by FLS/Sharing. |

**Crucial**: Even if an Agent *discovers* a tool (`delete_Invoice`), it cannot execute it if the linked User Profile lacks `DELETE` permission on `Invoice`.

---

## 4. Skills & Capabilities

**Agent Skills** (via AgentSkills.io pattern) are the "Instruction Manuals" for the Tools.

*   **Tools** (Layer 2) = "I *can* hammer a nail." (Capability)
*   **Skills** (Layer 3) = "Here is *how* to build a chair using a hammer." (Procedure)

When a new Object is created, the Tool is automatic. The Skill (best practices for using that object) can be:
1.  **Authored**: Admin writes a quick `manage-conference.md` skill.
2.  **Inferred**: Agent uses generic "CRUD Skill" to make reasonable guesses based on field names (`Start_Date` usually comes before `End_Date`).

---

## 5. Migration Strategy

1.  **Layer 1 (Done)**: The Foundation exists.
2.  **Layer 2 (Next)**: Build the **MCP Server** that exposes `MetadataService` as live tools.
    *   Implement `tools/list` to return `_System_Object` as CRUD tools.
    *   Implement `tools/call` to route requests to `PersistenceService`.
3.  **Verification**: Connect Claude Desktop to NexusCRM. Create a custom object in NexusCRM. Verify Claude *immediately* sees it and can interact with it.
