import React, { useState, useEffect, useCallback } from 'react';
import { X, Users, Users2, UserPlus } from 'lucide-react';
import { dataAPI } from '../../infrastructure/api/data';
import { useSuccessToast, useErrorToast } from '../../components/ui/Toast';
import type { SObject } from '../../types';

// Sub-components
import { ShareUserSearch, UserOption } from '../sharing/ShareUserSearch';
import { ShareList } from '../sharing/ShareList';

interface ShareRecordModalProps {
    isOpen: boolean;
    onClose: () => void;
    objectApiName: string;
    recordId: string;
    recordName: string;
}

interface ShareEntry {
    id: string;
    share_with_user_id?: string;
    share_with_group_id?: string;
    access_level: 'Read' | 'Edit';
    created_date: string;
    user_name?: string;
    group_name?: string;
}

export const ShareRecordModal: React.FC<ShareRecordModalProps> = ({
    isOpen,
    onClose,
    objectApiName,
    recordId,
    recordName
}) => {
    const showSuccess = useSuccessToast();
    const showError = useErrorToast();

    // State
    const [currentShares, setCurrentShares] = useState<ShareEntry[]>([]);
    const [searchResults, setSearchResults] = useState<UserOption[]>([]);
    const [selectedEntity, setSelectedEntity] = useState<UserOption | null>(null);
    const [accessLevel, setAccessLevel] = useState<'Read' | 'Edit'>('Read');
    const [loading, setLoading] = useState(false);
    const [searching, setSearching] = useState(false);
    const [activeTab, setActiveTab] = useState<'share' | 'team'>('share');

    // Fetch current shares
    const fetchShares = useCallback(async () => {
        try {
            const response = await dataAPI.query({
                objectApiName: '_system_recordshare',
                filterExpr: `object_api_name == '${objectApiName}' && record_id == '${recordId}'`
            });
            setCurrentShares((response as unknown as ShareEntry[]) || []);
        } catch {
            // Share fetch failure - handled via empty state
        }
    }, [objectApiName, recordId]);

    useEffect(() => {
        if (isOpen) {
            fetchShares();
            setSelectedEntity(null);
            setSearchResults([]);
        }
    }, [isOpen, fetchShares]);

    // Search users and groups
    const handleSearch = async (query: string): Promise<UserOption[]> => {
        setSearching(true);
        try {
            // Search users
            const usersRes = await dataAPI.query({
                objectApiName: '_system_user',
                filterExpr: `username LIKE '%${query}%'`
            });
            const users: UserOption[] = (usersRes || []).map((u: SObject) => ({
                id: u.id as string,
                name: u.username as string,
                email: u.email as string,
                type: 'user' as const
            }));

            // Search groups
            const groupsRes = await dataAPI.query({
                objectApiName: '_system_group',
                filterExpr: `name LIKE '%${query}%'`
            });
            const groups: UserOption[] = (groupsRes || []).map((g: SObject) => ({
                id: g.id as string,
                name: (g.label || g.name) as string,
                type: 'group' as const
            }));

            const results = [...users, ...groups];
            setSearchResults(results);
            return results;
        } catch {
            setSearchResults([]);
            return [];
        } finally {
            setSearching(false);
        }
    };

    // Add share
    const handleAddShare = async () => {
        if (!selectedEntity) return;

        setLoading(true);
        try {
            await dataAPI.createRecord('_system_recordshare', {
                object_api_name: objectApiName,
                record_id: recordId,
                share_with_user_id: selectedEntity.type === 'user' ? selectedEntity.id : null,
                share_with_group_id: selectedEntity.type === 'group' ? selectedEntity.id : null,
                access_level: accessLevel
            });

            showSuccess(`Shared with ${selectedEntity.name}`);
            setSelectedEntity(null);
            setSearchResults([]);
            fetchShares();
        } catch (err) {
            showError(err instanceof Error ? err.message : 'Failed to share');
        } finally {
            setLoading(false);
        }
    };

    // Remove share
    const handleRemoveShare = async (shareId: string) => {
        try {
            await dataAPI.deleteRecord('_system_recordshare', shareId);
            showSuccess('Share removed');
            fetchShares();
        } catch (err) {
            showError(err instanceof Error ? err.message : 'Failed to remove share');
        }
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
            {/* Backdrop */}
            <div
                className="absolute inset-0 bg-black/50 backdrop-blur-sm"
                onClick={onClose}
            />

            {/* Modal */}
            <div className="relative bg-white rounded-2xl shadow-2xl w-full max-w-lg mx-4 overflow-hidden animate-in fade-in zoom-in-95 duration-200">
                {/* Header */}
                <div className="bg-gradient-to-r from-blue-600 to-indigo-600 px-6 py-5 text-white">
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3">
                            <div className="p-2 bg-white/20 rounded-lg">
                                <Users className="w-5 h-5" />
                            </div>
                            <div>
                                <h2 className="text-lg font-semibold">Share Record</h2>
                                <p className="text-sm text-blue-100 truncate max-w-[280px]">{recordName}</p>
                            </div>
                        </div>
                        <button
                            onClick={onClose}
                            className="p-1.5 hover:bg-white/20 rounded-lg transition-colors"
                        >
                            <X className="w-5 h-5" />
                        </button>
                    </div>

                    {/* Tabs */}
                    <div className="flex gap-2 mt-4">
                        <button
                            onClick={() => setActiveTab('share')}
                            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${activeTab === 'share'
                                ? 'bg-white text-blue-600'
                                : 'bg-white/20 text-white hover:bg-white/30'
                                }`}
                        >
                            <Users className="w-4 h-4 inline mr-2" />
                            Sharing
                        </button>
                        <button
                            onClick={() => setActiveTab('team')}
                            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${activeTab === 'team'
                                ? 'bg-white text-blue-600'
                                : 'bg-white/20 text-white hover:bg-white/30'
                                }`}
                        >
                            <Users2 className="w-4 h-4 inline mr-2" />
                            Team
                        </button>
                    </div>
                </div>

                {/* Content */}
                <div className="p-6 max-h-[60vh] overflow-y-auto">
                    {activeTab === 'share' && (
                        <>
                            <ShareUserSearch
                                onSearch={handleSearch}
                                onSelect={setSelectedEntity}
                                searching={searching}
                                searchResults={searchResults}
                                selectedEntity={selectedEntity}
                                onClearSelection={() => setSelectedEntity(null)}
                                accessLevel={accessLevel}
                                onAccessLevelChange={setAccessLevel}
                            />

                            <ShareList
                                shares={currentShares}
                                onRemove={handleRemoveShare}
                            />
                        </>
                    )}

                    {activeTab === 'team' && (
                        <div className="text-center py-8 text-gray-500">
                            <Users2 className="w-12 h-12 mx-auto mb-3 text-gray-300" />
                            <p className="font-medium">Team Management</p>
                            <p className="text-sm mt-1">Add team members who work on this record</p>
                            <button className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors inline-flex items-center gap-2">
                                <UserPlus className="w-4 h-4" />
                                Add Team Member
                            </button>
                        </div>
                    )}
                </div>

                {/* Footer */}
                <div className="px-6 py-4 bg-gray-50 border-t flex justify-between items-center">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 text-gray-600 hover:text-gray-900 font-medium transition-colors"
                    >
                        Cancel
                    </button>
                    {selectedEntity && (
                        <button
                            onClick={handleAddShare}
                            disabled={loading}
                            className="px-5 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium disabled:opacity-50 flex items-center gap-2"
                        >
                            {loading ? (
                                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                            ) : (
                                <UserPlus className="w-4 h-4" />
                            )}
                            Share
                        </button>
                    )}
                </div>
            </div>
        </div>
    );
};

export default ShareRecordModal;
