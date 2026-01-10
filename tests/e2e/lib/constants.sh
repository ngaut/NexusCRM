#!/bin/bash
# tests/e2e/lib/constants.sh
# Standardized constants for E2E tests

# API Endpoints
export API_METADATA_OBJECTS="/api/metadata/objects"
export API_METADATA_FLOWS="/api/metadata/flows"
export API_DATA="/api/data"
export API_DATA_QUERY="/api/data/query"
export API_AUTH_LOGIN="/api/auth/login"
export API_AUTH_LOGOUT="/api/auth/logout"

# System Tables (Prefix: _system_)
export SYS_GROUP="_system_group"
export SYS_GROUP_MEMBER="_system_groupmember"
export SYS_SHARING_RULE="_system_sharingrule"
export SYS_USER="_system_user"
export SYS_ROLE="_system_role"
export SYS_PROFILE="_system_profile"
export SYS_APP="_system_app"

# Common Field Names
export FIELD_ID="id"
export FIELD_NAME="name"
export FIELD_LABEL="label"
export FIELD_TYPE="type"
export FIELD_EMAIL="email"
export FIELD_DESCRIPTION="description"
export FIELD_OWNER_ID="owner_id"
export FIELD_CREATED_BY_ID="created_by_id"
export FIELD_CREATED_DATE="created_date"
export FIELD_LAST_MODIFIED_DATE="last_modified_date"
export FIELD_IS_DELETED="is_deleted"
export FIELD_IS_ACTIVE="is_active"
export FIELD_OBJECT_API_NAME="object_api_name"
export FIELD_PLURAL_LABEL="plural_label"
export FIELD_IS_CUSTOM="is_custom"
export FIELD_SEARCHABLE="searchable"
export FIELD_STATUS="status"
export FIELD_ACCESS_LEVEL="access_level"
export FIELD_CRITERIA="criteria"
export FIELD_TRIGGER_OBJECT="trigger_object"
export FIELD_TRIGGER_TYPE="trigger_type"
export FIELD_TRIGGER_CONDITION="trigger_condition"
export FIELD_ACTION_TYPE="action_type"
export FIELD_ACTION_CONFIG="action_config"
export FIELD_FIELDS="fields"
export FIELD_SHARE_WITH_GROUP_ID="share_with_group_id"
export FIELD_GROUP_ID="group_id"
export FIELD_USER_ID="user_id"

# Common Constants
export FILTER_OP_EQ="="
export FILTER_OP_NEQ="!="
export FILTER_OP_GT=">"
export FILTER_OP_LT="<"
export FILTER_OP_LIKE="LIKE"

# Common Values
export VAL_GROUP_TYPE_REGULAR="Regular"
export VAL_GROUP_TYPE_QUEUE="Queue"
export VAL_STATUS_ACTIVE="Active"
export VAL_ACCESS_LEVEL_READ="Read"
export VAL_TRIGGER_TYPE_AFTER_CREATE="afterCreate"
export VAL_ACTION_TYPE_UPDATE_RECORD="updateRecord"
export VAL_FIELD_TYPE_EMAIL="Email"
export VAL_FIELD_TYPE_TEXT="Text"
