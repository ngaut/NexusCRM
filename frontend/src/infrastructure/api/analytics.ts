import { apiClient } from './client';
import { API_ENDPOINTS } from './endpoints';

export interface AnalyticsResult {
    data: Record<string, unknown>[];
}

export const analyticsAPI = {
    executeAdminQuery: async (sql: string, params: unknown[] = []): Promise<AnalyticsResult> => {
        const response = await apiClient.post<AnalyticsResult>(API_ENDPOINTS.ANALYTICS.QUERY, { sql, params });
        return response;
    }
};
