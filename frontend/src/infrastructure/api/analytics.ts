import { apiClient } from './client';

export interface AnalyticsResult {
    success: boolean;
    results: Record<string, unknown>[];
}

export const analyticsAPI = {
    executeAdminQuery: async (sql: string, params: unknown[] = []): Promise<AnalyticsResult> => {
        const response = await apiClient.post<AnalyticsResult>('/api/analytics/query', { sql, params });
        return response;
    }
};
