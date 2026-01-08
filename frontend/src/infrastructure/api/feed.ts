import { apiClient } from './client';
import { API_ENDPOINTS } from './endpoints';
import { COMMON_FIELDS } from '../../core/constants';

export interface SystemComment {
    [COMMON_FIELDS.ID]: string;
    body: string;
    [COMMON_FIELDS.OBJECT_API_NAME]: string;
    [COMMON_FIELDS.RECORD_ID]: string;
    parent_comment_id?: string;
    is_resolved: boolean;
    [COMMON_FIELDS.CREATED_BY]: string;
    [COMMON_FIELDS.CREATED_DATE]: string;
}

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
