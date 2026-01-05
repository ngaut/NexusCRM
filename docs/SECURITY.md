# Security Documentation

## Authentication

### JWT-Based
- **Algorithm**: HS256
- **Expiration**: 24 hours
- **Storage**: Client-side localStorage (`nexus_auth_token`)
- **Transmission**: Bearer token in Authorization header

### Password Security
- **Hashing**: bcrypt (12 rounds)
- **Requirements**: 8+ chars, uppercase, lowercase, number, special char
- **Default Admin**: `admin@test.com` / `Admin123!` (change immediately)

### Sessions
- Stored in `_System_Session` table
- Includes: session ID, user ID, expiration
- Logout invalidates session

---

## Authorization

### Profile-Based Permissions
Every user must have a **Profile**:
- `system_admin` - Full access
- `standard_user` - Limited access

**Permission Types**:
- **Object-Level**: Create, Read, Edit, Delete, ViewAll, ModifyAll
- **Field-Level**: Readable, Editable

### Role Hierarchy (Optional)
- **Profile** = What you can do (permissions)
- **Role** = Whose data you can see (hierarchy)
- Not all users need roles

### Row-Level Security (RLS)
Enforced in QueryService:
1. **Ownership**: Users see their own records
2. **Role Hierarchy**: Managers see subordinate records
3. **Sharing Rules**: Criteria-based access
4. **ViewAll/ModifyAll**: Profile overrides

---

## Database Security

### TiDB Cloud
- TLS encryption on all connections
- Parameterized queries (SQL injection prevention)
- No direct frontend database access

### Environment Variables
```bash
TIDB_HOST=<your-tidb-host>
TIDB_USER=<username>
TIDB_PASSWORD=<password>
TIDB_DATABASE=nexuscrm
JWT_SECRET=<openssl rand -base64 32>
```

---

## Best Practices

1. ✅ Never commit `.env` to version control
2. ✅ Rotate credentials if exposed
3. ✅ Change default admin password
4. ✅ Use strong passwords (12+ chars)
5. ✅ Principle of least privilege
