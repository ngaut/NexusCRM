import React from 'react';
import { Users, User, Trash2 } from 'lucide-react';
import type { UserOption } from './ShareUserSearch';

interface ShareEntry {
    id: string;
    share_with_user_id?: string;
    share_with_group_id?: string;
    access_level: 'Read' | 'Edit';
    created_date: string;
    user_name?: string;
    group_name?: string;
}

interface ShareListProps {
    shares: ShareEntry[];
    onRemove: (shareId: string) => void;
}

export const ShareList: React.FC<ShareListProps> = ({ shares, onRemove }) => {
    return (
        <div className="mt-6">
            <h3 className="text-sm font-semibold text-gray-500 uppercase tracking-wide mb-3">
                Current Access ({shares.length})
            </h3>
            {shares.length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                    <Users className="w-10 h-10 mx-auto mb-2 text-gray-300" />
                    <p>No one else has access to this record</p>
                </div>
            ) : (
                <div className="space-y-2">
                    {shares.map((share) => (
                        <div
                            key={share.id}
                            className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
                        >
                            <div className="flex items-center gap-3">
                                <div className={`p-2 rounded-full ${share.share_with_user_id ? 'bg-blue-100' : 'bg-purple-100'
                                    }`}>
                                    {share.share_with_user_id ? (
                                        <User className="w-4 h-4 text-blue-600" />
                                    ) : (
                                        <Users className="w-4 h-4 text-purple-600" />
                                    )}
                                </div>
                                <div>
                                    <div className="font-medium text-gray-900">
                                        {share.user_name || share.group_name || 'User/Group'}
                                    </div>
                                    <div className="text-xs text-gray-500">
                                        {share.access_level} access
                                    </div>
                                </div>
                            </div>
                            <button
                                onClick={() => onRemove(share.id)}
                                className="p-2 text-gray-400 hover:text-red-500 hover:bg-red-50 rounded-lg transition-colors"
                            >
                                <Trash2 className="w-4 h-4" />
                            </button>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};
