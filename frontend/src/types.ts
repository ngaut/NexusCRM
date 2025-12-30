
import { COMMON_FIELDS } from './core/constants';

// --- Core Schema Metadata ---

export type FieldType = 'Text' | 'Number' | 'Currency' | 'Date' | 'DateTime' | 'Picklist' | 'Email' | 'Phone' | 'TextArea' | 'LongTextArea' | 'RichText' | 'Lookup' | 'Url' | 'Boolean' | 'Formula' | 'Percent' | 'RollupSummary' | 'JSON' | 'Password' | string;

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

export interface Profile {
  id: string;
  name: string;
  description?: string;
  is_system?: boolean; // If true, cannot be deleted (e.g. System Admin)
}

export interface Role {
  id: string;
  name: string;
  parent_role_id?: string;
}

// --- Groups & Queues ---

export interface Group {
  id: string;
  name: string;
  label: string;
  type: 'Queue' | 'Regular';
  email?: string;
  created_date?: string;
}

export interface GroupMember {
  id: string;
  group_id: string;
  user_id: string;
  created_date?: string;
}

// --- Permission Sets ---

export interface PermissionSet {
  id: string;
  name: string;
  label: string;
  description?: string;
  is_active: boolean;
  created_date?: string;
}

export interface PermissionSetAssignment {
  id: string;
  assignee_id: string;
  permission_set_id: string;
  created_date?: string;
}

export interface ObjectPermission {
  profile_id?: string;
  permission_set_id?: string;
  object_api_name: string;
  allow_read: boolean;
  allow_create: boolean;
  allow_edit: boolean;
  allow_delete: boolean;
  view_all: boolean;
  modify_all: boolean;
}

export interface FieldPermission {
  profile_id?: string;
  permission_set_id?: string;
  object_api_name: string;
  field_api_name: string;
  readable: boolean;
  editable: boolean;
}

export interface SharingRule {
  id: string;
  object_api_name: string;
  name: string;
  criteria: string; // SQL/JS condition e.g. "Industry = 'Technology'"
  access_level: 'Read' | 'Edit';
  share_with_role_id: string; // The Role that gains access
}

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

export interface User {
  id: string;
  name: string;
  email: string;
  profile_id: string;
  role_id?: string;
  is_active: boolean;
  last_login?: string;
  created_date?: string;
  password?: string; // Optional for updates
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
  [COMMON_FIELDS.OBJECT_API_NAME]: string;
  [COMMON_FIELDS.LABEL]: string;
  [COMMON_FIELDS.FILTERS]?: string; // New formula filter
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
  filterExpr?: string;
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
  sql?: string; // For SQL Widgets

  // React-Grid-Layout props
  x?: number;
  y?: number;
  w?: number;
  h?: number;
  icon?: string;
  color?: string;
  content?: string; // For Text/Markdown widgets
  imageUrl?: string; // For Image widgets
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
  [COMMON_FIELDS.CREATED_DATE]?: string;
  [COMMON_FIELDS.OWNER_ID]?: string;
  [COMMON_FIELDS.CREATED_BY_ID]?: string;
  [COMMON_FIELDS.LAST_MODIFIED_DATE]?: string;
  [COMMON_FIELDS.LAST_MODIFIED_BY_ID]?: string;
  [COMMON_FIELDS.IS_DELETED]?: boolean; // Soft Delete Flag
  [key: string]: any;
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

// --- Platform Features ---

import { FlowStatus } from './core/constants/FlowConstants';

export interface Flow {
  id: string;
  name: string;
  description: string;
  trigger_object: string;
  trigger_condition: string;
  action_type: string;
  action_config: Record<string, unknown>; // Structured configuration for the action
  status: FlowStatus;
  last_modified: string;
}

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
  [COMMON_FIELDS.RECORD_ID]: string;
  [COMMON_FIELDS.OBJECT_API_NAME]: string;
  record_name: string;
  [COMMON_FIELDS.DELETED_BY]: string;
  [COMMON_FIELDS.DELETED_DATE]: string;
}

export interface SystemLog {
  id: string;
  timestamp: string;
  level: 'INFO' | 'WARN' | 'ERROR';
  source: string;
  message: string;
  details?: string;
}

export interface RecentItem {
  id: string;
  user_id: string;
  object_api_name: string;
  record_id: string;
  record_name: string;
  timestamp: string;
}
