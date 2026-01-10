
import { COMMON_FIELDS } from './core/constants';

// --- Core Schema Metadata ---

import { FieldType } from './core/constants/SchemaDefinitions';
export { type FieldType } from './core/constants/SchemaDefinitions';

export interface FieldTypeInfo {
  name: string;
  label: string;
  description: string;
  icon: string;
  sqlType: string;
  isSearchable: boolean;
  isGroupable: boolean;
  isSummable: boolean;
  isVirtual: boolean;
  operators: string[];
  isPlugin: boolean;
}

import {
  SystemUser,
  SystemProfile,
  SystemRole,
  SystemGroup,
  SystemGroupMember,
  SystemPermissionSet,
  SystemPermissionSetAssignment,
  SystemSharingRule,
  SystemObjectPerms,
  SystemFieldPerms
} from './generated-schema';

export type SharingModel = 'Private' | 'PublicRead' | 'PublicReadWrite';

export interface FieldMetadata {
  api_name: string;
  label: string;
  type: FieldType;
  description?: string;
  required?: boolean;
  unique?: boolean; // Data Integrity: Enforce Uniqueness
  is_name_field?: boolean; // Display Identity: Used as the primary record label (replaces hardcoded 'Name')
  options?: string[]; // For Picklists
  reference_to?: string[]; // For Lookups. Array of object names.
  is_polymorphic?: boolean; // If true, can reference multiple object types.
  delete_rule?: 'Restrict' | 'Cascade' | 'SetNull'; // Referential Integrity
  is_system?: boolean;
  formula?: string; // JavaScript expression for Formula type
  return_type?: FieldType; // For Formula display formatting
  default_value?: string; // Initial value for new records
  help_text?: string; // Tooltip text for end users
  is_master_detail?: boolean; // Parent-Child relationship
  relationship_name?: string; // Name of the relationship for subqueries
  track_history?: boolean; // Audit: Track changes to this field

  // Validation Metadata
  min_value?: number;
  max_value?: number;
  min_length?: number;
  max_length?: number;
  regex?: string;
  regex_message?: string;
  validator?: string; // Reference to a registered validator (e.g. 'USZip')
  decimal_places?: number;
  display_format?: string;
  starting_number?: number;

  // Dependency Logic
  controlling_field?: string;
  picklist_dependency?: Record<string, string[]>; // Map<ControllerValue, AllowedValues[]>

  // Rollup Logic
  rollup_config?: {
    summary_object: string; // The child object (e.g. Opportunity)
    summary_field: string;  // The field on child to aggregate (e.g. Amount)
    calc_type: 'COUNT' | 'SUM' | 'MIN' | 'MAX' | 'AVG';
    filter?: string;       // Optional JS condition (e.g. "record.stage_name === 'Closed Won'")
  };
}

export interface ObjectMetadata {
  api_name: string;
  app_id?: string; // App scoping
  label: string;
  plural_label: string;
  icon: string;
  description?: string;
  is_system?: boolean; // Added is_system property
  is_custom?: boolean; // Added is_custom property
  theme_color?: string; // Visual Identity: e.g. 'blue', 'orange', '#FF5733'
  sharing_model: SharingModel; // Security: OWD
  enable_hierarchy_sharing?: boolean; // Security: Grant Access Using Hierarchies
  fields: FieldMetadata[];
  default_list_view?: 'List' | 'Kanban';
  kanban_group_by?: string;
  kanban_summary_field?: string; // Field to aggregate in Kanban headers (e.g. Amount)
  list_fields?: string[]; // Fields to display in List View (Default fallback)
  searchable?: boolean; // Can be searched via Global Search
  path_field?: string; // Field to use for Path component (must be Picklist)
}

// --- Permissions & Security ---

// Note: Generated types (SystemProfile, SystemRole, etc.) already include id aliases from codegen
export interface Profile extends SystemProfile { }

export interface Role extends SystemRole { }

// --- Groups & Queues ---

export interface Group extends SystemGroup { }

export interface GroupMember extends SystemGroupMember { }

// --- Permission Sets ---

export interface PermissionSet extends SystemPermissionSet { }

export interface PermissionSetAssignment extends SystemPermissionSetAssignment { }

// ObjectPermission and FieldPermission: system fields are optional for new records (before save)
export interface ObjectPermission extends Partial<SystemObjectPerms> {
  object_api_name: string; // Required field
  allow_read: boolean;
  allow_create: boolean;
  allow_edit: boolean;
  allow_delete: boolean;
  view_all: boolean;
  modify_all: boolean;
}

export interface FieldPermission extends Partial<SystemFieldPerms> {
  object_api_name: string; // Required field
  field_api_name: string; // Required field  
  readable: boolean;
  editable: boolean;
}

export interface SharingRule extends SystemSharingRule { }

export interface ProfileLayoutAssignment {
  id: string; // Composite key: profileId-objectApiName
  profile_id: string;
  object_api_name: string;
  layout_id: string;
}



export interface LayoutSection {
  id: string;
  label: string;
  columns: 1 | 2;
  fields: string[];
  visibility_condition?: string; // Formula string
}

export interface UserSession {
  id: string;
  name: string;
  email: string;
  profile_id: string;
  role_id?: string;
}

export interface User extends SystemUser {
  // SystemUser has 'username', 'first_name', 'last_name'.
  // 'name' is often a computed display name (e.g. first + last or username).
  // We keep it here to avoid breaking UI that relies on 'user.name'.
  name: string;
  // SystemUser has 'last_login_date'. Map 'last_login' to it if needed, or allow both.
  last_login?: string;
  // Password is not part of the standard API response for User but is needed for creation/updates
  password?: string;
}

// --- Layout & Experience Metadata ---

export type ActionType = 'Standard' | 'CreateRecord' | 'UpdateRecord' | 'Flow' | 'Custom' | 'Url' | 'modal' | 'link' | 'api' | 'flow';
export type SectionType = 'Fields' | 'Component';

export interface ActionConfig {
  name: string;
  label: string;
  type: ActionType;
  icon?: string;
  target_object?: string;
  config?: Record<string, unknown>; // Unified configuration (defaults, updates, url params)
  visibility_condition?: string; // Formula: "record.status !== 'Closed'"
  component?: string; // For Custom Action UI (e.g. 'EmailComposer')
}

export interface ActionMetadata {
  id: string;
  object_api_name: string;
  name: string;
  label: string;
  type: ActionType;
  icon: string;
  target_object?: string;
  config?: Record<string, unknown>;
}

export interface PageSection {
  id: string;
  label: string;
  type?: SectionType; // Defaults to 'Fields' if undefined
  component_name?: string; // Required if type === 'Component'
  component_config?: Record<string, unknown>; // Props passed to the component
  columns: 1 | 2;
  fields: string[];
  visibility_condition?: string; // Formula string
}

export interface RelatedListConfig {
  id: string;
  label: string;
  object_api_name: string;
  lookup_field: string;
  fields: string[];
}

export type LayoutType = 'Detail' | 'Edit' | 'Create' | 'List';

export interface PageLayout {
  id: string;
  object_api_name: string;
  layout_name: string; // User friendly name for the layout
  type: LayoutType;
  is_default?: boolean;
  compact_layout: string[]; // Fields to display in the header/highlights panel
  tabs?: string[]; // Ordered list of tabs: 'Details', 'Related', 'Feed'
  sections: PageSection[];
  related_lists: RelatedListConfig[];
  header_actions: ActionConfig[];
  quick_actions: ActionConfig[];
}

export interface ListView {
  [COMMON_FIELDS.ID]: string;
  id?: string; // Alias for [COMMON_FIELDS.ID]
  [COMMON_FIELDS.OBJECT_API_NAME]: string;
  [COMMON_FIELDS.LABEL]: string;
  filter_expr?: string;
  fields?: string[]; // Columns to display
}

// --- Business Logic Metadata ---

export interface TransformationTarget {
  target_object: string;
  required: boolean;
  field_mapping: Record<string, string>; // TargetField -> Formula Expression (e.g. "record.amount * 0.1")
}

export interface TransformationConfig {
  id: string;
  name: string;
  source_object: string;
  status_field?: string;   // Logic: Field to check/update (Default: 'Status')
  trigger_status?: string; // Logic: e.g. 'Converted'
  target_status?: string;  // Logic: Status to set on source after conversion
  targets: TransformationTarget[];
}

export interface PromptTemplate {
  id: string;
  template: string;
  description?: string;
  model?: string;
}

export interface SystemConfig {
  key_name: string;
  value: string;
  is_secret: boolean;
  description?: string;
}

// --- Dashboard & Analytics Metadata ---

export interface AnalyticsQuery {
  object_api_name: string;
  operation: 'count' | 'sum' | 'avg' | 'group_by';
  field?: string; // The field to sum/avg or group by
  group_by?: string;
  filter_expr?: string;
}

export interface ChartDataEntry {
  name: string;
  value: number;
  [key: string]: unknown;
}

// Forward declaration from FilterBar to avoid circular deps if needed, strict typing for now
export interface GlobalFilters {
  ownerId?: string;
  startDate?: string;
  endDate?: string;
}

export interface WidgetRendererProps {
  id: string;
  title: string;
  data: unknown;
  loading: boolean;
  config: WidgetConfig;
  isEditing?: boolean;
  isVisible?: boolean;
  onToggle?: () => void;
  globalFilters?: GlobalFilters;
  onConfigUpdate?: (newConfig: Partial<WidgetConfig>) => void;
}


export interface WidgetConfig {
  id: string;
  title: string;
  type: string; // dynamic type (was 'metric' | 'chart-bar' | ...)
  query?: AnalyticsQuery; // Optional now? Or just used for standard widgets
  config: Record<string, unknown>; // Flexible config for widget-specific settings (SQL, Markdown content, etc.)

  // React-Grid-Layout props
  x?: number;
  y?: number;
  w?: number;
  h?: number;
  icon?: string;
  color?: string;
  scope?: 'mine' | 'all';
}

export interface DashboardConfig {
  id: string;
  label: string;
  description?: string;
  widgets: WidgetConfig[];
}

// --- App & Navigation Configuration ---

export type NavigationItemType = 'object' | 'page' | 'web' | 'dashboard';

export interface NavigationItem {
  id: string;
  type: NavigationItemType;
  object_api_name?: string; // For 'object' type - the object to navigate to
  page_url?: string; // For 'web' type - external URL
  dashboard_id?: string; // For 'dashboard' type - dashboard to navigate to
  label: string;
  icon: string;
}

export type UtilityItemType = 'notes' | 'recent' | 'history' | 'custom';

export interface UtilityItem {
  id: string;
  type: UtilityItemType;
  label: string;
  icon: string;
  panel_width?: number; // Default 340
  panel_height?: number; // Default 480
}

export interface ThemeColors {
  brand: string;
  brand_light: string;
  brand_dark: string;
  secondary: string;
  success: string;
  warning: string;
  danger: string;
  background: string;
  surface: string;
  text: string;
  text_secondary: string;
  border: string;
  [key: string]: string; // Allow other colors
}

export interface Theme {
  id: string;
  name: string;
  colors: ThemeColors;
  density: string;
  logo_url?: string;
}

export interface AppConfig {
  id: string;
  label: string;
  description: string;
  icon: string;
  color: string;

  navigation_items?: NavigationItem[]; // New: Inline navigation items
  assigned_profiles?: string[]; // Profile IDs that can access this app (empty = all profiles)
  utility_items?: UtilityItem[]; // Utility bar items (Notes, Recent, History, etc.)
}



// --- Runtime Data & Interaction ---

export interface SObject {
  [COMMON_FIELDS.ID]?: string;
  id?: string; // Alias for COMMON_FIELDS.ID
  [COMMON_FIELDS.CREATED_DATE]?: string;
  created_date?: string; // Alias for COMMON_FIELDS.CREATED_DATE
  [COMMON_FIELDS.OWNER_ID]?: string;
  owner_id?: string; // Alias for COMMON_FIELDS.OWNER_ID
  [COMMON_FIELDS.CREATED_BY_ID]?: string;
  created_by_id?: string; // Alias for COMMON_FIELDS.CREATED_BY_ID
  [COMMON_FIELDS.LAST_MODIFIED_DATE]?: string;
  last_modified_date?: string; // Alias for COMMON_FIELDS.LAST_MODIFIED_DATE
  [COMMON_FIELDS.LAST_MODIFIED_BY_ID]?: string;
  last_modified_by_id?: string; // Alias for COMMON_FIELDS.LAST_MODIFIED_BY_ID
  [COMMON_FIELDS.IS_DELETED]?: boolean; // Soft Delete Flag
  is_deleted?: boolean; // Alias for COMMON_FIELDS.IS_DELETED
  [key: string]: unknown;
}

export interface SearchResult {
  object_label: string;
  object_api_name: string;
  icon: string;
  matches: SObject[];
}

export interface ChatMessage {
  id: string;
  role: 'user' | 'model';
  text: string;
  timestamp: number;
}

// Note: Flow and FlowStep types are in infrastructure/api/flows.ts

export interface ValidationRule {
  id: string;
  object_api_name: string;
  name: string;
  active: boolean;
  condition: string; // JavaScript expression, e.g. "record.amount < 0"
  error_message: string;
}

export interface ApprovalProcess {
  id: string;
  name: string;
  object_api_name: string; // Target Object
  description?: string;
  entry_criteria?: string;
  approver_type: string; // 'User' | 'Manager' | 'Self'
  approver_id?: string;
  is_active: boolean;
  created_date?: string;
  last_modified_date?: string;
}

export interface RecycleBinItem {
  [COMMON_FIELDS.ID]: string;
  id?: string; // Alias for [COMMON_FIELDS.ID]
  [COMMON_FIELDS.RECORD_ID]: string;
  [COMMON_FIELDS.OBJECT_API_NAME]: string;
  record_name: string;
  [COMMON_FIELDS.DELETED_BY]: string;
  [COMMON_FIELDS.DELETED_DATE]: string;
}

// Note: SystemLog is available in generated-schema.ts (SystemSystemLog)
// Note: RecentItem can be added back if needed
