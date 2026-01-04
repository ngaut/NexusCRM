import React from 'react';
import * as Icons from 'lucide-react';
import { Box, Eye, EyeOff } from 'lucide-react';
import { WidgetRendererProps } from '../../types';
import { dataAPI } from '../../infrastructure/api/data';
import { COMMON_FIELDS } from '../../core/constants/CommonFields';
import { useRuntime } from '../../contexts/RuntimeContext';

export const MetricWidget: React.FC<WidgetRendererProps> = ({ title, config, data: initialData, loading: initialLoading, isEditing, isVisible, onToggle, globalFilters }) => {
    const Icon = config.icon ? (Icons as unknown as Record<string, React.ComponentType<{ size: number; className?: string }>>)[config.icon] : Box;
    const color = config.color || 'blue';
    const colorClass = `text-${color}-600`;
    const bgClass = `bg-${color}-100`;

    const [data, setData] = React.useState<number>(typeof initialData === 'number' ? initialData : 0);
    const [loading, setLoading] = React.useState(initialLoading);
    const { user } = useRuntime();

    React.useEffect(() => {
        if (config.query) {
            setLoading(true);
            const queryWithFilters = { ...config.query };
            let filterExpr = queryWithFilters.filter_expr || '';
            const globalFilterParts: string[] = [];

            // Apply Global Filters
            if (globalFilters) {
                if (filterExpr) globalFilterParts.push(`(${filterExpr})`);

                if (globalFilters.ownerId) {
                    globalFilterParts.push(`${COMMON_FIELDS.OWNER_ID} == '${globalFilters.ownerId}'`);
                }
                if (globalFilters.startDate) {
                    globalFilterParts.push(`${COMMON_FIELDS.CREATED_DATE} >= '${globalFilters.startDate}'`);
                }
                if (globalFilters.endDate) {
                    globalFilterParts.push(`${COMMON_FIELDS.CREATED_DATE} <= '${globalFilters.endDate}'`);
                }
            }
            if (globalFilterParts.length > 0) {
                filterExpr = globalFilterParts.join(' && ');
            }

            // Apply scope filter
            if (config.scope === 'mine' && user?.id) {
                const ownerCriteria = `${COMMON_FIELDS.OWNER_ID} == '${user.id}'`;
                filterExpr = filterExpr ? `(${filterExpr}) AND (${ownerCriteria})` : ownerCriteria;
            }

            queryWithFilters.filter_expr = filterExpr;

            dataAPI.runAnalytics(queryWithFilters)
                .then(result => setData(Number(result)))
                .catch(err => console.error("Metric widget error", err))
                .finally(() => setLoading(false));
        } else {
            setData(typeof initialData === 'number' ? initialData : 0);
        }
    }, [config, globalFilters, initialData, user]);

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
