import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { Edit3, X } from 'lucide-react';
import { ObjectMetadata, ListView, SObject } from '../types';
import { dataAPI } from '../infrastructure/api/data';
import { metadataAPI } from '../infrastructure/api/metadata';
import { Button } from './ui/Button';
import { useToast, useErrorToast, useSuccessToast } from './ui/Toast';
import { formatApiError, getOperationErrorMessage } from '../core/utils/errorHandling';
import { usePermissions } from '../contexts/PermissionContext';
import { CreateObjectWizard } from './modals/CreateObjectWizard';
import { BulkEditModal } from './modals/BulkEditModal';
import { SYSTEM_TABLE_NAMES } from '../generated-schema';
import { SYSTEM_FIELDS, COMMON_FIELDS } from '../core/constants/CommonFields';
import { ROUTES } from '../core/constants/Routes';
import { UI_TIMING } from '../core/constants';
import { KanbanBoard } from './KanbanBoard';
import { ListViewCharts } from './ListViewCharts';
import { useSplitView } from './SplitViewContainer';
import { MetadataRecordDetail } from './MetadataRecordDetail';
import { RecordListHeader } from './record-list/RecordListHeader';
import { RecordListTable } from './record-list/RecordListTable';

interface MetadataRecordListProps {
    objectMetadata: ObjectMetadata;
    filterExpr?: string;
    searchTerm?: string;
    onRecordClick?: (recordId: string) => void;
    onCreateNew?: () => void;
    customActions?: Array<{
        label: string;
        icon?: React.ReactNode;
        onClick: (selectedRecords: string[]) => void;
    }>;
}

export function MetadataRecordList({
    objectMetadata,
    filterExpr,
    searchTerm = '',
    onRecordClick,
    onCreateNew,
    customActions = [],
}: MetadataRecordListProps) {
    const navigate = useNavigate();
    const showError = useErrorToast();
    const { hasObjectPermission } = usePermissions();

    const [records, setRecords] = useState<SObject[]>([]);
    const [loading, setLoading] = useState(true);
    const [selectedRecords, setSelectedRecords] = useState<Set<string>>(new Set());
    const [sortField, setSortField] = useState<string>('');
    const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('asc');
    const [viewMode] = useState<'list' | 'kanban'>('list'); // setViewMode used to differ list/kanban, defaulting to list for now based on prev usage or could be prop
    const [showCharts] = useState(true);

    // Internal search state
    const [localSearch, setLocalSearch] = useState('');
    const [debouncedSearch, setDebouncedSearch] = useState('');
    const [createObjectWizardOpen, setCreateObjectWizardOpen] = useState(false);
    const [bulkEditModalOpen, setBulkEditModalOpen] = useState(false);

    // List View state
    const [listViews, setListViews] = useState<ListView[]>([]);
    const [selectedView, setSelectedView] = useState<ListView | null>(null);
    const [viewDropdownOpen, setViewDropdownOpen] = useState(false);
    const [saveViewModalOpen, setSaveViewModalOpen] = useState(false);
    const [newViewName, setNewViewName] = useState('');
    const showSuccess = useSuccessToast();

    // Split View state
    const {
        selectedRecordId: splitRecordId,
        isSplitMode,
        openInSplit,
        closeSplit,
    } = useSplitView();

    // Active Filters State
    const [activeFilterExpr, setActiveFilterExpr] = useState<string>(filterExpr || '');

    // Initialize/Update activeFilterExpr from props
    useEffect(() => {
        if (filterExpr) {
            setActiveFilterExpr(filterExpr);
        }
    }, [filterExpr]);

    // Debounce search input
    useEffect(() => {
        const timer = setTimeout(() => {
            setDebouncedSearch(localSearch);
        }, UI_TIMING.DEBOUNCE_FAST_MS);
        return () => clearTimeout(timer);
    }, [localSearch]);

    // Effective search term (prop overrides local)
    const effectiveSearch = searchTerm || debouncedSearch;

    // Load list views
    useEffect(() => {
        if (objectMetadata?.api_name) {
            metadataAPI.getListViews(objectMetadata.api_name)
                .then(res => setListViews(res.views || []))
                .catch(() => setListViews([]));
        }
    }, [objectMetadata?.api_name]);

    // Determine which fields to display in list view
    const displayFields = useMemo(() => {
        const fields = objectMetadata.fields || [];
        if (objectMetadata.list_fields && objectMetadata.list_fields.length > 0) {
            return fields.filter(f =>
                objectMetadata.list_fields!.includes(f.api_name)
            );
        }
        return fields
            .filter(f => !f.is_system || f.api_name === COMMON_FIELDS.NAME)
            .slice(0, 6);
    }, [objectMetadata]);

    // Load records
    const loadRecords = useCallback(async () => {
        setLoading(true);

        try {
            const parts: string[] = [];

            if (activeFilterExpr) {
                parts.push(`(${activeFilterExpr})`);
            }

            if (effectiveSearch) {
                const fields = objectMetadata.fields || [];
                const nameField = fields.find(f => f.is_name_field) ||
                    fields.find(f => f.api_name === COMMON_FIELDS.NAME) ||
                    fields.find(f => f.type === 'Text');

                if (nameField) {
                    parts.push(`contains(${nameField.api_name}, '${effectiveSearch}')`);
                }
            }

            const finalFilterExpr = parts.length > 0 ? parts.join(' && ') : undefined;

            const data = await dataAPI.query({
                objectApiName: objectMetadata.api_name === SYSTEM_TABLE_NAMES.SYSTEM_OBJECT ? SYSTEM_TABLE_NAMES.SYSTEM_OBJECT : objectMetadata.api_name,
                filterExpr: finalFilterExpr,
                sortField,
                sortDirection: sortDirection.toUpperCase(),
                limit: 100
            });
            setRecords(data);
        } catch (err) {
            const apiError = formatApiError(err);
            showError(
                getOperationErrorMessage('fetch', objectMetadata.label || objectMetadata.api_name, apiError)
            );
        } finally {
            setLoading(false);
        }
    }, [objectMetadata.api_name, objectMetadata.fields, objectMetadata.label, activeFilterExpr, sortField, sortDirection, effectiveSearch, showError]);

    const filtersKey = activeFilterExpr;

    useEffect(() => {
        loadRecords();
    }, [loadRecords]);

    // Handle bulk selection
    const toggleSelectAll = () => {
        if (selectedRecords.size === records.length) {
            setSelectedRecords(new Set());
        } else {
            setSelectedRecords(new Set(records.map(r => r[COMMON_FIELDS.ID] as string)));
        }
    };

    const toggleSelectRecord = (recordId: string) => {
        const newSelection = new Set(selectedRecords);
        if (newSelection.has(recordId)) {
            newSelection.delete(recordId);
        } else {
            newSelection.add(recordId);
        }
        setSelectedRecords(newSelection);
    };

    // Handle record click
    const handleRecordClick = (record: SObject) => {
        const recordId = record[COMMON_FIELDS.ID] as string;
        if (onRecordClick) {
            onRecordClick(recordId);
        } else if (isSplitMode) {
            openInSplit(recordId);
        } else {
            navigate(ROUTES.OBJECT.DETAIL(objectMetadata.api_name, recordId));
        }
    };

    const handleCreateNew = () => {
        if (onCreateNew) {
            onCreateNew();
        } else {
            setCreateObjectWizardOpen(true);
        }
    };

    const handleSort = (field: string) => {
        if (sortField === field) {
            setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
        } else {
            setSortField(field);
            setSortDirection('asc');
        }
    };

    const handleNewView = () => {
        setNewViewName('');
        // If we want "New" to start fresh, we could clear filters. 
        // But usually users want to save current state as new.
        // If "current state" is "All Records" (no filter), creating a new view means 
        // creating a view effectively equal to "All Records" until modified.
        // Let's assume "New" means "Save current state as new view".
        // If user wants a blank view, they clear filters then click New.
        setSaveViewModalOpen(true);
    };

    const handleCloneView = () => {
        if (selectedView) {
            setNewViewName(`Copy of ${selectedView.label}`);
            setSaveViewModalOpen(true);
        }
    };

    const handleDeleteView = async (viewId: string) => {
        if (!confirm('Are you sure you want to delete this list view?')) return;
        try {
            await metadataAPI.deleteListView(viewId);
            setListViews(listViews.filter(v => v.id !== viewId));
            if (selectedView?.id === viewId) {
                setSelectedView(null);
                setActiveFilterExpr('');
            }
            showSuccess('List view deleted');
        } catch (err) {
            showError(formatApiError(err).message);
        }
    };

    return (
        <>
            {showCharts && (
                <ListViewCharts
                    objectMetadata={objectMetadata}
                    filterExpr={activeFilterExpr}
                />
            )}

            <div className="space-y-4">
                <RecordListHeader
                    objectMetadata={objectMetadata}
                    recordsCount={records.length}
                    localSearch={localSearch}
                    setLocalSearch={setLocalSearch}
                    onRefresh={loadRecords}
                    activeFilterExpr={activeFilterExpr}
                    setActiveFilterExpr={setActiveFilterExpr}
                    customActions={customActions}
                    selectedRecordsCount={selectedRecords.size}
                    onCustomActionClick={(action) => action.onClick(Array.from(selectedRecords))}
                    onCreateNew={handleCreateNew}
                    viewDropdownOpen={viewDropdownOpen}
                    setViewDropdownOpen={setViewDropdownOpen}
                    selectedView={selectedView}
                    setSelectedView={setSelectedView}
                    listViews={listViews}
                    onSaveViewClick={() => { setNewViewName(''); setSaveViewModalOpen(true); setViewDropdownOpen(false); }}
                    onCloneViewClick={handleCloneView}
                    onNewViewClick={handleNewView}
                    onDeleteViewClick={handleDeleteView}
                />

                {/* Bulk Actions Bar */}
                {selectedRecords.size > 0 && (
                    <div className="bg-blue-50 border border-blue-200 rounded-lg px-4 py-3 flex items-center justify-between">
                        <span className="text-sm font-medium text-blue-900">
                            {selectedRecords.size} {selectedRecords.size === 1 ? 'record' : 'records'} selected
                        </span>
                        <div className="flex items-center gap-2">
                            {hasObjectPermission(objectMetadata.api_name, 'edit') && (
                                <Button
                                    variant="secondary"
                                    size="sm"
                                    icon={<Edit3 className="w-4 h-4" />}
                                    onClick={() => setBulkEditModalOpen(true)}
                                >
                                    Edit Selected
                                </Button>
                            )}
                            <Button variant="ghost" size="sm" onClick={() => setSelectedRecords(new Set())}>
                                Deselect All
                            </Button>
                        </div>
                    </div>
                )}

                {/* Content: List or Board */}
                {viewMode === 'list' ? (
                    <RecordListTable
                        records={records}
                        displayFields={displayFields}
                        selectedRecords={selectedRecords}
                        toggleSelectAll={toggleSelectAll}
                        toggleSelectRecord={toggleSelectRecord}
                        sortField={sortField}
                        sortDirection={sortDirection}
                        onSort={handleSort}
                        handleRecordClick={handleRecordClick}
                        objectMetadata={objectMetadata}
                        onNavigate={(obj, id) => {
                            if (onRecordClick) {
                                onRecordClick(id);
                            } else if (isSplitMode) {
                                openInSplit(id);
                            } else {
                                navigate(ROUTES.OBJECT.DETAIL(obj || objectMetadata.api_name, id));
                            }
                        }}
                        onCreateNew={handleCreateNew}
                    />
                ) : (
                    <div className="flex-1 min-h-[500px] overflow-hidden">
                        <KanbanBoard
                            objectMetadata={objectMetadata}
                            records={records}
                            onRecordClick={(r) => handleRecordClick(r)}
                            onUpdateRecord={async (id, updates) => {
                                try {
                                    await dataAPI.updateRecord(objectMetadata.api_name, id, updates);
                                    showSuccess('Record updated');
                                    loadRecords();
                                } catch (err) {
                                    showError(formatApiError(err).message);
                                }
                            }}
                            loading={loading}
                        />
                    </div>
                )}
            </div>

            {/* Save View Modal */}
            {saveViewModalOpen && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-[100]">
                    <div className="bg-white rounded-lg shadow-xl p-6 w-96">
                        <h3 className="text-lg font-semibold mb-4">Save List View</h3>
                        <input
                            type="text"
                            value={newViewName}
                            onChange={(e) => setNewViewName(e.target.value)}
                            placeholder="View name"
                            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 mb-4"
                            autoFocus
                        />
                        <div className="flex justify-end gap-2">
                            <Button variant="ghost" onClick={() => { setSaveViewModalOpen(false); setNewViewName(''); }}>
                                Cancel
                            </Button>
                            <Button
                                variant="primary"
                                onClick={async () => {
                                    if (!newViewName.trim()) return;
                                    try {
                                        const res = await metadataAPI.createListView({
                                            [COMMON_FIELDS.OBJECT_API_NAME]: objectMetadata.api_name,
                                            [COMMON_FIELDS.LABEL]: newViewName.trim(),
                                            filter_expr: activeFilterExpr,
                                            fields: displayFields.map(f => f.api_name),
                                        });
                                        setListViews([...listViews, res.view]);
                                        setSelectedView(res.view);
                                        showSuccess('List view saved!');
                                        setSaveViewModalOpen(false);
                                        setNewViewName('');
                                    } catch (err) {
                                        showError(formatApiError(err).message);
                                    }
                                }}
                                disabled={!newViewName.trim()}
                            >
                                Save
                            </Button>
                        </div>
                    </div>
                </div >
            )
            }

            {/* Object Creation Wizard */}
            <CreateObjectWizard
                isOpen={createObjectWizardOpen}
                onClose={() => setCreateObjectWizardOpen(false)}
                onSuccess={(objectId) => {
                    loadRecords();
                    navigate(ROUTES.OBJECT.DETAIL(SYSTEM_TABLE_NAMES.SYSTEM_OBJECT, objectId));
                }}
            />

            {/* Bulk Edit Modal */}
            <BulkEditModal
                isOpen={bulkEditModalOpen}
                onClose={() => setBulkEditModalOpen(false)}
                objectMetadata={objectMetadata}
                selectedRecordIds={Array.from(selectedRecords)}
                onSuccess={() => {
                    loadRecords();
                    setSelectedRecords(new Set());
                }}
            />

            {/* Split View Panel */}
            {
                isSplitMode && splitRecordId && (
                    <div className="fixed inset-0 z-[100] bg-black/20" onClick={closeSplit}>
                        <div
                            className="absolute right-0 top-0 bottom-0 w-1/2 bg-white shadow-2xl overflow-auto"
                            onClick={(e) => e.stopPropagation()}
                        >
                            <div className="sticky top-0 z-10 bg-white border-b border-gray-200 px-4 py-3 flex items-center justify-between">
                                <h3 className="font-semibold text-gray-900">Record Detail</h3>
                                <button
                                    onClick={closeSplit}
                                    className="p-1 rounded hover:bg-gray-100"
                                >
                                    <X className="w-5 h-5 text-gray-500" />
                                </button>
                            </div>
                            <div className="p-4">
                                <MetadataRecordDetail
                                    objectMetadata={objectMetadata}
                                    recordId={splitRecordId}
                                    onBack={closeSplit}
                                />
                            </div>
                        </div>
                    </div>
                )
            }
        </>
    );
}
