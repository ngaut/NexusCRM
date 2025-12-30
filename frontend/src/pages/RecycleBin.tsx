import React, { useEffect, useState } from 'react';
import { dataAPI } from '../infrastructure/api/data';
import { Trash2, RotateCcw, Filter, CheckSquare, Square } from 'lucide-react';
import { useNotification } from '../contexts/NotificationContext';
import { ConfirmationModal } from '../components/modals/ConfirmationModal';
import { useRuntime } from '../contexts/RuntimeContext';
import { COMMON_FIELDS } from '../core/constants';
import type { RecycleBinItem } from '../types';

export const RecycleBin: React.FC = () => {
    const [items, setItems] = useState<RecycleBinItem[]>([]);
    const [loading, setLoading] = useState(true);
    const { success, error: showError } = useNotification();
    const { user } = useRuntime();

    // Filter State
    const [scope, setScope] = useState<'mine' | 'all'>('mine');

    // Selection State
    const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());

    // Confirmation logic
    const [action, setAction] = useState<'restore' | 'purge' | 'empty_org' | null>(null);
    const [processing, setProcessing] = useState(false);

    const load = async () => {
        setLoading(true);
        try {
            const data = await dataAPI.getRecycleBinItems(scope);
            setItems(data);
            setSelectedIds(new Set());
        } catch (err) {
            showError("Failed to load", err instanceof Error ? err.message : "Could not load recycle bin items");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => { load(); }, [scope]);

    const handleSelectAll = () => {
        if (selectedIds.size === items.length) {
            setSelectedIds(new Set());
        } else {
            setSelectedIds(new Set(items.map(i => i[COMMON_FIELDS.RECORD_ID] as string))); // Store original record_ids for actions
        }
    };

    const handleSelectOne = (id: string) => {
        const newSet = new Set(selectedIds);
        if (newSet.has(id)) newSet.delete(id);
        else newSet.add(id);
        setSelectedIds(newSet);
    };

    const handleExecute = async () => {
        if (!action || action === 'empty_org') return;
        setProcessing(true);
        try {
            const ids = Array.from(selectedIds);
            const promises = ids.map(id =>
                action === 'restore' ? dataAPI.restoreRecord(id) : dataAPI.purgeRecord(id)
            );

            await Promise.all(promises);
            success(action === 'restore' ? 'Restored' : 'Purged', `Successfully processed ${ids.length} records.`);
            await load();
            setAction(null);
        } catch (err) {
            showError('Error', err instanceof Error ? err.message : 'Operation failed');
        } finally {
            setProcessing(false);
        }
    };

    const isAdmin = user?.profile_id === 'system_admin';

    return (
        <div className="max-w-7xl mx-auto p-6">
            <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6">
                <div>
                    <h1 className="text-2xl font-bold text-slate-800 flex items-center gap-2">
                        <Trash2 className="text-slate-600" /> Recycle Bin
                    </h1>
                    <p className="text-slate-500">Manage deleted records.</p>
                </div>

                <div className="flex items-center gap-3">
                    {/* Scope Filter */}
                    <div className="relative">
                        <select
                            value={scope}
                            onChange={e => setScope(e.target.value as 'mine' | 'all')}
                            className="appearance-none bg-white border border-slate-300 text-slate-700 py-2 pl-4 pr-10 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent font-medium"
                        >
                            <option value="mine">My Recycle Bin</option>
                            <option value="all" disabled={!isAdmin}>Org Recycle Bin {(!isAdmin) && '(Admin Only)'}</option>
                        </select>
                        <Filter size={16} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 pointer-events-none" />
                    </div>

                    {selectedIds.size > 0 && (
                        <>
                            <button
                                onClick={() => setAction('restore')}
                                className="px-4 py-2 bg-blue-600 text-white rounded-lg flex items-center gap-2 hover:bg-blue-700 transition-colors font-medium shadow-sm"
                            >
                                <RotateCcw size={16} /> Restore ({selectedIds.size})
                            </button>
                            <button
                                onClick={() => setAction('purge')}
                                className="px-4 py-2 bg-white border border-red-200 text-red-600 rounded-lg flex items-center gap-2 hover:bg-red-50 transition-colors font-medium"
                            >
                                <Trash2 size={16} /> Delete ({selectedIds.size})
                            </button>
                        </>
                    )}
                </div>
            </div>

            <div className="bg-white rounded-xl shadow border border-slate-200 overflow-hidden">
                {loading ? (
                    <div className="p-8 text-center text-slate-500">Loading...</div>
                ) : items.length === 0 ? (
                    <div className="p-12 text-center text-slate-500 flex flex-col items-center gap-3">
                        <Trash2 size={48} className="text-slate-300" />
                        <p>Recycle bin is empty.</p>
                    </div>
                ) : (
                    <div className="overflow-x-auto">
                        <table className="w-full text-left text-sm">
                            <thead className="bg-slate-50 border-b border-slate-200">
                                <tr>
                                    <th className="p-4 w-10">
                                        <button
                                            onClick={handleSelectAll}
                                            className="text-slate-400 hover:text-slate-600"
                                        >
                                            {selectedIds.size === items.length && items.length > 0 ? <CheckSquare size={18} /> : <Square size={18} />}
                                        </button>
                                    </th>
                                    <th className="p-4 font-semibold text-slate-600">Name</th>
                                    <th className="p-4 font-semibold text-slate-600">Object</th>
                                    <th className="p-4 font-semibold text-slate-600">Deleted Date</th>
                                    <th className="p-4 font-semibold text-slate-600">Deleted By</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-slate-100">
                                {items.map(item => (
                                    <tr key={item[COMMON_FIELDS.ID] as string} className={`hover:bg-slate-50 transition-colors ${selectedIds.has(item[COMMON_FIELDS.RECORD_ID] as string) ? 'bg-blue-50/50' : ''}`}>
                                        <td className="p-4">
                                            <button
                                                onClick={() => handleSelectOne(item[COMMON_FIELDS.RECORD_ID] as string)}
                                                className={`${selectedIds.has(item[COMMON_FIELDS.RECORD_ID] as string) ? 'text-blue-600' : 'text-slate-300 hover:text-slate-400'}`}
                                            >
                                                {selectedIds.has(item[COMMON_FIELDS.RECORD_ID] as string) ? <CheckSquare size={18} /> : <Square size={18} />}
                                            </button>
                                        </td>
                                        <td className="p-4 font-medium text-slate-900">{item.record_name}</td>
                                        <td className="p-4 font-mono text-xs text-slate-500 bg-slate-100 rounded px-2 w-fit">{item[COMMON_FIELDS.OBJECT_API_NAME]}</td>
                                        <td className="p-4 text-slate-500">{new Date(item[COMMON_FIELDS.DELETED_DATE] as string).toLocaleString()}</td>
                                        <td className="p-4 text-slate-500">{item[COMMON_FIELDS.DELETED_BY]}</td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>

            <ConfirmationModal
                isOpen={!!action}
                onClose={() => setAction(null)}
                onConfirm={handleExecute}
                title={action === 'restore' ? 'Restore Records' : 'Permanent Delete'}
                message={action === 'restore'
                    ? `Are you sure you want to restore ${selectedIds.size} selected record(s)?`
                    : `Are you sure you want to permanently delete ${selectedIds.size} record(s)? This cannot be undone.`
                }
                confirmLabel={action === 'restore' ? 'Restore' : 'Delete Forever'}
                cancelLabel="Cancel"
                variant={action === 'purge' ? 'danger' : 'info'}
                loading={processing}
            />
        </div>
    );
};
