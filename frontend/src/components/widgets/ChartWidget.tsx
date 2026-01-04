import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Eye, EyeOff, Loader2 } from 'lucide-react';
import { ResponsiveContainer, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, PieChart, Pie, Cell } from 'recharts';
import { WidgetRendererProps, ChartDataEntry } from '../../types';
import { COMMON_FIELDS } from '../../core/constants/CommonFields';
import { UI_DEFAULTS, ROUTES, buildRoute } from '../../core/constants';
import { dataAPI } from '../../infrastructure/api/data';

export const ChartWidget: React.FC<WidgetRendererProps> = ({ title, config, data: initialData, loading: initialLoading, isEditing, isVisible, onToggle, globalFilters }) => {
    const [data, setData] = React.useState<ChartDataEntry[]>(Array.isArray(initialData) ? initialData as ChartDataEntry[] : []);
    const [loading, setLoading] = React.useState(initialLoading);
    const navigate = useNavigate();

    React.useEffect(() => {
        if (config.query) {
            setLoading(true);
            const queryWithFilters = { ...config.query };

            // Apply Global Filters
            if (globalFilters) {
                const parts: string[] = [];
                if (queryWithFilters.filter_expr) parts.push(`(${queryWithFilters.filter_expr})`);

                if (globalFilters.ownerId) {
                    parts.push(`${COMMON_FIELDS.OWNER_ID} == '${globalFilters.ownerId}'`);
                }
                if (globalFilters.startDate) {
                    parts.push(`${COMMON_FIELDS.CREATED_DATE} >= '${globalFilters.startDate}'`);
                }
                if (globalFilters.endDate) {
                    parts.push(`${COMMON_FIELDS.CREATED_DATE} <= '${globalFilters.endDate}'`);
                }
                if (parts.length > 0) {
                    queryWithFilters.filter_expr = parts.join(' && ');
                }
            }

            dataAPI.runAnalytics(queryWithFilters)
                .then(res => setData(Array.isArray(res) ? res as ChartDataEntry[] : []))
                .catch(err => console.error("Chart widget error", err))
                .finally(() => setLoading(false));
        } else {
            setData(initialData as ChartDataEntry[]);
        }
    }, [config, globalFilters, initialData]);

    const handleDrillDown = React.useCallback((entry: ChartDataEntry) => {
        if (!config.query?.object_api_name) return;

        const parts: string[] = [];

        if (globalFilters) {
            if (globalFilters.ownerId) parts.push(`${COMMON_FIELDS.OWNER_ID} == '${globalFilters.ownerId}'`);
            if (globalFilters.startDate) parts.push(`${COMMON_FIELDS.CREATED_DATE} >= '${globalFilters.startDate}'`);
            if (globalFilters.endDate) parts.push(`${COMMON_FIELDS.CREATED_DATE} <= '${globalFilters.endDate}'`);
        }

        // Add grouping filter
        if (config.query.group_by && entry && entry.name) {
            parts.push(`${config.query.group_by} == '${entry.name}'`);
        }

        const filterStr = encodeURIComponent(parts.join(' && '));
        navigate(buildRoute(ROUTES.OBJECT.LIST(config.query.object_api_name), { filterExpr: filterStr }));
    }, [config, globalFilters, navigate]);

    return (
        <div className={`bg-white p-6 rounded-lg border shadow-sm min-h-[350px] flex flex-col relative ${isEditing ? 'border-dashed border-2 border-slate-300' : 'border-slate-200'}`}>
            {isEditing && (
                <button onClick={onToggle} className={`absolute top-2 right-2 p-1.5 rounded-full z-10 ${isVisible ? 'bg-blue-100 text-blue-600' : 'bg-slate-200 text-slate-500'}`}>
                    {isVisible ? <Eye size={16} /> : <EyeOff size={16} />}
                </button>
            )}
            <h3 className="text-lg font-bold text-slate-800 mb-6">{title}</h3>
            <div className="flex-1 w-full relative" style={{ height: 300, minHeight: 300 }}>
                {loading ? (
                    <div className="w-full h-full flex items-center justify-center text-slate-400"><Loader2 className="animate-spin mr-2" /> Loading...</div>
                ) : (!data || (Array.isArray(data) && data.length === 0)) ? (
                    <div className="w-full h-full flex flex-col items-center justify-center text-slate-300">
                        <Loader2 size={32} className="mb-2 opacity-20" /> {/* Reuse loader icon or use something else generic like BarChart2 */}
                        <span className="text-sm font-medium">No Data Available</span>
                    </div>
                ) : (
                    <ResponsiveContainer width="100%" height="100%">
                        {config.type === 'chart-pie' ? (
                            <PieChart>
                                <Pie
                                    data={Array.isArray(data) ? data : []}
                                    cx="50%" cy="50%"
                                    innerRadius={60}
                                    outerRadius={80}
                                    paddingAngle={5}
                                    dataKey="value"
                                    onClick={(data) => handleDrillDown(data as unknown as ChartDataEntry)}
                                    cursor="pointer"
                                >
                                    {(Array.isArray(data) ? data : []).map((entry: ChartDataEntry, index: number) => <Cell key={`cell-${index}`} fill={UI_DEFAULTS.CHART_COLORS[index % UI_DEFAULTS.CHART_COLORS.length]} cursor="pointer" />)}
                                </Pie>
                                <Tooltip />
                            </PieChart>
                        ) : (
                            <BarChart data={Array.isArray(data) ? data : []} margin={{ top: 20, right: 30, left: 20, bottom: 5 }}>
                                <CartesianGrid strokeDasharray="3 3" vertical={false} />
                                <XAxis dataKey="name" axisLine={false} tickLine={false} />
                                <YAxis axisLine={false} tickLine={false} />
                                <Tooltip cursor={{ fill: '#f1f5f9' }} contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }} />
                                <Bar
                                    dataKey="value"
                                    fill={UI_DEFAULTS.CHART_COLORS[0]}
                                    radius={[4, 4, 0, 0]}
                                    barSize={40}
                                    onClick={(data) => handleDrillDown(data as unknown as ChartDataEntry)}
                                    cursor="pointer"
                                >
                                    {(Array.isArray(data) ? data : []).map((entry: ChartDataEntry, index: number) => (
                                        <Cell key={`cell-${index}`} fill={UI_DEFAULTS.CHART_COLORS[index % UI_DEFAULTS.CHART_COLORS.length]} cursor="pointer" />
                                    ))}
                                </Bar>
                            </BarChart>
                        )}
                    </ResponsiveContainer>
                )}
            </div>
        </div>
    );
};
