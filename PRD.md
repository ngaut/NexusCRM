# NexusCRM - Lessons Learned & Best Practices

## Technical Standards

### 1. Modal & Overlay Components
- **Use React Portal**: All modals must use `createPortal(content, document.body)`
- **Z-Index Standard**: Use `z-[100]` for all modal overlays
- **Reason**: Prevents stacking context issues where fixed headers appear above modals

```tsx
import { createPortal } from 'react-dom';

return createPortal(
    <div className="fixed inset-0 z-[100] ...">
        {/* modal content */}
    </div>,
    document.body
);
```

### 2. Schema & Constants
- **Single Source of Truth**: All schema definitions come from `system_tables.json`
- **Generated Code**: Use code generation for Go models and TypeScript types
- **No Magic Strings**: Always use generated constants for field/table names

### 3. Type Safety
- Replace `any` with `unknown` in error handling
- Use generated types from `generated-schema.ts`
- Avoid `as any` casts - refactor instead

### 4. UI Patterns
- Use toast notifications instead of `console.error`
- Feature flags for configurable behaviors (e.g., `ENABLE_AUTO_FILL_API_NAME`)
- Take screenshots before/after UI operations for verification

## Process Standards

### Server Management
```bash
# Backend
./backend/restart-server.sh

# Frontend (requires TTY)
npm run dev
```

### Code Changes
1. Audit affected files with grep before widespread changes
2. Fix base/reusable components first
3. Run `npm run lint` after each batch
4. Browser verification for UI changes

### Documentation
- Keep walkthroughs concise with before/after comparisons
- Embed verification screenshots
- Update docs in sync with code
