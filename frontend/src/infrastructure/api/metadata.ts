import { api } from './client';
import { API_ENDPOINTS } from './endpoints';
import { COMMON_FIELDS } from '../../core/constants';
import type { ObjectMetadata, FieldMetadata, PageLayout, AppConfig, DashboardConfig } from '../../types';

export const metadataAPI = {
  // Schema operations
  getSchemas: () => api.get<{ schemas: ObjectMetadata[] }>(`${API_ENDPOINTS.METADATA.OBJECTS}?t=${Date.now()}`),
  getSchema: (api_name: string) => api.get<{ schema: ObjectMetadata }>(`${API_ENDPOINTS.METADATA.OBJECTS}/${api_name}?t=${Date.now()}`),
  createSchema: (schema: Partial<ObjectMetadata>) => api.post(API_ENDPOINTS.METADATA.OBJECTS, schema),
  updateSchema: (api_name: string, updates: Partial<ObjectMetadata>) => api.patch<{ message: string; schema: ObjectMetadata }>(`${API_ENDPOINTS.METADATA.OBJECTS}/${api_name}`, updates),
  deleteSchema: (api_name: string) => api.delete<{ message: string }>(`${API_ENDPOINTS.METADATA.OBJECTS}/${api_name}`),

  // Field operations
  createField: (objectApiName: string, field: Partial<FieldMetadata>) =>
    api.post(API_ENDPOINTS.METADATA.FIELDS(objectApiName), field),
  updateField: (objectApiName: string, fieldApiName: string, updates: Partial<FieldMetadata>) =>
    api.patch(API_ENDPOINTS.METADATA.FIELD(objectApiName, fieldApiName), updates),
  deleteField: (objectApiName: string, fieldApiName: string) =>
    api.delete(API_ENDPOINTS.METADATA.FIELD(objectApiName, fieldApiName)),

  // Layout operations
  getLayout: (objectApiName: string) => api.get<{ layout: PageLayout }>(API_ENDPOINTS.METADATA.LAYOUT(objectApiName)),
  saveLayout: (layout: PageLayout) => api.post(API_ENDPOINTS.METADATA.LAYOUTS, layout),
  deleteLayout: (layoutId: string) => api.delete(API_ENDPOINTS.METADATA.LAYOUT_ID(layoutId)),
  assignLayoutToProfile: (profileId: string, objectApiName: string, layoutId: string) =>
    api.post(API_ENDPOINTS.METADATA.LAYOUT_ASSIGN, { [COMMON_FIELDS.PROFILE_ID]: profileId, [COMMON_FIELDS.OBJECT_API_NAME]: objectApiName, layout_id: layoutId }),

  // Action operations
  getActions: (objectApiName: string) => api.get<{ actions: import('../../types').ActionMetadata[] }>(API_ENDPOINTS.METADATA.ACTIONS(objectApiName)),


  // App operations
  getApps: () => api.get<{ apps: AppConfig[] }>(API_ENDPOINTS.METADATA.APPS),
  createApp: (app: AppConfig) => api.post<{ message: string; app: AppConfig }>(API_ENDPOINTS.METADATA.APPS, app),
  updateApp: (appId: string, updates: Partial<AppConfig>) => api.patch<{ message: string; app: AppConfig }>(`${API_ENDPOINTS.METADATA.APPS}/${appId}`, updates),
  deleteApp: (appId: string) => api.delete<{ message: string }>(`${API_ENDPOINTS.METADATA.APPS}/${appId}`),

  // Dashboard operations
  getDashboards: () => api.get<{ dashboards: DashboardConfig[] }>(API_ENDPOINTS.METADATA.DASHBOARDS),
  getDashboard: (id: string) => api.get<{ dashboard: DashboardConfig }>(API_ENDPOINTS.METADATA.DASHBOARD(id)),
  createDashboard: (dashboard: DashboardConfig) => api.post<{ message: string; dashboard: DashboardConfig }>(API_ENDPOINTS.METADATA.DASHBOARDS, dashboard),
  updateDashboard: (id: string, updates: Partial<DashboardConfig>) => api.patch<{ message: string; dashboard: DashboardConfig }>(API_ENDPOINTS.METADATA.DASHBOARD(id), updates),
  deleteDashboard: async (id: string) => {
    return api.delete<{ message: string }>(API_ENDPOINTS.METADATA.DASHBOARD(id));
  },

  // Validation Rules
  getValidationRules: (objectApiName: string) => api.get<{ rules: import('../../types').ValidationRule[] }>(`${API_ENDPOINTS.METADATA.VALIDATION_RULES}?objectApiName=${objectApiName}`),
  createValidationRule: (rule: Partial<import('../../types').ValidationRule>) => api.post<{ message: string; rule: import('../../types').ValidationRule }>(API_ENDPOINTS.METADATA.VALIDATION_RULES, rule),
  updateValidationRule: (id: string, updates: Partial<import('../../types').ValidationRule>) => api.patch<{ message: string }>(API_ENDPOINTS.METADATA.VALIDATION_RULE(id), updates),
  deleteValidationRule: (id: string) => api.delete<{ message: string }>(API_ENDPOINTS.METADATA.VALIDATION_RULE(id)),

  // Field Types
  getFieldTypes: () => api.get<{ fieldTypes: import('../../types').FieldTypeInfo[] }>(API_ENDPOINTS.METADATA.FIELD_TYPES),

  // List View operations
  getListViews: (objectApiName: string) => api.get<{ views: import('../../types').ListView[] }>(`${API_ENDPOINTS.METADATA.LIST_VIEWS}?objectApiName=${objectApiName}`),
  createListView: (view: Partial<import('../../types').ListView>) => api.post<{ view: import('../../types').ListView; message: string }>(API_ENDPOINTS.METADATA.LIST_VIEWS, view),
  updateListView: (id: string, updates: Partial<import('../../types').ListView>) => api.patch<{ view: import('../../types').ListView; message: string }>(API_ENDPOINTS.METADATA.LIST_VIEW(id), updates),
  deleteListView: (id: string) => api.delete<{ message: string }>(API_ENDPOINTS.METADATA.LIST_VIEW(id)),

  // Theme
  getActiveTheme: () => api.get<{ theme: import('../../types').Theme }>(API_ENDPOINTS.METADATA.THEMES).then(res => res.theme),
};
