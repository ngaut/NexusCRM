import React, { useState, useEffect, useRef } from 'react';
import { Bell, Check, Trash2, X } from 'lucide-react';
import { SObject } from '../types';
import { SYSTEM_TABLE_NAMES } from '../generated-schema';
import { dataAPI } from '../infrastructure/api/data';
import { useNavigate } from 'react-router-dom';
import { formatDistanceToNow } from 'date-fns';

interface Notification extends SObject {
    recipient_id: string;
    title: string;
    body: string;
    link?: string;
    is_read: boolean;
    created_date: string;
}

export const NotificationCenter: React.FC = () => {
    const [notifications, setNotifications] = useState<Notification[]>([]);
    const [unreadCount, setUnreadCount] = useState(0);
    const [isOpen, setIsOpen] = useState(false);
    const menuRef = useRef<HTMLDivElement>(null);
    const navigate = useNavigate();

    const fetchNotifications = async () => {
        try {
            const results = await dataAPI.query({
                objectApiName: SYSTEM_TABLE_NAMES.SYSTEM_NOTIFICATION,
                sortField: 'created_date',
                sortDirection: 'DESC',
                filterExpr: "recipient_id == 'CURRENT_USER'"
            });
            // Mocking for now if backend doesn't support CURRENT_USER in query criteria automatically (it usually doesn't). e.g. persistence service
            // actually we should filter by current user ID. But we need user ID context.
            // For now, I'll rely on RLS (Row Level Security) if enabled, or simple query.
            // But wait, `dataAPI.query` takes explicit criteria.
            // I need `user.id`.

            // Re-fetch logic needed with User ID context.
            // Assuming this component is wrapped or used inside Layout where we can get user.
        } catch {
            // Notification loading failure - silently falls back; will be handled by actual component
        }
    };

    // Since I need user ID, I'll wrap the logic in a component that uses useRuntime
    return <NotificationCenterInner />;
};

import { useRuntime } from '../contexts/RuntimeContext';

const NotificationCenterInner: React.FC = () => {
    const { user } = useRuntime();
    const [notifications, setNotifications] = useState<Notification[]>([]);
    const [isOpen, setIsOpen] = useState(false);
    const menuRef = useRef<HTMLDivElement>(null);
    const navigate = useNavigate();

    useEffect(() => {
        if (user?.id) {
            loadNotifications();
            // Poll every 30s
            const interval = setInterval(loadNotifications, 30000);
            return () => clearInterval(interval);
        }
    }, [user?.id]);

    const loadNotifications = async () => {
        if (!user?.id) return;
        try {
            const results = await dataAPI.query({
                objectApiName: SYSTEM_TABLE_NAMES.SYSTEM_NOTIFICATION,
                filterExpr: `owner_id == '${user.id}' AND is_read == false`,
                sortField: 'created_date',
                sortDirection: 'DESC',
                limit: 50
            });

            // Hydrate
            const mapped = results.map(r => ({
                id: r.id as string,
                recipient_id: r.recipient_id as string,
                title: r.title as string,
                body: r.body as string,
                link: r.link as string,
                is_read: !!r.is_read,
                created_date: r.created_date as string
            }));
            setNotifications(mapped);

            // Mark as read immediately on open? Or just fetch?
            // Actually, we usually want to verify we have unread ones.
            const count = await dataAPI.query({
                objectApiName: SYSTEM_TABLE_NAMES.SYSTEM_NOTIFICATION,
                filterExpr: `owner_id == '${user.id}' AND is_read == false`,
                limit: 1 // Just need to know if > 0
            });

        } catch {
            // Notification load failure - handled via empty state
        }
    };

    const unreadCount = notifications.filter(n => !n.is_read).length;

    const handleMarkAsRead = async (id: string) => {
        try {
            await dataAPI.updateRecord(SYSTEM_TABLE_NAMES.SYSTEM_NOTIFICATION, id, { is_read: true });
            setNotifications(prev => prev.map(n => n.id === id ? { ...n, is_read: true } : n));
        } catch {
            // Mark as read failure - silently continue
        }
    };

    const handleMarkAllRead = async () => {
        // Implement bulk update if API supports, or parallel
        const unread = notifications.filter(n => !n.is_read);
        for (const n of unread) {
            handleMarkAsRead(n.id);
        }
    };

    // Close on click outside
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
                setIsOpen(false);
            }
        };
        if (isOpen) document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, [isOpen]);

    return (
        <div className="relative" ref={menuRef}>
            <button
                onClick={() => setIsOpen(!isOpen)}
                className="p-2 text-slate-500 hover:bg-slate-100 rounded-full transition-colors relative"
            >
                <Bell size={20} />
                {unreadCount > 0 && (
                    <span className="absolute top-2 right-2 w-2 h-2 bg-red-500 rounded-full border-2 border-white"></span>
                )}
            </button>

            {isOpen && (
                <div className="absolute right-0 mt-2 w-80 bg-white border border-slate-200 rounded-lg shadow-lg py-1 z-50 max-h-[400px] flex flex-col">
                    <div className="px-4 py-3 border-b border-slate-100 flex justify-between items-center">
                        <h3 className="font-semibold text-slate-900">Notifications</h3>
                        {unreadCount > 0 && (
                            <button onClick={handleMarkAllRead} className="text-xs text-blue-600 hover:text-blue-700 font-medium">
                                Mark all as read
                            </button>
                        )}
                    </div>
                    <div className="overflow-y-auto flex-1">
                        {notifications.length === 0 ? (
                            <div className="p-8 text-center text-slate-500 text-sm">
                                No notifications
                            </div>
                        ) : (
                            <div>
                                {notifications.map(n => (
                                    <div
                                        key={n.id}
                                        className={`px - 4 py - 3 border - b border - slate - 50 hover: bg - slate - 50 transition - colors cursor - pointer ${!n.is_read ? 'bg-blue-50/30' : ''} `}
                                        onClick={() => {
                                            if (n.link) {
                                                navigate(n.link);
                                                setIsOpen(false);
                                            }
                                            if (!n.is_read) handleMarkAsRead(n.id);
                                        }}
                                    >
                                        <div className="flex gap-3">
                                            <div className="flex-1">
                                                <p className={`text - sm ${!n.is_read ? 'font-semibold text-slate-900' : 'text-slate-700'} `}>
                                                    {n.title}
                                                </p>
                                                <p className="text-xs text-slate-500 mt-1 line-clamp-2">{n.body}</p>
                                                <p className="text-[10px] text-slate-400 mt-1">
                                                    {n.created_date ? formatDistanceToNow(new Date(n.created_date), { addSuffix: true }) : ''}
                                                </p>
                                            </div>
                                            {!n.is_read && (
                                                <div className="mt-1">
                                                    <div className="w-2 h-2 rounded-full bg-blue-500" />
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                </div>
            )}
        </div>
    );
};
