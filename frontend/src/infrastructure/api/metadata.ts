import { api } from './client';
import { API_ENDPOINTS } from './endpoints';
import { COMMON_FIELDS } from '../../core/constants';
import type { ObjectMetadata, FieldMetadata, PageLayout, AppConfig, DashboardConfig } from '../../types';

export const metadataAPI = {
  // Schema operations
  getSchemas: () => api.get<{ data: ObjectMetadata[] }>(`${API_ENDPOINTS.METADATA.OBJECTS}?t=${Date.now()}`).then(r => ({ schemas: r.data })),
  getSchema: (api_name: string) => api.get<{ data: ObjectMetadata }>(`${API_ENDPOINTS.METADATA.OBJECTS}/${api_name}?t=${Date.now()}`).then(r => ({ schema: r.data })),
  createSchema: (schema: Partial<ObjectMetadata>) => api.post<{ data: ObjectMetadata; message: string }>(API_ENDPOINTS.METADATA.OBJECTS, schema).then(r => ({ schema: r.data, message: r.message })),
  updateSchema: (api_name: string, updates: Partial<ObjectMetadata>) => api.patch<{ message: string; data: ObjectMetadata }>(`${API_ENDPOINTS.METADATA.OBJECTS}/${api_name}`, updates).then(r => ({ schema: r.data, message: r.message })),
  deleteSchema: (api_name: string) => api.delete<{ message: string }>(`${API_ENDPOINTS.METADATA.OBJECTS}/${api_name}`),

  // Field operations
  createField: (objectApiName: string, field: Partial<FieldMetadata>) =>
    api.post(API_ENDPOINTS.METADATA.FIELDS(objectApiName), field),
  updateField: (objectApiName: string, fieldApiName: string, updates: Partial<FieldMetadata>) =>
    api.patch(API_ENDPOINTS.METADATA.FIELD(objectApiName, fieldApiName), updates),
  deleteField: (objectApiName: string, fieldApiName: string) =>
    api.delete(API_ENDPOINTS.METADATA.FIELD(objectApiName, fieldApiName)),

  // Layout operations
  getLayout: (objectApiName: string) => api.get<{ data: PageLayout }>(API_ENDPOINTS.METADATA.LAYOUT(objectApiName)).then(r => ({ layout: r.data })),
  saveLayout: (layout: PageLayout) => api.post(API_ENDPOINTS.METADATA.LAYOUTS, layout),
  deleteLayout: (layoutId: string) => api.delete(API_ENDPOINTS.METADATA.LAYOUT_ID(layoutId)),
  assignLayoutToProfile: (profileId: string, objectApiName: string, layoutId: string) =>
    api.post(API_ENDPOINTS.METADATA.LAYOUT_ASSIGN, { [COMMON_FIELDS.PROFILE_ID]: profileId, [COMMON_FIELDS.OBJECT_API_NAME]: objectApiName, layout_id: layoutId }),

  // Action operations
  getActions: (objectApiName: string) => api.get<{ data: import('../../types').ActionMetadata[] }>(API_ENDPOINTS.METADATA.ACTIONS(objectApiName)).then(r => ({ actions: r.data || [] })),


  // App operations
  getApps: () => api.get<{ data: AppConfig[] }>(API_ENDPOINTS.METADATA.APPS).then(r => ({ apps: r.data })),
  createApp: (app: AppConfig) => api.post<{ message: string; app: AppConfig }>(API_ENDPOINTS.METADATA.APPS, app),
  updateApp: (appId: string, updates: Partial<AppConfig>) => api.patch<{ message: string; app: AppConfig }>(`${API_ENDPOINTS.METADATA.APPS}/${appId}`, updates),
  deleteApp: (appId: string) => api.delete<{ message: string }>(`${API_ENDPOINTS.METADATA.APPS}/${appId}`),

  // Dashboard operations
  getDashboards: () => api.get<{ data: DashboardConfig[] }>(API_ENDPOINTS.METADATA.DASHBOARDS).then(r => ({ dashboards: r.data })),
  getDashboard: (id: string) => api.get<{ data: DashboardConfig }>(API_ENDPOINTS.METADATA.DASHBOARD(id)).then(r => ({ dashboard: r.data })),
  createDashboard: (dashboard: DashboardConfig) => api.post<{ message: string; data: DashboardConfig }>(API_ENDPOINTS.METADATA.DASHBOARDS, dashboard).then(r => ({ dashboard: r.data, message: r.message })),
  updateDashboard: (id: string, updates: Partial<DashboardConfig>) => api.patch<{ message: string; data: DashboardConfig }>(API_ENDPOINTS.METADATA.DASHBOARD(id), updates).then(r => ({ dashboard: r.data, message: r.message })),
  deleteDashboard: async (id: string) => {
    return api.delete<{ message: string }>(API_ENDPOINTS.METADATA.DASHBOARD(id));
  },

  // Validation Rules
  getValidationRules: (objectApiName: string) => api.get<{ data: import('../../types').ValidationRule[] }>(`${API_ENDPOINTS.METADATA.VALIDATION_RULES}?objectApiName=${objectApiName}`).then(r => ({ rules: r.data })),
  createValidationRule: (rule: Partial<import('../../types').ValidationRule>) => api.post<{ message: string; data: import('../../types').ValidationRule }>(API_ENDPOINTS.METADATA.VALIDATION_RULES, rule).then(r => ({ rule: r.data, message: r.message })),
  updateValidationRule: (id: string, updates: Partial<import('../../types').ValidationRule>) => api.patch<{ message: string }>(API_ENDPOINTS.METADATA.VALIDATION_RULE(id), updates),
  deleteValidationRule: (id: string) => api.delete<{ message: string }>(API_ENDPOINTS.METADATA.VALIDATION_RULE(id)),

  // Field Types
  getFieldTypes: () => api.get<{ data: import('../../types').FieldTypeInfo[] }>(API_ENDPOINTS.METADATA.FIELD_TYPES).then(r => ({ fieldTypes: r.data })),

  // List View operations
  getListViews: (objectApiName: string) => api.get<{ data: import('../../types').ListView[] }>(`${API_ENDPOINTS.METADATA.LIST_VIEWS}?objectApiName=${objectApiName}`).then(r => ({ views: r.data })),
  createListView: (view: Partial<import('../../types').ListView>) => api.post<{ data: import('../../types').ListView; message: string }>(API_ENDPOINTS.METADATA.LIST_VIEWS, view).then(r => ({ view: r.data, message: r.message })),
  updateListView: (id: string, updates: Partial<import('../../types').ListView>) => api.patch<{ data: import('../../types').ListView; message: string }>(API_ENDPOINTS.METADATA.LIST_VIEW(id), updates).then(r => ({ view: r.data, message: r.message })),
  deleteListView: (id: string) => api.delete<{ message: string }>(API_ENDPOINTS.METADATA.LIST_VIEW(id)),

  // Theme
  getActiveTheme: () => api.get<{ data: import('../../types').Theme }>(API_ENDPOINTS.METADATA.THEMES).then(res => res.data),
};
