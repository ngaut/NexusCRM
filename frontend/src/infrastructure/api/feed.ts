import { apiClient } from './client';
import { API_ENDPOINTS } from './endpoints';
import { SystemComment } from '../../generated-schema';

// Re-export for consumers
export type { SystemComment } from '../../generated-schema';

export const feedAPI = {
    /**
     * Get comments for a record
     */
    async getComments(recordId: string): Promise<SystemComment[]> {
        const response = await apiClient.get<{ data: SystemComment[] }>(
            API_ENDPOINTS.FEED.RECORD(recordId)
        );
        return response.data;
    },

    /**
     * Create a new comment
     */
    async createComment(comment: Partial<SystemComment>): Promise<SystemComment> {
        const response = await apiClient.post<{ data: SystemComment }>(
            API_ENDPOINTS.FEED.COMMENTS,
            comment
        );
        return response.data;
    }
};
