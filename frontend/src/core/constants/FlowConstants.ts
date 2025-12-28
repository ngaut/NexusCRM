/**
 * Flow status constants
 */

export const FLOW_STATUS = {
    ACTIVE: 'Active',
    DRAFT: 'Draft',
    INACTIVE: 'Inactive',
} as const;

export type FlowStatus = typeof FLOW_STATUS[keyof typeof FLOW_STATUS];
