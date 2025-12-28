import React from 'react';
import { Eye, EyeOff, Loader2 } from 'lucide-react';
import { ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';
import { WidgetRendererProps, ChartDataEntry } from '../../types';
import { dataAPI } from '../../infrastructure/api/data';
import { FieldCreatedDate, FieldOwnerID } from '../../constants';

export const GaugeWidget: React.FC<WidgetRendererProps> = ({ title, config, data: initialData, loading: initialLoading, isEditing, isVisible, onToggle, globalFilters }) => {
    const [data, setData] = React.useState<number | ChartDataEntry[]>(initialData as number | ChartDataEntry[]);
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
                .then(res => setData(res as number | ChartDataEntry[]))
                .catch(err => console.error("Gauge widget error", err))
                .finally(() => setLoading(false));
        } else {
            setData(initialData as number | ChartDataEntry[]);
        }
    }, [config, globalFilters, initialData]);

    const val = typeof data === 'number' ? data : (Array.isArray(data) && (data as ChartDataEntry[])[0]?.value) || 0;
    const max = 100;
    const percent = Math.min(100, Math.max(0, (val / max) * 100));

    const chartData = [
        { name: 'Value', value: val },
        { name: 'Remaining', value: max - val }
    ];

    return (
        <div className={`bg-white p-6 rounded-lg border shadow-sm min-h-[350px] flex flex-col relative ${isEditing ? 'border-dashed border-2 border-slate-300' : 'border-slate-200'} h-full`}>
            {isEditing && (
                <button onClick={onToggle} className={`absolute top-2 right-2 p-1.5 rounded-full z-10 ${isVisible ? 'bg-blue-100 text-blue-600' : 'bg-slate-200 text-slate-500'}`}>
                    {isVisible ? <Eye size={16} /> : <EyeOff size={16} />}
                </button>
            )}
            <h3 className="text-lg font-bold text-slate-800 mb-2">{title}</h3>
            <div className="flex-1 w-full relative min-h-[200px] flex flex-col items-center justify-center">
                {loading ? (
                    <div className="w-full h-full flex items-center justify-center text-slate-400"><Loader2 className="animate-spin mr-2" /> Loading...</div>
                ) : (
                    <>
                        <ResponsiveContainer width="100%" height={200}>
                            <PieChart>
                                <Pie
                                    data={chartData}
                                    cx="50%" cy="100%"
                                    startAngle={180}
                                    endAngle={0}
                                    innerRadius={80}
                                    outerRadius={120}
                                    paddingAngle={0}
                                    dataKey="value"
                                >
                                    <Cell key="val" fill={config.color ? `var(--color-${config.color}-500)` : '#10b981'} />
                                    <Cell key="rem" fill="#e2e8f0" />
                                </Pie>
                            </PieChart>
                        </ResponsiveContainer>
                        <div className="text-center -mt-10">
                            <span className="text-4xl font-bold text-slate-900">{val}</span>
                            <p className="text-sm text-slate-500">Goal: {max}</p>
                        </div>
                    </>
                )}
            </div>
        </div>
    );
};
