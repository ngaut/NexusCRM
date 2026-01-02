import React from 'react';
import { ConversationSummary } from '../../infrastructure/api/agent';
import './ConversationSidebar.css';

interface ConversationSidebarProps {
    conversations: ConversationSummary[];
    currentConversationId: string | null;
    isOpen: boolean;
    onToggle: () => void;
    onSelectConversation: (id: string) => void;
    onNewChat: () => void;
    onDeleteConversation: (id: string) => void;
    isLoading?: boolean;
}

export const ConversationSidebar: React.FC<ConversationSidebarProps> = ({
    conversations,
    currentConversationId,
    isOpen,
    onToggle,
    onSelectConversation,
    onNewChat,
    onDeleteConversation,
    isLoading = false,
}) => {
    const formatRelativeTime = (dateStr: string): string => {
        const date = new Date(dateStr);
        const now = new Date();
        const diffMs = now.getTime() - date.getTime();
        const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

        if (diffDays === 0) return 'Today';
        if (diffDays === 1) return 'Yesterday';
        if (diffDays < 7) return `${diffDays} days ago`;
        return date.toLocaleDateString();
    };

    return (
        <div className={`conversation-sidebar ${isOpen ? 'open' : 'collapsed'}`}>
            <button
                className="sidebar-toggle"
                onClick={onToggle}
                title={isOpen ? 'Collapse sidebar' : 'Expand sidebar'}
            >
                {isOpen ? '◀' : '▶'}
            </button>

            {isOpen && (
                <div className="sidebar-content">
                    <button
                        className="new-chat-button"
                        onClick={onNewChat}
                        disabled={isLoading}
                    >
                        <span className="plus-icon">+</span>
                        New Chat
                    </button>

                    <div className="conversations-list">
                        {conversations.length === 0 ? (
                            <div className="no-conversations">No conversations yet</div>
                        ) : (
                            conversations.map((conv) => (
                                <div
                                    key={conv.id}
                                    className={`conversation-item ${conv.id === currentConversationId ? 'active' : ''}`}
                                    onClick={() => onSelectConversation(conv.id)}
                                >
                                    <div className="conversation-info">
                                        <div className="conversation-title" title={conv.title}>
                                            {conv.is_active && <span className="active-dot">●</span>}
                                            {conv.title}
                                        </div>
                                        <div className="conversation-time">
                                            {formatRelativeTime(conv.last_modified_date)}
                                        </div>
                                    </div>
                                    <button
                                        className="delete-button"
                                        onClick={(e) => {
                                            e.stopPropagation();
                                            e.preventDefault();
                                            onDeleteConversation(conv.id);
                                        }}
                                        title="Delete conversation"
                                    >
                                        🗑️
                                    </button>
                                </div>
                            ))
                        )}
                    </div>
                </div>
            )}
        </div>
    );
};
