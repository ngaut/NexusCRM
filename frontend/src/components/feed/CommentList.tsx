import React, { useState } from 'react';
import { Paperclip, Send, File, MessageSquare } from 'lucide-react';
import { formatDistanceToNow } from 'date-fns';
import DOMPurify from 'dompurify';
import { RichTextEditor } from '../RichTextEditor';
import { useErrorToast } from '../ui/Toast';
import { ConfirmationModal } from '../modals/ConfirmationModal';
import { dataAPI } from '../../infrastructure/api/data';
import { feedAPI } from '../../infrastructure/api/feed';
import { filesAPI } from '../../infrastructure/api/files';
import { useRuntime } from '../../contexts/RuntimeContext';
import { SYSTEM_TABLE_NAMES } from '../../generated-schema';

interface Attachment {
    id: string;
    parent_id: string;
    name: string;
    storage_path: string;
    mime_type: string;
}

export interface Comment {
    id: string;
    body: string; // HTML
    created_date: string;
    created_by: string;
    author_name: string;
    replies?: Comment[];
    parent_comment_id?: string;
    is_resolved?: boolean;
    attachments?: Attachment[];
    raw_date?: number; // Timestamp for sorting
}

interface CommentListProps {
    objectApiName: string;
    recordId: string;
    comments: Comment[];
    onRefresh: () => void;
}

export const CommentList: React.FC<CommentListProps> = ({ objectApiName, recordId, comments, onRefresh }) => {
    const errorToast = useErrorToast();
    const { user } = useRuntime();
    const [newComment, setNewComment] = useState('');
    const [replyTo, setReplyTo] = useState<string | null>(null);
    const [pendingAttachments, setPendingAttachments] = useState<{ name: string, path: string, size: number, mime: string }[]>([]);
    const [showConfirmModal, setShowConfirmModal] = useState(false);
    const fileInputRef = React.useRef<HTMLInputElement>(null);

    const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
        if (!e.target.files || e.target.files.length === 0) return;
        const file = e.target.files[0];

        try {
            const data = await filesAPI.upload(file);

            setPendingAttachments(prev => [...prev, {
                name: data.name,
                path: data.path,
                size: data.size,
                mime: data.mime
            }]);
        } catch (error) {
            console.warn('CommentList: File upload failed:', error instanceof Error ? error.message : error);
            errorToast('File upload failed');
        }
    };

    const handlePostClick = () => {
        if (!newComment.trim() && pendingAttachments.length === 0) return;
        setShowConfirmModal(true);
    };

    const handleConfirmPost = async () => {
        setShowConfirmModal(false);
        try {
            const commentRes = await feedAPI.createComment({
                object_api_name: objectApiName,
                record_id: recordId,
                body: newComment, // HTML from editor
                parent_comment_id: replyTo || undefined
            });

            // Create Attachments
            if (pendingAttachments.length > 0) {
                await Promise.all(pendingAttachments.map(f =>
                    dataAPI.createRecord(SYSTEM_TABLE_NAMES.SYSTEM_FILE, {
                        parent_id: commentRes.id,
                        name: f.name,
                        storage_path: f.path,
                        mime_type: f.mime,
                        size_bytes: f.size,
                        created_by_id: user?.id
                    })
                ));
            }

            setNewComment('');
            setPendingAttachments([]);
            setReplyTo(null);
            onRefresh();
        } catch {
            errorToast('Failed to post comment');
        }
    };

    const renderComment = (c: Comment, isReply = false) => (
        <div key={c.id} className={`group ${isReply ? 'ml-11 mt-3 space-y-3 border-l-2 border-slate-100 pl-4' : ''}`}>
            <div className="flex gap-3">
                <div className={`rounded-full bg-slate-100 flex items-center justify-center text-slate-600 font-bold ${isReply ? 'w-6 h-6 text-[10px]' : 'w-8 h-8 text-xs'}`}>
                    {c.author_name?.substring(0, 2).toUpperCase() || 'U'}
                </div>
                <div className="flex-1">
                    <div className="bg-slate-50 p-3 rounded-lg rounded-tl-none relative">
                        <div className="flex justify-between items-start mb-1">
                            <span className={`font-medium text-slate-900 ${isReply ? 'text-xs' : 'text-sm'}`}>{c.author_name || 'User'}</span>
                            <div className="flex items-center gap-2">
                                <span className={`text-slate-400 ${isReply ? 'text-[10px]' : 'text-xs'}`}>{c.created_date}</span>
                                {c.is_resolved && <span className="text-xs bg-green-100 text-green-700 px-1.5 py-0.5 rounded-full font-medium">Resolved</span>}
                            </div>
                        </div>
                        {/* Render HTML Body - sanitized to prevent XSS */}
                        <div className={`prose prose-sm max-w-none text-slate-700 ${isReply ? 'text-xs' : 'text-sm'}`} dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(c.body) }} />

                        {/* Attachments */}
                        {c.attachments && c.attachments.length > 0 && (
                            <div className="mt-3 flex flex-wrap gap-2">
                                {c.attachments.map(a => (
                                    <div key={a.id} className="flex items-center gap-2 bg-white border border-slate-200 px-2 py-1 rounded text-xs text-blue-600">
                                        <Paperclip size={12} />
                                        <a href={`/${a.storage_path}`} target="_blank" rel="noopener noreferrer" className="hover:underline">{a.name}</a>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>

                    <div className="flex gap-4 mt-1 ml-1">
                        <button onClick={() => setReplyTo(c.id)} className="text-xs text-slate-500 hover:text-blue-600 font-medium">Reply</button>
                        {!isReply && !c.is_resolved && (
                            <button onClick={() => {
                                // handle resolve
                                dataAPI.updateRecord(SYSTEM_TABLE_NAMES.SYSTEM_COMMENT, c.id, { is_resolved: true }).then(onRefresh);
                            }} className="text-xs text-slate-500 hover:text-green-600 font-medium">Resolve</button>
                        )}
                    </div>
                </div>
            </div>
            {/* Replies */}
            {c.replies?.map(r => renderComment(r, true))}
        </div>
    );

    return (
        <div className="space-y-4">
            {/* Comment Input */}
            <div className="flex gap-3">
                <div className="w-8 h-8 rounded-full bg-blue-100 flex items-center justify-center text-blue-600 font-bold text-xs">
                    {user?.name?.substring(0, 2).toUpperCase()}
                </div>
                <div className="flex-1">
                    {replyTo && (
                        <div className="flex justify-between items-center bg-blue-50 p-2 rounded mb-2 text-xs text-blue-600">
                            <span>Replying to comment...</span>
                            <button onClick={() => setReplyTo(null)} className="hover:underline">Cancel</button>
                        </div>
                    )}

                    <RichTextEditor
                        value={newComment}
                        onChange={setNewComment}
                        placeholder={replyTo ? "Reply..." : "Write a comment... (@ to mention)"}
                    />

                    {/* Pending Attachments */}
                    {pendingAttachments.length > 0 && (
                        <div className="mt-2 text-xs text-slate-500 flex flex-wrap gap-2">
                            {pendingAttachments.map((f, i) => (
                                <div key={i} className="bg-slate-100 rounded px-2 py-1 flex items-center gap-1">
                                    <File size={12} />
                                    {f.name}
                                </div>
                            ))}
                        </div>
                    )}

                    <div className="mt-2 flex justify-between items-center">
                        <div className="flex gap-2">
                            <input
                                type="file"
                                ref={fileInputRef}
                                className="hidden"
                                onChange={handleFileUpload}
                            />
                            <button
                                onClick={() => fileInputRef.current?.click()}
                                className="text-slate-500 hover:text-blue-600 p-1 rounded hover:bg-slate-100"
                                title="Attach File"
                            >
                                <Paperclip size={16} />
                            </button>
                        </div>
                        <button
                            onClick={handlePostClick}
                            className="bg-blue-600 text-white px-3 py-1.5 rounded text-sm font-medium hover:bg-blue-700 flex items-center gap-2"
                            disabled={!newComment && pendingAttachments.length === 0}
                        >
                            <Send size={14} />
                            Post
                        </button>
                    </div>
                </div>
            </div>

            {/* List */}
            <div className="space-y-6 pt-4">
                {comments.length === 0 && <div className="text-center text-slate-400 text-sm py-4">No comments yet</div>}
                {comments.map(c => renderComment(c))}
            </div>

            <ConfirmationModal
                isOpen={showConfirmModal}
                onClose={() => setShowConfirmModal(false)}
                onConfirm={handleConfirmPost}
                title="Post Comment"
                message="Are you sure you want to post this comment?"
                confirmLabel="Post"
                variant="info"
            />
        </div>
    );
};
