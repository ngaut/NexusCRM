/**
 * Platform Constants
 * 
 * This file provides typed constants loaded from the shared JSON definition.
 * These values are the single source of truth for trigger types, action types,
 * flow statuses, and event types across the entire platform.
 * 
 * @see /shared/constants/triggers.json
 * @see /shared/proto/constants.proto (Protobuf schema)
 */

import triggersData from '@shared/constants/triggers.json';

// Type definitions
export interface TriggerConstants {
    triggers: Record<string, string>;
    flowStatus: Record<string, string>;
    actionTypes: Record<string, string>;
    eventTypes: Record<string, string>;
}

// Load from shared JSON
const constants = triggersData as TriggerConstants;

// ============================================================================
// TRIGGER TYPES
// ============================================================================

export const TriggerType = {
    BEFORE_CREATE: constants.triggers.BEFORE_CREATE,
    AFTER_CREATE: constants.triggers.AFTER_CREATE,
    BEFORE_UPDATE: constants.triggers.BEFORE_UPDATE,
    AFTER_UPDATE: constants.triggers.AFTER_UPDATE,
    BEFORE_DELETE: constants.triggers.BEFORE_DELETE,
    AFTER_DELETE: constants.triggers.AFTER_DELETE,
} as const;

export type TriggerTypeValue = typeof TriggerType[keyof typeof TriggerType];

// ============================================================================
// FLOW STATUS
// ============================================================================

export const FlowStatus = {
    ACTIVE: constants.flowStatus.ACTIVE,
    INACTIVE: constants.flowStatus.INACTIVE,
    DRAFT: constants.flowStatus.DRAFT,
} as const;

export type FlowStatusValue = typeof FlowStatus[keyof typeof FlowStatus];

// ============================================================================
// ACTION TYPES
// ============================================================================

export const ActionType = {
    CREATE_RECORD: constants.actionTypes.CREATE_RECORD,
    UPDATE_RECORD: constants.actionTypes.UPDATE_RECORD,
    DELETE_RECORD: constants.actionTypes.DELETE_RECORD,
    SEND_EMAIL: constants.actionTypes.SEND_EMAIL,
    CALL_WEBHOOK: constants.actionTypes.CALL_WEBHOOK,
    COMPOSITE: constants.actionTypes.COMPOSITE,
} as const;

export type ActionTypeValue = typeof ActionType[keyof typeof ActionType];

// ============================================================================
// EVENT TYPES
// ============================================================================

export const EventType = {
    RECORD_BEFORE_CREATE: constants.eventTypes.RECORD_BEFORE_CREATE,
    RECORD_AFTER_CREATE: constants.eventTypes.RECORD_AFTER_CREATE,
    RECORD_CREATED: constants.eventTypes.RECORD_CREATED,
    RECORD_BEFORE_UPDATE: constants.eventTypes.RECORD_BEFORE_UPDATE,
    RECORD_AFTER_UPDATE: constants.eventTypes.RECORD_AFTER_UPDATE,
    RECORD_UPDATED: constants.eventTypes.RECORD_UPDATED,
    RECORD_BEFORE_DELETE: constants.eventTypes.RECORD_BEFORE_DELETE,
    RECORD_AFTER_DELETE: constants.eventTypes.RECORD_AFTER_DELETE,
    RECORD_DELETED: constants.eventTypes.RECORD_DELETED,
    OBJECT_CREATED: constants.eventTypes.OBJECT_CREATED,
    FIELD_CREATED: constants.eventTypes.FIELD_CREATED,
    SYSTEM_STARTUP: constants.eventTypes.SYSTEM_STARTUP,
} as const;

export type EventTypeValue = typeof EventType[keyof typeof EventType];

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

/**
 * Get all trigger types as an array of [key, value] pairs
 */
export function getAllTriggers(): [string, string][] {
    return Object.entries(constants.triggers);
}

/**
 * Get all action types as an array of [key, value] pairs
 */
export function getAllActionTypes(): [string, string][] {
    return Object.entries(constants.actionTypes);
}

/**
 * Get all flow statuses as an array of [key, value] pairs
 */
export function getAllFlowStatuses(): [string, string][] {
    return Object.entries(constants.flowStatus);
}

/**
 * Check if a value is a valid trigger type
 */
export function isValidTriggerType(value: string): value is TriggerTypeValue {
    return Object.values(TriggerType).includes(value as TriggerTypeValue);
}

/**
 * Check if a value is a valid action type
 */
export function isValidActionType(value: string): value is ActionTypeValue {
    return Object.values(ActionType).includes(value as ActionTypeValue);
}

/**
 * Check if a trigger is a "before" trigger (runs before persistence)
 */
export function isBeforeTrigger(triggerType: string): boolean {
    return triggerType === TriggerType.BEFORE_CREATE ||
        triggerType === TriggerType.BEFORE_UPDATE ||
        triggerType === TriggerType.BEFORE_DELETE;
}

/**
 * Check if a trigger is an "after" trigger (runs after persistence)
 */
export function isAfterTrigger(triggerType: string): boolean {
    return triggerType === TriggerType.AFTER_CREATE ||
        triggerType === TriggerType.AFTER_UPDATE ||
        triggerType === TriggerType.AFTER_DELETE;
}
