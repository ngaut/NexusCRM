# NexusCRM UI/UX Test Scenarios

> Specific, actionable test scenarios for manual UI testing.

---

## 1. Login & Authentication

### Scenario: Successful Login
1. Navigate to `http://localhost:5173/`
2. Enter email: `admin@test.com`
3. Enter password: `Admin123!`
4. Click **"Log In"** button
5. **Expected**: Redirects to Dashboard, top-right shows user avatar

### Scenario: Invalid Credentials
1. Enter email: `wrong@test.com`, password: `wrong123`
2. Click **"Log In"**
3. **Expected**: Red error message: "Invalid credentials"

### Scenario: Session Persistence
1. Login successfully
2. Refresh browser (F5)
3. **Expected**: Still logged in, not redirected to login

---

## 2. Sidebar Navigation

### Scenario: Navigate to Object List
1. In sidebar, click **"Accounts"** (or first object in nav)
2. **Expected**: URL changes to `/object/account`, list view loads

### Scenario: Collapse/Expand Sidebar
1. Click **chevron icon** at bottom of sidebar
2. **Expected**: Sidebar collapses to ~64px width, only icons visible
3. Hover over any icon
4. **Expected**: Tooltip shows label (e.g., "Accounts")
5. Click chevron again
6. **Expected**: Sidebar expands, labels visible

### Scenario: App Launcher
1. Click **grid icon** (top-left, 9-dot icon)
2. **Expected**: Modal opens showing available apps
3. Click a different app
4. **Expected**: Navigation items change to that app's items

---

## 3. Object List View

### Test Object: Account

### Scenario: View Account List
1. Navigate to **/object/account**
2. **Expected**: See table with columns: Name, Industry, Phone, Owner
3. **Expected**: Pagination shows "1-25 of X"

### Scenario: Sort by Column
1. Click **"Name"** column header
2. **Expected**: Rows sort A-Z, arrow indicator shows on header
3. Click again
4. **Expected**: Rows sort Z-A

### Scenario: Search Records
1. Type "Acme" in search box
2. **Expected**: List filters to show only accounts containing "Acme"

### Scenario: Create New Account
1. Click **"New"** button (top right)
2. In modal, enter:
   - **Name**: "Test Corp"
   - **Industry**: "Technology"
   - **Phone**: "555-0100"
3. Click **"Save"**
4. **Expected**: Modal closes, success toast "Account created", new record appears in list

### Scenario: Required Field Validation
1. Click **"New"** button
2. Leave **Name** empty
3. Click **"Save"**
4. **Expected**: Error message under Name field: "Name is required"

---

## 4. Record Detail View

### Test Record: Any Account

### Scenario: View Record Details
1. Click on account "Test Corp" in list
2. **Expected**: Detail page shows:
   - Highlights panel with key fields
   - Full record form with all fields
   - Related Lists section (Contacts, Opportunities)

### Scenario: Edit Record
1. Click **"Edit"** button
2. Change **Industry** from "Technology" to "Finance"
3. Click **"Save"**
4. **Expected**: Toast "Record updated", field shows new value

### Scenario: Inline Edit
1. On detail page, click directly on **Phone** field value
2. **Expected**: Field becomes editable inline
3. Change value, press Enter
4. **Expected**: Value saves automatically

### Scenario: Delete Record
1. Click **"Delete"** button
2. **Expected**: Confirmation modal: "Are you sure you want to delete Test Corp?"
3. Click **"Delete"** in modal
4. **Expected**: Redirects to list, toast "Record deleted"

---

## 5. Dashboard

### Scenario: View Dashboard Widgets
1. Navigate to **/dashboard** or click Dashboard in sidebar
2. **Expected**: See widgets:
   - **Metric widgets**: Show numbers (e.g., "Total Revenue: $125,000")
   - **Chart widgets**: Show bar/pie/line charts with legends
   - **Record list widgets**: Show 5-10 recent records

### Scenario: Widget Types Render Correctly
- **metric** widget → Shows single number with label
- **chart-bar** widget → Shows vertical bar chart
- **chart-pie** widget → Shows pie chart with segments
- **record-list** widget → Shows table of records

---

## 6. AI Assistant (Agentic AI)

### Scenario: Open AI Panel
1. Click **"AI"** button with sparkle icon (header, right side)
2. **Expected**: Panel slides in from right (~400px wide)

### Scenario: Create Record via AI
1. Type: `Create a new contact named Jane Smith with email jane@test.com`
2. Press Enter
3. **Expected**: 
   - AI shows "Thinking..." indicator
   - Tool call card appears: `create_record` with parameters
   - AI confirms: "Created contact Jane Smith"
4. Navigate to Contacts list
5. **Expected**: "Jane Smith" appears in list

### Scenario: Query Records via AI
1. Type: `Show me all accounts in Technology industry`
2. **Expected**:
   - AI executes `query_object` tool
   - Results displayed in chat as table or list

### Scenario: Update Record via AI
1. Type: `Update Jane Smith's phone to 555-0199`
2. **Expected**:
   - AI searches for contact
   - Executes update
   - Confirms: "Updated Jane Smith's phone"

### Scenario: Create Object Schema via AI
1. Type: `Create a new object called Project with fields: name, status, due_date`
2. **Expected**:
   - AI creates schema
   - AI creates each field
   - Confirms object is ready

### Scenario: AI Navigation
1. Type: `Take me to the Account list`
2. **Expected**: Browser navigates to `/object/account`

### Scenario: AI Context Awareness
1. Open a Contact record detail page
2. Open AI panel
3. Type: `What's this person's email?`
4. **Expected**: AI returns the current contact's email

### Scenario: AI Admin Tasks
1. Type: `Create a new user for bob@company.com with password Temp123!`
2. **Expected**: AI creates user, confirms email

---

## 7. App Studio

### Scenario: Open Studio
1. Navigate to **/studio/{app_id}** or click "App Studio" in Setup
2. **Expected**: Studio loads with sidebar showing objects/dashboards

### Scenario: Edit Object Fields
1. Click an object in sidebar
2. **Expected**: See list of fields with types
3. Click **"+ Add Field"**
4. **Expected**: Field wizard opens

### Scenario: Create Field
1. In wizard, select **"Text"** field type
2. Enter Label: "Nickname"
3. Click **"Create"**
4. **Expected**: Field added to object

### Scenario: Edit Dashboard
1. Click a dashboard in sidebar
2. **Expected**: Dashboard canvas loads
3. Drag **"Metric"** widget from left palette
4. Drop onto canvas
5. **Expected**: Widget appears at drop location
6. Double-click widget
7. **Expected**: Configuration panel opens

---

## 8. Setup Pages

### Scenario: User Management
1. Navigate to **Setup → Users**
2. **Expected**: See user list with columns: Name, Email, Profile, Active
3. Click **"New User"**
4. Fill: Name, Email, Password, Profile
5. **Save**
6. **Expected**: User created

### Scenario: Profile Permissions
1. Navigate to **Setup → Profiles**
2. Click a profile (e.g., "Standard User")
3. **Expected**: See object permissions table
4. Toggle **"Create"** permission for Account
5. **Save**
6. **Expected**: Permission saved

### Scenario: Approval Queue
1. Navigate to **/approvals**
2. **Expected**: See list of pending approvals
3. Click **"Approve"** on any item
4. **Expected**: Modal asks for comments, then approves

---

## 9. Error States

### Scenario: 404 Record Not Found
1. Navigate to `/object/account/nonexistent-id`
2. **Expected**: "Record not found" message with "Go Back" button

### Scenario: Network Error
1. Disconnect network
2. Try to save a record
3. **Expected**: Error toast "Connection failed. Check your network."

### Scenario: Permission Denied
1. Login as limited user
2. Try to access admin page
3. **Expected**: "Access Denied" or redirect

---

## 10. Responsive Design

### Scenario: Mobile Width (375px)
1. Resize browser to 375px width
2. **Expected**: 
   - Sidebar hidden
   - Hamburger menu (☰) appears
3. Tap hamburger
4. **Expected**: Sidebar slides in as overlay

---

## Test Environment

| Item | Value |
|------|-------|
| URL | `http://localhost:5173/` |
| Email | `admin@test.com` |
| Password | `Admin123!` |
| Backend | `http://localhost:3001/` |

---

*Last updated: 2025-12-31*
