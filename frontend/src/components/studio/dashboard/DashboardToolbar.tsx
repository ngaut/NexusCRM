import React from 'react';
import { Save, Layout as LayoutIcon } from 'lucide-react';

interface DashboardToolbarProps {
    title: string;
    setTitle: (title: string) => void;
    onAddWidget: (type: string) => void;
    onDeleteWidget: () => void;
    onSave: () => void;
    saving: boolean;
    hasSelectedWidget: boolean;
}

export const DashboardToolbar: React.FC<DashboardToolbarProps> = ({
    title,
    setTitle,
    onAddWidget,
    onDeleteWidget,
    onSave,
    saving,
    hasSelectedWidget,
}) => {
    return (
        <div className="h-14 bg-white border-b border-slate-200 flex items-center justify-between px-4 z-10 shadow-sm">
            <div className="flex items-center gap-3">
                <LayoutIcon className="text-blue-600" size={20} />
                <input
                    type="text"
                    value={title}
                    onChange={e => setTitle(e.target.value)}
                    placeholder="Dashboard Title"
                    className="font-semibold text-lg bg-transparent border-none focus:ring-0 p-0 text-slate-800 placeholder:text-slate-400"
                />
            </div>
            <div className="flex items-center gap-2">
                <button
                    onClick={() => onAddWidget('metric')}
                    className="flex items-center gap-2 px-3 py-1.5 bg-white border border-slate-300 rounded-lg shadow-sm text-slate-700 hover:bg-slate-50 text-sm font-medium"
                >
                    <LayoutIcon size={16} />
                    Add Metric
                </button>
                <button
                    onClick={() => onAddWidget('chart-bar')}
                    className="flex items-center gap-2 px-3 py-1.5 bg-white border border-slate-300 rounded-lg shadow-sm text-slate-700 hover:bg-slate-50 text-sm font-medium"
                >
                    <LayoutIcon size={16} />
                    Add Bar Chart
                </button>
                {hasSelectedWidget && (
                    <button
                        onClick={onDeleteWidget}
                        className="text-red-600 hover:bg-red-50 px-3 py-1.5 rounded text-sm font-medium transition-colors"
                    >
                        Delete Widget
                    </button>
                )}
                <button
                    onClick={onSave}
                    disabled={saving}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium shadow-sm transition-colors disabled:opacity-50"
                >
                    <Save size={16} />
                    {saving ? 'Saving...' : 'Save'}
                </button>
            </div>
        </div>
    );
};
