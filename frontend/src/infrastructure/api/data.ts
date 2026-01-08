import { apiClient } from './client';
import { API_ENDPOINTS } from './endpoints';
import { COMMON_FIELDS } from '../../core/constants';
import type { SObject, SearchResult, AnalyticsQuery, RecycleBinItem } from '../../types';

export interface QueryRequest {
  objectApiName: string;
  filterExpr?: string; // Formula expression for filtering
  sortField?: string;
  sortDirection?: string;
  limit?: number;
}

export const dataAPI = {
  /**
   * Query records with filter expression
   */
  async query<T = SObject>(request: QueryRequest): Promise<T[]> {
    const payload = {
      [COMMON_FIELDS.OBJECT_API_NAME]: request.objectApiName,
      filter_expr: request.filterExpr,
      sort_field: request.sortField,
      sort_direction: request.sortDirection,
      limit: request.limit,
    };
    const response = await apiClient.post<{ data: T[] }>(
      API_ENDPOINTS.DATA.QUERY,
      payload
    );
    return response.data;
  },

  /**
   * Global search across all objects
   */
  async search(term: string): Promise<SearchResult[]> {
    const response = await apiClient.post<{ data: SearchResult[] }>(
      API_ENDPOINTS.DATA.SEARCH,
      { term }
    );
    return response.data;
  },

  /**
   * Search within a single object
   */
  async searchSingleObject<T = SObject>(objectApiName: string, term: string): Promise<T[]> {
    const response = await apiClient.get<{ data: T[] }>(
      `${API_ENDPOINTS.DATA.SEARCH_OBJECT(objectApiName)}?term=${encodeURIComponent(term)}`
    );
    return response.data;
  },

  /**
   * Get a single record by ID
   */
  async getRecord<T = SObject>(objectApiName: string, id: string): Promise<T> {
    const response = await apiClient.get<{ data: T }>(
      API_ENDPOINTS.DATA.RECORD(objectApiName, id)
    );
    return response.data;
  },

  /**
   * Create a new record
   */
  async createRecord<T = SObject>(objectApiName: string, data: Partial<T>): Promise<T & { [COMMON_FIELDS.ID]: string }> {
    const response = await apiClient.post<{ data: T & { [COMMON_FIELDS.ID]: string } }>(
      API_ENDPOINTS.DATA.RECORDS(objectApiName),
      data
    );
    return response.data;
  },

  /**
   * Update an existing record
   */
  async updateRecord(objectApiName: string, id: string, updates: Partial<SObject>): Promise<void> {
    await apiClient.patch(
      API_ENDPOINTS.DATA.RECORD(objectApiName, id),
      updates
    );
  },

  /**
   * Delete a record (soft delete to recycle bin)
   */
  async deleteRecord(objectApiName: string, id: string): Promise<void> {
    await apiClient.delete(
      API_ENDPOINTS.DATA.RECORD(objectApiName, id)
    );
  },

  /**
   * Get recycle bin items
   */
  async getRecycleBinItems(scope: 'mine' | 'all' = 'mine'): Promise<RecycleBinItem[]> {
    const response = await apiClient.get<{ data: RecycleBinItem[] }>(`${API_ENDPOINTS.DATA.RECYCLE_BIN}?scope=${scope}`);
    return response.data;
  },

  /**
   * Restore record from recycle bin
   */
  async restoreRecord(id: string): Promise<void> {
    await apiClient.post(API_ENDPOINTS.DATA.RESTORE(id));
  },

  /**
   * Permanently delete record from recycle bin
   */
  async purgeRecord(id: string): Promise<void> {
    await apiClient.delete(API_ENDPOINTS.DATA.PURGE(id));
  },

  /**
   * Run analytics query
   */
  async runAnalytics(query: AnalyticsQuery): Promise<unknown> {
    const response = await apiClient.post<{ data: unknown }>(API_ENDPOINTS.DATA.ANALYTICS, query);
    return response.data;
  },
  /**
   * Calculate formula fields for a record
   */
  async calculate(objectApiName: string, record: SObject): Promise<SObject> {
    const response = await apiClient.post<{ data: SObject }>(
      API_ENDPOINTS.DATA.CALCULATE(objectApiName),
      record
    );
    return response.data;
  },

  /**
   * Execute a server-side action
   */
  async executeAction(actionId: string, payload: Record<string, unknown>): Promise<unknown> {
    const response = await apiClient.post<{ data: unknown }>(
      API_ENDPOINTS.ACTIONS.EXECUTE(actionId),
      payload
    );
    return response.data;
  }
};
