/**
 * Constants
 * 
 * Re-exports all platform constants from the shared JSON definitions.
 * This ensures a single source of truth across frontend and backend.
 */

export {
    TriggerType,
    FlowStatus,
    ActionType,
    EventType,
    getAllTriggers,
    getAllActionTypes,
    getAllFlowStatuses,
    isValidTriggerType,
    isValidActionType,
    isBeforeTrigger,
    isAfterTrigger,
} from './triggers';

export type {
    TriggerTypeValue,
    FlowStatusValue,
    ActionTypeValue,
    EventTypeValue,
} from './triggers';

export {
    FieldType,
    getFieldTypeMetadata,
    getAllFieldTypes,

    isFieldTypeSearchable,
    isFieldTypeGroupable,
    isFieldTypeSummable,
    isValidFieldType,
    getFieldTypeIcon,
    getFieldTypeLabel,
} from './fieldTypes';

export type {
    FieldTypeName,
    FieldTypeValue,
    FieldTypeMetadata,
} from './fieldTypes';

// Field name constants
export * from './fields';

// Table name constants
export * from './tables';

// API and configuration constants
export * from './api';
