/**
 * System Table Constants
 * 
 * Imports from shared JSON - single source of truth matching protobuf enums.
 * @see shared/constants/system.json
 * @see shared/proto/constants.proto (SystemTable enum)
 */

import systemConstants from '@shared/constants/system.json';

// Extract system tables from shared JSON
const { systemTables } = systemConstants;

// Table names - derived from shared JSON
export const TableObject = systemTables.OBJECT;
export const TableField = systemTables.FIELD;
export const TableRelationship = systemTables.RELATIONSHIP;
export const TableRecordType = systemTables.RECORD_TYPE;
export const TableLayout = systemTables.LAYOUT;
export const TableApp = systemTables.APP;
export const TableDashboard = systemTables.DASHBOARD;
export const TableListView = systemTables.LIST_VIEW;
export const TableSetupPage = systemTables.SETUP_PAGE;
export const TableProfile = systemTables.PROFILE;
export const TableUser = systemTables.USER;
export const TableObjectPerms = systemTables.OBJECT_PERMS;
export const TableFieldPerms = systemTables.FIELD_PERMS;
export const TableFlow = systemTables.FLOW;
export const TableAction = systemTables.ACTION;
export const TableValidation = systemTables.VALIDATION;
export const TableRecycleBin = systemTables.RECYCLE_BIN;
export const TableLog = systemTables.LOG;
export const TableRecent = systemTables.RECENT;
export const TableConfig = systemTables.CONFIG;
export const TableUIComponent = systemTables.UI_COMPONENT;

/**
 * Check if a table name is a system table
 */
export function isSystemTable(tableName: string): boolean {
    return tableName.startsWith('_System_');
}
