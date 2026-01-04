/**
 * Flow status constants
 */

export const FLOW_STATUS = {
    ACTIVE: 'Active',
    DRAFT: 'Draft',
    INACTIVE: 'Inactive',
} as const;

export type FlowStatus = typeof FLOW_STATUS[keyof typeof FLOW_STATUS];

// ============================================================================
// TRIGGER TYPES
// ============================================================================

export const TRIGGER_TYPE = {
    BEFORE_CREATE: 'BeforeCreate',
    AFTER_CREATE: 'AfterCreate',
    BEFORE_UPDATE: 'BeforeUpdate',
    AFTER_UPDATE: 'AfterUpdate',
    BEFORE_DELETE: 'BeforeDelete',
    AFTER_DELETE: 'AfterDelete',
} as const;

export type TriggerType = typeof TRIGGER_TYPE[keyof typeof TRIGGER_TYPE];

export const isBeforeTrigger = (t: TriggerType) => t.startsWith('Before');
export const isAfterTrigger = (t: TriggerType) => t.startsWith('After');

// ============================================================================
// ACTION TYPES
// ============================================================================

export const ACTION_TYPE = {
    CREATE_RECORD: 'CreateRecord',
    UPDATE_RECORD: 'UpdateRecord',
    DELETE_RECORD: 'DeleteRecord',
    SEND_EMAIL: 'SendEmail',
    CALL_WEBHOOK: 'CallWebhook',
    COMPOSITE: 'Composite',
    APPROVAL: 'Approval',
} as const;

export type ActionType = typeof ACTION_TYPE[keyof typeof ACTION_TYPE];

// ============================================================================
// EVENT TYPES
// ============================================================================

export const EVENT_TYPE = {
    RECORD_CREATED: 'RecordCreated',
    RECORD_UPDATED: 'RecordUpdated',
    RECORD_DELETED: 'RecordDeleted',
    OBJECT_CREATED: 'ObjectCreated',
    FIELD_CREATED: 'FieldCreated',
    SYSTEM_STARTUP: 'SystemStartup',
} as const;

export type EventType = typeof EVENT_TYPE[keyof typeof EVENT_TYPE];

