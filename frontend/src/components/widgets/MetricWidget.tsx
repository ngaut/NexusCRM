import React from 'react';
import * as Icons from 'lucide-react';
import { Box, Eye, EyeOff } from 'lucide-react';
import { WidgetRendererProps } from '../../types';
import { dataAPI } from '../../infrastructure/api/data';
import { FieldCreatedDate, FieldOwnerID } from '../../constants';

export const MetricWidget: React.FC<WidgetRendererProps> = ({ title, config, data: initialData, loading: initialLoading, isEditing, isVisible, onToggle, globalFilters }) => {
    const Icon = config.icon ? (Icons as unknown as Record<string, React.ComponentType<{ size: number; className?: string }>>)[config.icon] : Box;
    const color = config.color || 'blue';
    const colorClass = `text-${color}-600`;
    const bgClass = `bg-${color}-100`;

    const [data, setData] = React.useState(initialData);
    const [loading, setLoading] = React.useState(initialLoading);

    React.useEffect(() => {
        if (config.query) {
            setLoading(true);
            const queryWithFilters = { ...config.query };

            // Apply Global Filters
            if (globalFilters) {
                const parts: string[] = [];
                if (queryWithFilters.filterExpr) parts.push(`(${queryWithFilters.filterExpr})`);

                if (globalFilters.ownerId) {
                    parts.push(`${FieldOwnerID} == '${globalFilters.ownerId}'`);
                }
                if (globalFilters.startDate) {
                    parts.push(`${FieldCreatedDate} >= '${globalFilters.startDate}'`);
                }
                if (globalFilters.endDate) {
                    parts.push(`${FieldCreatedDate} <= '${globalFilters.endDate}'`);
                }
                if (parts.length > 0) {
                    queryWithFilters.filterExpr = parts.join(' && ');
                }
            }

            dataAPI.runAnalytics(queryWithFilters)
                .then(result => setData(result))
                .catch(console.error)
                .finally(() => setLoading(false));
        } else {
            setData(initialData);
        }
    }, [config, globalFilters, initialData]);

    return (
        <div className={`relative bg-white p-6 rounded-lg border shadow-sm transition-all h-full flex flex-col justify-between ${isEditing ? 'border-dashed border-2 border-slate-300' : 'border-slate-200'} ${!isVisible ? 'opacity-40' : ''}`}>
            {isEditing && (
                <button onClick={onToggle} className={`absolute top-2 right-2 p-1.5 rounded-full z-10 ${isVisible ? 'bg-blue-100 text-blue-600' : 'bg-slate-200 text-slate-500'}`}>
                    {isVisible ? <Eye size={16} /> : <EyeOff size={16} />}
                </button>
            )}
            <div className="flex items-center justify-between mb-4">
                <p className="text-sm font-medium text-slate-500 uppercase tracking-wide">{title}</p>
                <div className={`p-2 rounded-lg ${bgClass}`}>
                    <Icon size={20} className={colorClass} />
                </div>
            </div>
            <div>
                {loading ? <div className="h-8 w-16 bg-slate-100 animate-pulse rounded"></div> : <h3 className="text-3xl font-bold text-slate-900">{data === null || data === undefined ? '--' : (typeof data === 'number' ? data.toLocaleString() : String(data))}</h3>}
            </div>
        </div>
    );
};
