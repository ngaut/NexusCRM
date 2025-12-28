# Security Documentation

## Overview

NexusCRM implements enterprise-grade security features with Go backend and industry-standard authentication.

---

## Authentication

### JWT-Based Authentication
- **Token Type**: JSON Web Tokens (JWT)
- **Algorithm**: HS256
- **Expiration**: 24 hours
- **Token Storage**: Client-side localStorage (`nexus_auth_token`)
- **Token Transmission**: Bearer token in Authorization header

### Session Management
- Sessions stored in `_System_Session` table
- Session tracking includes:
  - Unique session ID (jti claim)
  - User ID and email
  - Creation timestamp
  - Expiration timestamp
- Logout invalidates the session

### Password Security

**Hashing**:
- Algorithm: **bcrypt**
- Rounds: **12** (configurable via `BCRYPT_ROUNDS`)
- Salt generated per-password
- Industry-standard security

**Requirements**:
- Minimum 8 characters
- At least one uppercase letter (A-Z)
- At least one lowercase letter (a-z)
- At least one number (0-9)
- At least one special character (!@#$%^&*(),.?":{}|<>_-+=[]\\/;'`~)

**Default Admin Account**:
- Email: `admin@test.com`
- Password: `Admin123!`
- ⚠️ **CHANGE THIS PASSWORD IMMEDIATELY AFTER FIRST LOGIN**

---

## Authorization

### Profile-Based Permissions (Required)

Every user must have a **Profile** that defines permissions:

**System Profiles**:
- `system_admin` - Full access to all resources
- `standard_user` - Limited access based on permissions

**Permission Types**:
- **Object-Level** (`_System_ObjectPerms`):
  - Create, Read, Edit, Delete
  - ViewAll, ModifyAll (bypass ownership rules)
- **Field-Level** (`_System_FieldPerms`):
  - Readable (can see field)
  - Editable (can modify field)

### Role Hierarchy (Optional)

**Salesforce-Inspired Design**:
- **Profile** = What you can do (permissions)
- **Role** = Whose data you can see (hierarchy)

**Role Features**:
- Tree structure (CEO > VP > Manager > Rep)
- Higher roles inherit data access from lower roles
- Not all users need roles (e.g., flat organizations)
- Computed in `PermissionService`

**Database Schema**:
```sql
_System_User:
  ProfileId varchar(255) NOT NULL  -- Required
 RoleId    varchar(255) NULL       -- Optional
```

### Row-Level Security (RLS)

**Enforced in QueryService**:
1. **Ownership**: Users see their own records
2. **Role Hierarchy**: Managers see subordinate records
3. **Sharing Rules**: Criteria-based access (`_System_SharingRule`)
4. **ViewAll/ModifyAll**: Profile overrides

---

## Network Security

### HTTPS/TLS
- **Backend → TiDB Cloud**: TLS 1.2+ encryption
- **Client → Backend**: HTTPS in production (configure reverse proxy)

### CORS Configuration
- Configurable allowed origins
- Credentials support enabled
- Proper preflight handling

---

## Database Security

### TiDB Cloud Connection
- **TLS Encryption**: All connections encrypted
- **Connection Pooling**: Secure connection management
- **Parameterized Queries**: SQL injection prevention
- **No Direct Frontend Access**: All database operations through authenticated backend

**Environment Variables** (.env):
```bash
TIDB_HOST=gateway01.us-west-2.prod.aws.tidbcloud.com
TIDB_PORT=4000
TIDB_USER=<your-username>
TIDB_PASSWORD=<your-password>
TIDB_DATABASE=nexuscrm
JWT_SECRET=<generate-with-openssl-rand-base64-32>
```

### SQL Injection Prevention
- **All queries parameterized**: No string concatenation
- **Whitelist validation**: Column names validated
- **Input sanitization**: User input escaped
- **Field permissions**: Only allowed fields queried

---

## Input Validation

### Email Validation
- RFC 5322 compliant
- Applied to authentication endpoints

### Request Validation
- JSON schema validation
- Type checking
- Size limits (prevent DoS)

### Field-Level Validation
- Configured in `_System_Validation` metadata
- Custom validation rules
- Client and server-side enforcement

---

## Security Best Practices

### Environment Variables

**Required in Production**:
```bash
TIDB_HOST=<your-tidb-host>
TIDB_USER=<your-username>
TIDB_PASSWORD=<your-password>
TIDB_DATABASE=nexuscrm
JWT_SECRET=<generate-with-openssl-rand-base64-32>
```

**Generate Strong JWT Secret**:
```bash
openssl rand -base64 32
```

### Configuration Checklist

1. ✅ **Never commit `.env` to version control**
   - Added to `.gitignore`
   - Use `.env.example` as template

2. ✅ **Rotate credentials if exposed**
   - Change database password
   - Generate new JWT secret
   - Invalidate all sessions

3. ✅ **Use environment-specific configs**
   - Development: Relaxed for debugging
   - Production: Strict enforcement

4. ✅ **Change default admin password**
   - On first login
   - Use password manager
   - 20+ characters recommended

---

## Security Architecture (Go Backend)

### Clean Architecture Security
```
Frontend → REST API (JWT Auth) → Application Services → Database
           ↓ Validation         ↓ Permission Check    ↓ TLS
           ✅ Secure             ✅ Secure              ✅ Secure
```

### Security Layers

**1. Interface Layer** (`internal/interfaces/rest/`):
- JWT validation middleware
- Input validation
- CORS enforcement

**2. Application Layer** (`internal/application/services/`):
- PermissionService checks
- Business logic validation

**3. Infrastructure Layer** (`internal/infrastructure/`):
- TLS database connections
- Secure credential management

---

## Audit & Monitoring

### Error Handling
- Detailed errors logged server-side only
- Generic error messages to clients
- No information leakage

---


