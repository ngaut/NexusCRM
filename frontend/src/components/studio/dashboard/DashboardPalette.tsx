import React from 'react';
import {
    Layout,
    Type,
    Image,
    BarChart2,
    PieChart,
    Activity,
    TrendingUp,
    List,
    Divide,

    Columns,
    Code
} from 'lucide-react';

interface PaletteItemProps {
    type: string;
    label: string;
    icon: React.ReactNode;
    onDragStart: (e: React.DragEvent, type: string) => void;
    onClick: (type: string) => void;
}

const PaletteItem: React.FC<PaletteItemProps> = ({ type, label, icon, onDragStart, onClick }) => {
    return (
        <div
            draggable={true}
            onDragStart={(e) => onDragStart(e, type)}
            onClick={() => onClick(type)}
            className="flex items-center gap-3 p-3 bg-white border border-slate-200 rounded-lg cursor-grab hover:border-blue-500 hover:shadow-md hover:bg-blue-50 transition-all mb-2 select-none active:scale-95"
            role="button"
            tabIndex={0}
        >
            <div className="text-slate-500">
                {icon}
            </div>
            <span className="text-sm font-medium text-slate-700">{label}</span>
        </div>
    );
};

export const DashboardPalette: React.FC<{ onAddWidget: (type: string) => void }> = ({ onAddWidget }) => {
    const handleDragStart = (e: React.DragEvent, type: string) => {
        e.dataTransfer.setData('text/plain', type);
        e.dataTransfer.effectAllowed = 'all';
    };

    return (
        <div className="w-64 bg-slate-50 border-r border-slate-200 flex flex-col h-full">
            <div className="p-4 border-b border-slate-200">
                <h3 className="text-sm font-semibold text-slate-900 uppercase tracking-wider">Widgets</h3>
                <p className="text-xs text-slate-400 mt-1">Drag or Click to add</p>
            </div>

            <div className="flex-1 overflow-y-auto p-4">
                <div className="mb-6">
                    <h4 className="text-xs font-medium text-slate-500 mb-3 uppercase">Charts & Data</h4>
                    <PaletteItem
                        type="metric"
                        label="Metric Card"
                        icon={<Activity size={18} />}
                        onDragStart={handleDragStart}
                        onClick={onAddWidget}
                    />
                    <PaletteItem
                        type="chart-bar"
                        label="Bar Chart"
                        icon={<BarChart2 size={18} />}
                        onDragStart={handleDragStart}
                        onClick={onAddWidget}
                    />
                    <PaletteItem
                        type="chart-pie"
                        label="Pie Chart"
                        icon={<PieChart size={18} />}
                        onDragStart={handleDragStart}
                        onClick={onAddWidget}
                    />
                    <PaletteItem
                        type="chart-line"
                        label="Line Chart"
                        icon={<TrendingUp size={18} />}
                        onDragStart={handleDragStart}
                        onClick={onAddWidget}
                    />
                    <PaletteItem
                        type="chart-funnel"
                        label="Funnel"
                        icon={<List size={18} />}
                        onDragStart={handleDragStart}
                        onClick={onAddWidget}
                    />
                </div>

                <div className="mb-6">
                    <h4 className="text-xs font-medium text-slate-500 mb-3 uppercase">Records</h4>
                    <PaletteItem
                        type="record-list"
                        label="Record List"
                        icon={<List size={18} />}
                        onDragStart={handleDragStart}
                        onClick={onAddWidget}
                    />
                    <PaletteItem
                        type="kanban"
                        label="Kanban Board"
                        icon={<Columns size={18} />}
                        onDragStart={handleDragStart}
                        onClick={onAddWidget}
                    />
                    <PaletteItem
                        type="sql-chart"
                        label="SQL Analytics"
                        icon={<Code size={18} />}
                        onDragStart={handleDragStart}
                        onClick={onAddWidget}
                    />
                </div>

                <div>
                    <h4 className="text-xs font-medium text-slate-500 mb-3 uppercase">Layout</h4>
                    <PaletteItem
                        type="text"
                        label="Text Block"
                        icon={<Type size={18} />}
                        onDragStart={handleDragStart}
                        onClick={onAddWidget}
                    />
                    <PaletteItem
                        type="image"
                        label="Image"
                        icon={<Image size={18} />}
                        onDragStart={handleDragStart}
                        onClick={onAddWidget}
                    />
                </div>
            </div>
        </div>
    );
};
