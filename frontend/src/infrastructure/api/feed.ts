import { apiClient } from './client';

export interface SystemComment {
    id: string;
    body: string;
    object_api_name: string;
    record_id: string;
    parent_comment_id?: string;
    is_resolved: boolean;
    created_by: string;
    created_date: string;
}

export const feedAPI = {
    /**
     * Get comments for a record
     */
    async getComments(recordId: string): Promise<SystemComment[]> {
        const response = await apiClient.get<{ comments: SystemComment[] }>(
            `/api/feed/${encodeURIComponent(recordId)}`
        );
        return response.comments;
    },

    /**
     * Create a new comment
     */
    async createComment(comment: Partial<SystemComment>): Promise<SystemComment> {
        const response = await apiClient.post<{ comment: SystemComment }>(
            '/api/feed/comments',
            comment
        );
        return response.comment;
    }
};
