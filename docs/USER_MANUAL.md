# NexusCRM User Manual

Welcome to **NexusCRM**, a metadata-driven CRM platform designed to manage your customer relationships and business processes.

---

## 💡 Quick Start

### Default Credentials
- **Email**: `admin@test.com`
- **Password**: `Admin123!`
- ⚠️ **Change password immediately after first login**

### Accessing the Application
1. **Easy Start**: Run `npm run dev:full` from the project root (starts both backend and frontend).
2. **Alternative**:
   - Backend: `npm run dev:server` (or `cd backend && go run cmd/server/main.go`)
   - Frontend: `npm run dev:client`
3. Open browser: `http://localhost:5173`
4. Login with default credentials

---

## 📚 Table of Contents

1. [Getting Started](#1-getting-started)
2. [Navigation](#2-navigation)
3. [Working with Data](#3-working-with-data)
4. [Analytics & Reporting](#4-analytics--reporting)
5. [Administrator Guide](#5-administrator-guide)
6. [Security & Permissions](#6-security--permissions)
7. [Tips & Best Practices](#7-tips--best-practices)
8. [Business Processes](#8-business-processes)
9. [Getting Help](#9-getting-help)

---

## 1. Getting Started

### First Login
1. Navigate to `http://localhost:5173`
2. Enter default credentials
3. You'll be directed to the main dashboard

### Changing Your Password
1. Click your profile icon (top-right)
2. Select "Settings"
3. Enter current password
4. Enter new strong password (12+ characters, mixed case, numbers, special chars)
5. Save changes

---

## 2. Navigation

### Sidebar
The left sidebar is your primary navigation:
- **Apps**: Switch between Sales, Marketing, Service, etc.
- **Tabs**: Object list views (Accounts, Contacts, Opportunities)
- **Dashboards**: Analytics and KPIs
- **Setup**: Admin configuration (if you're an admin)

### Global Search
- Located at the top of the page
- Search across all objects
- Type record name, email, or any field value
- Results show matching records with quick actions

### App Launcher
- Click the grid icon (top-left)
- Browse available applications
- Switch between different business areas

---

## 3. Working with Data

### List Views

**Viewing Records**:
- Navigate to any object tab (e.g., Accounts)
- See all records you have access to
- Click any record to view details

**Filtering**:
- Click "Filter" icon
- Add conditions (e.g., `Industry equals Technology`)
- Combine multiple filters with AND/OR logic
- Save filters for reuse

**Sorting**:
- Click column headers to sort
- Toggle ascending/descending order

**Kanban View** (for picklist fields like Stage, Status):
- Click "Kanban" toggle
- Drag and drop cards between columns
- Changes automatically save

### Creating Records

1. Click "New" button on list view
2. Fill required fields (marked with red asterisk)
3. Optional: Fill additional fields
4. Click "Save"
5. Record is created and you're redirected to detail view

### Editing Records

**Method 1: Inline Edit**
- Click pencil icon next to field
- Edit value
- Changes save automatically

**Method 2: Edit Mode**
- Click "Edit" button at top
- Modify multiple fields
- Click "Save" when done

### Deleting Records

**Soft Delete** (Recoverable):
1. Open record detail page
2. Click "Delete" button
3. Confirm deletion
4. Record moves to Recycle Bin





## 4. Analytics & Reporting

### Dashboards

**Viewing Dashboards**:
- Click "Dashboards" tab
- Select dashboard from list
- View real-time metrics and charts

**Dashboard Widgets**:
- **Metrics**: Key numbers (Total Revenue, Deal Count)
- **Charts**: Visual data (Bar, Pie, Line charts)
- **Tables**: Lists of records meeting criteria
- **SQL Analytics**: Advanced SQL queries with visualization (Admin only)

**Filtering Dashboards**:
- Use date range picker
- Apply filters to focus on specific data
- Filters apply to all widgets

### Creating Reports

1. Navigate to Reports tab
2. Click "New Report"
3. Select data source (Object)
4. Choose fields to include
5. Add filters (optional)
6. Add grouping (optional)
7. Choose chart type (Bar, Pie, Table)
8. Save report

---

## 5. Administrator Guide

NexusCRM is a **metadata-driven platform** - you can customize without code.

### Object Manager

**Creating Custom Objects**:
1. Navigate to Setup → Object Manager
2. Click "New Object"
3. Enter:
   - Label (e.g., "Project")
   - API Name (auto-generated: `Project__c`)
   - Description
4. Save
5. New object appears in database and UI

**Adding Custom Fields**:
1. Open Object Manager
2. Select object
3. Click "New Field"
4. Choose field type:
   - Text, Number, Currency
   - Email, Phone, URL
   - Picklist (dropdown)
   - Checkbox (true/false)
   - Date, DateTime
   - Lookup (relationship to another object)
   - Formula (calculated field)
5. Configure field properties
6. Save
7. Field appears in layouts and forms

**Field Types Explained**:
- **Formula**: Calculated using Excel-like syntax (e.g., `Amount * 0.1` for 10% commission)
- **Lookup**: Creates relationship to another object (e.g., Account lookup on Contact)
- **Picklist**: Dropdown with predefined options

### Page Layouts

**Customizing Record Pages**:
1. Setup → Object Manager → [Object] → Layouts
2. Drag and drop sections
3. Add/remove fields
4. Reorder fields
5. Set field required status
6. Save layout
7. Changes appear immediately for users

### Validation Rules

**Preventing Bad Data**:
1. Setup → Object Manager → [Object] → Validation Rules
2. Click "New Rule"
3. Enter formula that must be TRUE for save to succeed
4. Example: `Amount <= 0` with error "Amount must be positive"
5. Save
6. Rule enforced on all creates/updates

### Automation (Flows)

**Creating Workflows**:
1. Setup → Flows
2. Click "New Flow"
3. Define trigger:
   - Object (e.g., Opportunity)
   - Event (After Create, After Update)
   - Condition (e.g., `Stage == 'Closed Won'`)
4. Add actions:
   - Create Task
   - Send Email
   - Update Field
5. Save and activate
6. Flow runs automatically when conditions met

---

## 6. Security & Permissions

### Profile System

**Understanding Profiles**:
- Every user has ONE profile
- Profile defines WHAT you can do (permissions)
- Examples: `system_admin`, `standard_user`

**Profile Permissions**:
- **Object Permissions**: Create, Read, Edit, Delete per object
- **Field Permissions**: Which fields you can see/edit
- **ViewAll/ModifyAll**: Bypass ownership rules



### Managing Users

**Creating Users** (Admin Only):
1. Setup → Users
2. Click "New User"
3. Enter:
   - Email (used for login)
   - Name
   - Profile (required)
   - Role (optional)
4. Save
5. User receives email with temp password

**Assigning Permissions**:
1. Setup → Profiles
2. Select profile
3. Configure:
   - Object permissions (CRUD)
   - Field permissions (Read/Edit)
4. Save
5. All users with this profile inherit permissions



---

## 7. Tips & Best Practices

### Data Quality
- ✅ Use validation rules to enforce data standards
- ✅ Make important fields required
- ✅ Use picklists instead of free text when possible
- ✅ Regular data cleanup and deduplication

### Performance
- ✅ Use filters to reduce dataset size
- ✅ Archive old records
- ✅ Limit formula complexity
- ✅ Use indexed fields for filtering

### Security
- ✅ Principle of least privilege - don't over-grant permissions
- ✅ Regular permission audits
- ✅ Use roles for hierarchy, not for permissions
- ✅ Strong passwords (12+ chars, mixed case, numbers, symbols)

### Customization
- ✅ Plan your data model before creating objects
- ✅ Use standard objects when possible (Account, Contact)
- ✅ Document custom fields and objects
- ✅ Test changes in sandbox environment
- ✅ Train users on new features

## 8. Business Processes

### Lead Conversion
1. Open a **Lead** record
2. Click the **Convert** action button
3. The system will automatically:
   - Create a new **Account** (Company)
   - Create a new **Contact** (Person)
   - Create a new **Opportunity** (Deal)
   - Update Lead status to "Converted"

### AI Assistant (Nexus AI)
- Click the **Nexus AI** button (🤖 icon) in the top navigation bar to open the assistant.
- **Capabilities**:
  - Answer questions about your data ("Show me high-value opportunities")
  - Navigate the system ("Go to Setup")
  - Perform actions ("Create a task to call John Doe")

---

## 9. Getting Help



### Documentation
- Main README: Project overview
- ARCHITECTURE.md: Technical architecture
- SECURITY.md: Security details
- This manual: User guide

### Support
- Contact your System Administrator for access and support issues.
- GitHub Issues: Bug reports and feature requests


