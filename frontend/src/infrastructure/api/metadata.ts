import { api } from './client';
import type { ObjectMetadata, FieldMetadata, PageLayout, AppConfig, DashboardConfig } from '../../types';

export const metadataAPI = {
  // Schema operations
  getSchemas: () => api.get<{ schemas: ObjectMetadata[] }>(`/api/metadata/schemas?t=${Date.now()}`),
  getSchema: (api_name: string) => api.get<{ schema: ObjectMetadata }>(`/api/metadata/schemas/${api_name}?t=${Date.now()}`),
  createSchema: (schema: Partial<ObjectMetadata>) => api.post('/api/metadata/schemas', schema),
  updateSchema: (api_name: string, updates: Partial<ObjectMetadata>) => api.patch<{ message: string; schema: ObjectMetadata }>(`/api/metadata/schemas/${api_name}`, updates),
  deleteSchema: (api_name: string) => api.delete<{ message: string }>(`/api/metadata/schemas/${api_name}`),

  // Field operations
  createField: (objectApiName: string, field: Partial<FieldMetadata>) =>
    api.post(`/api/metadata/schemas/${objectApiName}/fields`, field),
  updateField: (objectApiName: string, fieldApiName: string, updates: Partial<FieldMetadata>) =>
    api.patch(`/api/metadata/schemas/${objectApiName}/fields/${fieldApiName}`, updates),
  deleteField: (objectApiName: string, fieldApiName: string) =>
    api.delete(`/api/metadata/schemas/${objectApiName}/fields/${fieldApiName}`),

  // Layout operations
  getLayout: (objectApiName: string) => api.get<{ layout: PageLayout }>(`/api/metadata/layouts/${objectApiName}`),
  saveLayout: (layout: PageLayout) => api.post('/api/metadata/layouts', layout),
  deleteLayout: (layoutId: string) => api.delete(`/api/metadata/layouts/${layoutId}`),
  assignLayoutToProfile: (profileId: string, objectApiName: string, layoutId: string) =>
    api.post('/api/metadata/layouts/assign', { profile_id: profileId, object_api_name: objectApiName, layout_id: layoutId }),

  // Action operations
  getActions: (objectApiName: string) => api.get<{ actions: import('../../types').ActionMetadata[] }>(`/api/metadata/actions/${objectApiName}`),


  // App operations
  getApps: () => api.get<{ apps: AppConfig[] }>('/api/metadata/apps'),
  createApp: (app: AppConfig) => api.post<{ message: string; app: AppConfig }>('/api/metadata/apps', app),
  updateApp: (appId: string, updates: Partial<AppConfig>) => api.patch<{ message: string; app: AppConfig }>(`/api/metadata/apps/${appId}`, updates),
  deleteApp: (appId: string) => api.delete<{ message: string }>(`/api/metadata/apps/${appId}`),

  // Dashboard operations
  getDashboards: () => api.get<{ dashboards: DashboardConfig[] }>('/api/metadata/dashboards'),
  getDashboard: (id: string) => api.get<{ dashboard: DashboardConfig }>(`/api/metadata/dashboards/${id}`),
  createDashboard: (dashboard: DashboardConfig) => api.post<{ message: string; dashboard: DashboardConfig }>('/api/metadata/dashboards', dashboard),
  updateDashboard: (id: string, updates: Partial<DashboardConfig>) => api.patch<{ message: string; dashboard: DashboardConfig }>(`/api/metadata/dashboards/${id}`, updates),
  deleteDashboard: async (id: string) => {
    return api.delete<{ message: string }>(`/api/metadata/dashboards/${id}`);
  },

  // Validation Rules
  getValidationRules: (objectApiName: string) => api.get<{ rules: import('../../types').ValidationRule[] }>(`/api/metadata/validation-rules?objectApiName=${objectApiName}`),
  createValidationRule: (rule: Partial<import('../../types').ValidationRule>) => api.post<{ message: string; rule: import('../../types').ValidationRule }>('/api/metadata/validation-rules', rule),
  updateValidationRule: (id: string, updates: Partial<import('../../types').ValidationRule>) => api.patch<{ message: string }>(`/api/metadata/validation-rules/${id}`, updates),
  deleteValidationRule: (id: string) => api.delete<{ message: string }>(`/api/metadata/validation-rules/${id}`),

  // Field Types
  getFieldTypes: () => api.get<{ fieldTypes: import('../../types').FieldTypeInfo[] }>('/api/metadata/fieldtypes'),

  // List View operations
  getListViews: (objectApiName: string) => api.get<{ views: import('../../types').ListView[] }>(`/api/metadata/listviews?objectApiName=${objectApiName}`),
  createListView: (view: Partial<import('../../types').ListView>) => api.post<{ view: import('../../types').ListView; message: string }>('/api/metadata/listviews', view),
  updateListView: (id: string, updates: Partial<import('../../types').ListView>) => api.patch<{ view: import('../../types').ListView; message: string }>(`/api/metadata/listviews/${id}`, updates),
  deleteListView: (id: string) => api.delete<{ message: string }>(`/api/metadata/listviews/${id}`),

  // Theme
  getActiveTheme: () => api.get<{ theme: import('../../types').Theme }>('/api/metadata/themes/active').then(res => res.theme),
};
