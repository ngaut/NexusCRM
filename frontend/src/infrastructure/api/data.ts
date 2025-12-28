import { apiClient } from './client';
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
      object_api_name: request.objectApiName,
      filter_expr: request.filterExpr,
      sort_field: request.sortField,
      sort_direction: request.sortDirection,
      limit: request.limit,
    };
    const response = await apiClient.post<{ records: T[] }>(
      '/api/data/query',
      payload
    );
    return response.records;
  },

  /**
   * Global search across all objects
   */
  async search(term: string): Promise<SearchResult[]> {
    const response = await apiClient.post<{ results: SearchResult[] }>(
      '/api/data/search',
      { term }
    );
    return response.results;
  },

  /**
   * Search within a single object
   */
  async searchSingleObject<T = SObject>(objectApiName: string, term: string): Promise<T[]> {
    const response = await apiClient.get<{ records: T[] }>(
      `/api/data/search/${encodeURIComponent(objectApiName)}?term=${encodeURIComponent(term)}`
    );
    return response.records;
  },

  /**
   * Get a single record by ID
   */
  async getRecord<T = SObject>(objectApiName: string, id: string): Promise<T> {
    const response = await apiClient.get<{ record: T }>(
      `/api/data/${encodeURIComponent(objectApiName)}/${encodeURIComponent(id)}`
    );
    return response.record;
  },

  /**
   * Create a new record
   */
  async createRecord<T = SObject>(objectApiName: string, data: Partial<T>): Promise<T & { id: string }> {
    const response = await apiClient.post<{ record: T & { id: string } }>(
      `/api/data/${encodeURIComponent(objectApiName)}`,
      data
    );
    return response.record;
  },

  /**
   * Update an existing record
   */
  async updateRecord(objectApiName: string, id: string, updates: Partial<SObject>): Promise<void> {
    await apiClient.patch(
      `/api/data/${encodeURIComponent(objectApiName)}/${encodeURIComponent(id)}`,
      updates
    );
  },

  /**
   * Delete a record (soft delete to recycle bin)
   */
  async deleteRecord(objectApiName: string, id: string): Promise<void> {
    await apiClient.delete(
      `/api/data/${encodeURIComponent(objectApiName)}/${encodeURIComponent(id)}`
    );
  },

  /**
   * Get recycle bin items
   */
  async getRecycleBinItems(scope: 'mine' | 'all' = 'mine'): Promise<RecycleBinItem[]> {
    const response = await apiClient.get<{ items: RecycleBinItem[] }>(`/api/data/recyclebin/items?scope=${scope}`);
    return response.items;
  },

  /**
   * Restore record from recycle bin
   */
  async restoreRecord(id: string): Promise<void> {
    await apiClient.post(`/api/data/recyclebin/restore/${encodeURIComponent(id)}`);
  },

  /**
   * Permanently delete record from recycle bin
   */
  async purgeRecord(id: string): Promise<void> {
    await apiClient.delete(`/api/data/recyclebin/${encodeURIComponent(id)}`);
  },

  /**
   * Run analytics query
   */
  async runAnalytics(query: AnalyticsQuery): Promise<unknown> {
    const response = await apiClient.post<{ result: unknown }>('/api/data/analytics', query);
    return response.result;
  },
  /**
   * Calculate formula fields for a record
   */
  async calculate(objectApiName: string, record: SObject): Promise<SObject> {
    const response = await apiClient.post<{ record: SObject }>(
      `/api/data/${encodeURIComponent(objectApiName)}/calculate`,
      record
    );
    return response.record;
  },

  /**
   * Execute a server-side action
   */
  async executeAction(actionId: string, payload: Record<string, unknown>): Promise<unknown> {
    const response = await apiClient.post<{ result: unknown }>(
      `/api/actions/${encodeURIComponent(actionId)}/execute`,
      payload
    );
    return response.result;
  }
};
