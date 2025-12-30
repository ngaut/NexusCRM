import React, { useState, useEffect } from 'react';
import { MessageSquare, History } from 'lucide-react';
import { useErrorToast } from './ui/Toast';
import { dataAPI } from '../infrastructure/api/data';
import { feedAPI } from '../infrastructure/api/feed';
import { formatDistanceToNow } from 'date-fns';
import { CommentList, Comment } from './feed/CommentList';
import { AuditLogList, AuditLog } from './feed/AuditLogList';
import { SYSTEM_TABLE_NAMES } from '../generated-schema';

interface ActivityFeedProps {
    objectApiName: string;
    recordId: string;
}

type Tab = 'comments' | 'history';

export const ActivityFeed: React.FC<ActivityFeedProps> = ({ objectApiName, recordId }) => {
    const errorToast = useErrorToast();
    const [activeTab, setActiveTab] = useState<Tab>('comments');
    const [comments, setComments] = useState<Comment[]>([]);
    const [history, setHistory] = useState<AuditLog[]>([]);
    const [loading, setLoading] = useState(false);

    useEffect(() => {
        loadData();
    }, [objectApiName, recordId, activeTab]);

    const loadData = async () => {
        setLoading(true);
        try {
            if (activeTab === 'comments') {
                // Fetch Comments via Feed API
                const systemComments = await feedAPI.getComments(recordId);

                const commentIds = systemComments.map(r => r.id);
                // Collect unique user IDs
                const userIds = Array.from(new Set(systemComments.map(c => c.created_by).filter(Boolean)));

                // Fetch Users (Names)
                let userMap = new Map<string, string>();
                if (userIds.length > 0) {
                    const users = await dataAPI.query({
                        objectApiName: SYSTEM_TABLE_NAMES.SYSTEM_USER,
                    });
                    users.forEach(u => {
                        let name = (u.name || u.Name || u.full_name || u.Full_Name) as string;
                        if (!name && u.first_name) {
                            name = `${u.first_name} ${u.last_name || ''}`.trim();
                        }
                        if (!name) {
                            name = u.email as string;
                        }

                        if (u.id && name) {
                            userMap.set(u.id as string, name);
                        }
                    });
                }

                let allAttachments: { id: string; parent_id: string; name: string; storage_path: string; mime_type: string }[] = [];
                if (commentIds.length > 0) {
                    // Fetch attachments for found comments
                    const filePromises = systemComments.map(c =>
                        dataAPI.query({
                            objectApiName: SYSTEM_TABLE_NAMES.SYSTEM_FILE,
                            filterExpr: `parent_id == '${c.id}'`,
                            sortField: 'created_date',
                            sortDirection: 'ASC'
                        })
                    );
                    const filesResults = await Promise.all(filePromises);
                    allAttachments = filesResults.flat().map(f => ({
                        id: f.id as string,
                        parent_id: f.parent_id as string,
                        name: f.name as string,
                        storage_path: f.storage_path as string,
                        mime_type: f.mime_type as string
                    }));
                }

                const rawComments = systemComments.map(r => {
                    let timeDisplay = 'Just now';
                    try {
                        if (r.created_date) {
                            timeDisplay = formatDistanceToNow(new Date(r.created_date), { addSuffix: true });
                        }
                    } catch (e) {
                        console.debug('ActivityFeed: Failed to parse comment date:', r.created_date, e);
                    }

                    // Find attachments
                    const myFiles = allAttachments.filter(f => f.parent_id === r.id).map(f => ({
                        id: f.id,
                        parent_id: f.parent_id,
                        name: f.name,
                        storage_path: f.storage_path,
                        mime_type: f.mime_type
                    }));

                    return {
                        id: r.id,
                        body: r.body,
                        created_date: timeDisplay,
                        created_by: r.created_by,
                        author_name: userMap.get(r.created_by) || 'User',
                        parent_comment_id: r.parent_comment_id,
                        is_resolved: !!r.is_resolved,
                        attachments: myFiles,
                        replies: [],
                        raw_date: r.created_date ? new Date(r.created_date).getTime() : 0
                    } as Comment;
                });

                // Build Thread
                const commentMap = new Map<string, Comment>();
                rawComments.forEach(c => {
                    c.replies = [];
                    commentMap.set(c.id, c as Comment);
                });

                const rootComments: Comment[] = [];
                rawComments.forEach(c => {
                    if (c.parent_comment_id && commentMap.has(c.parent_comment_id)) {
                        commentMap.get(c.parent_comment_id)!.replies!.push(c as Comment);
                    } else {
                        rootComments.push(c as Comment);
                    }
                });

                // Sort root comments by raw_date DESC (Newest first)
                rootComments.sort((a, b) => (b.raw_date || 0) - (a.raw_date || 0));

                setComments(rootComments);
            } else {
                // History Logic
                const results = await dataAPI.query({
                    objectApiName: SYSTEM_TABLE_NAMES.SYSTEM_AUDITLOG,
                    filterExpr: `record_id == '${recordId}' && object_api_name == '${objectApiName}'`,
                    sortField: 'changed_at',
                    sortDirection: 'DESC'
                });
                const mappedHistory = results.map(r => {
                    let timeDisplay = 'Unknown time';
                    try {
                        if (r.changed_at) {
                            timeDisplay = formatDistanceToNow(new Date(r.changed_at as string), { addSuffix: true });
                        }
                    } catch (e) {
                        console.debug('ActivityFeed: Failed to parse history date:', r.changed_at, e);
                    }
                    return {
                        id: r.id as string,
                        field_name: r.field_name as string,
                        old_value: r.old_value as string,
                        new_value: r.new_value as string,
                        changed_by: (r.changed_by_id__name as string) || 'User',
                        changed_at: timeDisplay
                    };
                });
                setHistory(mappedHistory as AuditLog[]);
            }
        } catch (err) {
            errorToast('Failed to load activity feed: ' + (err instanceof Error ? err.message : 'Unknown error'));
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="bg-white border border-slate-200 rounded-lg shadow-sm h-full flex flex-col">
            {/* Tabs Header */}
            <div className="flex border-b border-slate-200">
                <button
                    onClick={() => setActiveTab('comments')}
                    className={`flex-1 py-3 text-sm font-medium flex items-center justify-center gap-2 ${activeTab === 'comments'
                        ? 'text-blue-600 border-b-2 border-blue-600'
                        : 'text-slate-500 hover:text-slate-700'
                        }`}
                >
                    <MessageSquare size={16} />
                    Comments
                </button>
                <button
                    onClick={() => setActiveTab('history')}
                    className={`flex-1 py-3 text-sm font-medium flex items-center justify-center gap-2 ${activeTab === 'history'
                        ? 'text-blue-600 border-b-2 border-blue-600'
                        : 'text-slate-500 hover:text-slate-700'
                        }`}
                >
                    <History size={16} />
                    History
                </button>
            </div>

            <div className="flex-1 overflow-auto p-4 content-start">
                {activeTab === 'comments' ? (
                    <CommentList
                        objectApiName={objectApiName}
                        recordId={recordId}
                        comments={comments}
                        onRefresh={loadData}
                    />
                ) : (
                    <AuditLogList history={history} />
                )}
            </div>
        </div >
    );
};
