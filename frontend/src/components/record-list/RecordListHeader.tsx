import React from 'react';
import { Plus, RefreshCw, Search, ChevronDown, Save, List } from 'lucide-react';
import { Button } from '../ui/Button';
import { COMMON_FIELDS } from '../../core/constants';
import { ListFilters } from '../ListFilters';
import { ObjectMetadata, ListView } from '../../types';
import { usePermissions } from '../../contexts/PermissionContext';

export interface CustomAction {
    label: string;
    icon?: React.ReactNode;
    onClick: (selectedRecords: string[]) => void;
}

interface RecordListHeaderProps {
    objectMetadata: ObjectMetadata;
    recordsCount: number;
    localSearch: string;
    setLocalSearch: (value: string) => void;
    onRefresh: () => void;
    activeFilterExpr: string;
    setActiveFilterExpr: (expr: string) => void;
    customActions: CustomAction[];
    selectedRecordsCount: number;
    onCustomActionClick: (action: CustomAction) => void;
    onCreateNew: () => void;
    viewDropdownOpen: boolean;
    setViewDropdownOpen: (open: boolean) => void;
    selectedView: ListView | null;
    setSelectedView: (view: ListView | null) => void;
    listViews: ListView[];
    onSaveViewClick: () => void;
}

export const RecordListHeader: React.FC<RecordListHeaderProps> = ({
    objectMetadata,
    recordsCount,
    localSearch,
    setLocalSearch,
    onRefresh,
    activeFilterExpr,
    setActiveFilterExpr,
    customActions,
    selectedRecordsCount,
    onCustomActionClick,
    onCreateNew,
    viewDropdownOpen,
    setViewDropdownOpen,
    selectedView,
    setSelectedView,
    listViews,
    onSaveViewClick
}) => {
    const { hasObjectPermission } = usePermissions();

    return (
        <div className="flex justify-between items-center">
            <div className="flex items-center gap-2">
                {/* List View Selector */}
                <div className="relative">
                    <button
                        onClick={() => setViewDropdownOpen(!viewDropdownOpen)}
                        className="flex items-center gap-2 px-3 py-1.5 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                        <List size={16} />
                        {selectedView ? selectedView.label : 'All Records'}
                        <ChevronDown size={16} />
                    </button>
                    {viewDropdownOpen && (
                        <div className="absolute left-0 mt-1 w-56 bg-white rounded-md shadow-lg border border-gray-200 z-50">
                            <div className="py-1">
                                <button
                                    onClick={() => {
                                        setSelectedView(null);
                                        setActiveFilterExpr(''); // Reset
                                        setViewDropdownOpen(false);
                                    }}
                                    className={`w-full text-left px-4 py-2 text-sm hover:bg-gray-100 ${!selectedView ? 'bg-blue-50 text-blue-700' : 'text-gray-700'}`}
                                >
                                    All Records
                                </button>
                                {listViews.map(view => (
                                    <button
                                        key={view.id}
                                        onClick={() => {
                                            setSelectedView(view);
                                            const filterExpr = view[COMMON_FIELDS.FILTERS];
                                            if (filterExpr) {
                                                setActiveFilterExpr(filterExpr);
                                            } else {
                                                setActiveFilterExpr('');
                                            }
                                            setViewDropdownOpen(false);
                                        }}
                                        className={`w-full text-left px-4 py-2 text-sm hover:bg-gray-100 ${selectedView?.id === view.id ? 'bg-blue-50 text-blue-700' : 'text-gray-700'}`}
                                    >
                                        {view.label}
                                    </button>
                                ))}
                                {activeFilterExpr && !selectedView && (
                                    <>
                                        <hr className="my-1" />
                                        <button
                                            onClick={onSaveViewClick}
                                            className="w-full text-left px-4 py-2 text-sm text-blue-600 hover:bg-blue-50 flex items-center gap-2"
                                        >
                                            <Save size={14} /> Save Current View
                                        </button>
                                    </>
                                )}
                            </div>
                        </div>
                    )}
                </div>

                <span className="text-sm text-gray-500">({recordsCount})</span>

                {/* List Search Input */}
                <div className="relative ml-4 flex-1 max-w-xs">
                    <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                        <Search size={14} className="text-gray-400" />
                    </div>
                    <input
                        type="text"
                        value={localSearch}
                        onChange={(e) => setLocalSearch(e.target.value)}
                        placeholder={`Search ${objectMetadata.plural_label || (objectMetadata.label + 's')}...`}
                        className="block w-full pl-9 pr-3 py-1.5 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 sm:text-sm"
                    />
                </div>

                <Button
                    variant="ghost"
                    size="sm"
                    icon={<RefreshCw className="w-4 h-4" />}
                    onClick={onRefresh}
                >
                    Refresh
                </Button>

                {/* Filter Component */}
                <ListFilters
                    objectApiName={objectMetadata.api_name}
                    fields={objectMetadata.fields || []}
                    activeFilterExpr={activeFilterExpr}
                    onFiltersChange={(expr) => {
                        setActiveFilterExpr(expr);
                        setSelectedView(null);
                    }}
                />
            </div>

            <div className="flex items-center gap-2">
                {customActions.map((action, idx) => (
                    <Button
                        key={idx}
                        variant="secondary"
                        size="sm"
                        icon={action.icon}
                        onClick={() => onCustomActionClick(action)}
                        disabled={selectedRecordsCount === 0}
                    >
                        {action.label}
                    </Button>
                ))}

                {hasObjectPermission(objectMetadata.api_name, 'create') && (
                    <Button
                        variant="primary"
                        size="sm"
                        icon={<Plus className="w-4 h-4" />}
                        onClick={onCreateNew}
                    >
                        New {objectMetadata.label}
                    </Button>
                )}
            </div>
        </div>
    );
};
