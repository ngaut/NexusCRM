# Agent-Native UI/UX Design & Interaction Model

## 1. Philosophy: "Invisible Integration"
The Agent should not be a separate "AI Tab". It must **inhabit** the existing UI. We leverage the metadata-driven frontend to inject Agent capabilities contextually.

## 2. User Interaction (The Manager)

### 2.1 The Omnipresent Copilot (Top Bar) (Implemented)
**Component**: `components/layout/TopBar.tsx` / `components/AIAssistant.tsx`
**Location**: Persistent in the global **Top Bar**.
*   **Context-Aware**: When viewing `Account/123`, the Copilot knows it.
*   **Interaction**:
    *   **Chat**: "Enrich this account." (Implemented)
    *   **Magic Fill**: Clicking "Magic Wand" icon on a Form (`MetadataRecordForm`) triggers the agent to draft fields. (Planned)

### 2.2 In-Flow Proposals (Action Injection) (Planned)
*Future Feature (Not Yet Implemented)*: Agents don't just chat; they propose state changes.
**Mechanism**:
1.  On a Record Page (`MetadataRecordDetail`), an **Agent Notification** appears: *"I found 3 missing fields for this Lead."*
2.  User clicks **"Review"**.
3.  **Diff UI**: Using the existing `MetadataRecordForm` in "Diff Mode", user sees Current vs Proposed values.
4.  **Interaction**: User accepts/rejects changes.

### 2.3 Escalation Queue (Inbox) (Planned)
**Component**: Future `EscalationList` component.
*   **Trigger**: Agent hits "Stop and Wait" (e.g., low confidence).
*   **UI**: A list of "Pending Decisions".
*   **Action**: Approve (resume agent) or Reject (cancel goal).

## 3. Admin Interaction (The Architect) (Planned)

### 3.1 Agent Studio
*Future Feature (Not Yet Implemented)*: Admins define Agents just like they define Objects.
*   **Builder UI**:
    *   **Profile**: Name, Avatar, Base Model (Claude/GPT).
    *   **Permissions**: Link to `_System_Profile`.
    *   **Skills**: Drag-and-drop enabled Skills from the Skills Registry.

### 3.2 Skills Monitor
*Future Feature (Not Yet Implemented)*: A dashboard to see "What are my agents doing?"
*   **Live Stream**: Leverages `_System_AgentActivity` audit logs.
*   **Replay**: Visual playback of Tool Calls.

## 4. Integration with Existing UI Architecture

| Agent Feature | Existing Component Leverage | Implementation Strategy |
| :--- | :--- | :--- |
| **Action Execution** | `ActionRenderer.tsx` | Agent can trigger any `_System_Action` (Send Email, Convert) virtually. |
| **Notification** | `NotificationCenter.tsx` | Agent "Help Requests" appear as standard notifications. |
| **Data View** | `MetadataRecordDetail.tsx` | Agent "working memory" looks just like a standard Record View. |
| **Forms** | `MetadataRecordForm.tsx` | Used for "Magic Fill" and "Diff Review". |

## 5. Interaction Flow Example (Planned Scenario)

**Scenario**: Data Enrichment
1.  **User (UI)**: Views `Lead` page. Clicks "Enrich" in Copilot.
2.  **Agent (Backend)**: Runs MCP tools. Finds data.
3.  **Agent (UI)**: Does NOT auto-save. Sends a **"Proposal"** event.
4.  **UI**: `MetadataRecordDetail` shows a "Proposal Banner".
5.  **User**: Clicks "Review".
6.  **UI**: Opens `MetadataRecordForm` (Pre-filled with Agent data, highlighted).
7.  **User**: Hits "Save". (Standard ACID transaction).

## 6. Design Update Summary
*   **No new "AI Portal"**.
*   **Deep Linking**: Agents live in the Top Bar and Notifications.
*   **Re-use**: We wrap `MetadataRecordForm` for AI proposals.
