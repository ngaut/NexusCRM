# NexusCRM User Manual

Welcome to **NexusCRM**, a metadata-driven CRM platform for managing customer relationships and business processes.

---

## Quick Start

**Default Credentials:**
- Email: `admin@test.com`
- Password: `Admin123!`
- ⚠️ Change password after first login

**Running the Application:**
```bash
npm run dev:full    # Starts both backend and frontend
```
Open browser: `http://localhost:5173`

---

## Navigation

### Sidebar
- **Apps**: Switch between business areas
- **Tabs**: Object list views
- **Dashboards**: Analytics
- **Setup**: Admin configuration (admin only)

### App Launcher
Click the grid icon (top-left) to browse and switch between available applications.

### Global Search
Top navigation - search across all objects by name, email, or field values.

---

## Working with Data

### List Views
- Navigate to any object tab
- Filter: Click filter icon, add conditions
- Sort: Click column headers
- Kanban: Toggle view for picklist fields (drag/drop cards)

### Creating Records
1. Click "New" button
2. Fill required fields (red asterisk)
3. Click "Save"

### Editing Records
- **Inline**: Click field to edit directly
- **Full Edit**: Click "Edit" button, modify fields, save

### Deleting Records
Records move to Recycle Bin (recoverable).

---

## Dashboards

### Viewing
- Click "Dashboards" tab
- Select a dashboard
- View real-time metrics and charts

### Widget Types
- **Metrics**: Key numbers
- **Charts**: Bar, Pie, Line visualizations
- **Tables**: Record lists
- **Kanban**: Board view
- **SQL Analytics**: Advanced queries (admin only)

---

## Administrator Guide

NexusCRM is **metadata-driven** - customize without code.

### Object Manager

**Creating Objects:**
Setup → Object Manager → New Object
- Enter Label, API Name, Description
- New object appears in database and UI

**Adding Fields:**
Setup → Object Manager → [Object] → New Field
- Text, Number, Currency, Email, Phone, URL
- Picklist (dropdown), Checkbox
- Date, DateTime, Lookup, Formula

### Page Layouts
Setup → Object Manager → [Object] → Layouts
- Drag/drop sections and fields
- Changes appear immediately

### Validation Rules
Setup → Object Manager → [Object] → Validation Rules
- Define conditions that must be true to save
- Example: `Amount <= 0` → "Amount must be positive"

### Automation (Flows)
Setup → Flows → New Flow
- Trigger: Object + Event (After Create/Update) + Condition
- Actions: Create Task, Send Email, Update Field

---

## Security & Permissions

### Profiles
- Every user has ONE profile
- Profile defines permissions (CRUD per object)
- Field-level: Read/Edit per field

### Managing Users
Setup → Users → New User
- Email, Name, Profile (required), Role (optional)

### Assigning Permissions
Setup → Profiles → Select Profile
- Object permissions (Create, Read, Edit, Delete)
- Field permissions (Read, Edit)

---

## AI Assistant (Nexus AI)

Click the **Nexus AI** button in the top navigation bar.

**Capabilities:**
- Answer questions about your data
- Navigate the system
- Perform actions

---

## Best Practices

### Data Quality
- ✅ Use validation rules
- ✅ Make important fields required
- ✅ Use picklists over free text

### Security
- ✅ Principle of least privilege
- ✅ Strong passwords (8+ chars, mixed case, numbers, special chars)
- ✅ Use roles for hierarchy, not permissions

---

## Documentation

- [README.md](../README.md) - Project overview
- [ARCHITECTURE.md](./ARCHITECTURE.md) - Technical design
- [SECURITY.md](./SECURITY.md) - Security features
