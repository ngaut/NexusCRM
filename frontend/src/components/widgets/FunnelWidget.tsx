import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Eye, EyeOff, Loader2 } from 'lucide-react';
import { ResponsiveContainer, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Cell } from 'recharts';
import { WidgetRendererProps, ChartDataEntry } from '../../types';
import { COMMON_FIELDS } from '../../core/constants/CommonFields';
import { dataAPI } from '../../infrastructure/api/data';
import { useRuntime } from '../../contexts/RuntimeContext';
import { UI_DEFAULTS, ROUTES, buildRoute } from '../../core/constants';


export const FunnelWidget: React.FC<WidgetRendererProps> = ({ title, config, data: initialData, loading: initialLoading, isEditing, isVisible, onToggle, globalFilters }) => {
    const [data, setData] = React.useState<ChartDataEntry[]>(Array.isArray(initialData) ? initialData as ChartDataEntry[] : []);
    const [loading, setLoading] = React.useState(initialLoading);
    const navigate = useNavigate();
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
                .then(res => setData(Array.isArray(res) ? res as ChartDataEntry[] : []))
                .catch(err => console.error("Funnel widget error", err))
                .finally(() => setLoading(false));
        } else {
            setData(initialData as ChartDataEntry[]);
        }
    }, [config, globalFilters, initialData, user]);

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

    // Transform logic: Funnel usually expects sorted data. 
    // We assume backend returns grouped data.
    const sortedData = Array.isArray(data) ? [...data].sort((a, b) => b.value - a.value) : [];

    return (
        <div className={`bg-white p-6 rounded-lg border shadow-sm min-h-[350px] flex flex-col relative ${isEditing ? 'border-dashed border-2 border-slate-300' : 'border-slate-200'} h-full`}>
            {isEditing && (
                <button onClick={onToggle} className={`absolute top-2 right-2 p-1.5 rounded-full z-10 ${isVisible ? 'bg-blue-100 text-blue-600' : 'bg-slate-200 text-slate-500'}`}>
                    {isVisible ? <Eye size={16} /> : <EyeOff size={16} />}
                </button>
            )}
            <h3 className="text-lg font-bold text-slate-800 mb-6">{title}</h3>
            <div className="flex-1 w-full relative min-h-[250px]">
                {loading ? (
                    <div className="w-full h-full flex items-center justify-center text-slate-400"><Loader2 className="animate-spin mr-2" /> Loading...</div>
                ) : (
                    <ResponsiveContainer width="100%" height="100%">
                        <BarChart
                            layout="vertical"
                            data={sortedData}
                            margin={{ top: 20, right: 30, left: 40, bottom: 5 }}
                        >
                            <CartesianGrid strokeDasharray="3 3" horizontal={true} vertical={false} />
                            <XAxis type="number" hide />
                            <YAxis dataKey="name" type="category" width={100} />
                            <Tooltip cursor={{ fill: 'transparent' }} />
                            <Bar
                                dataKey="value"
                                barSize={30}
                                radius={[0, 4, 4, 0]}
                                onClick={(data) => handleDrillDown(data as unknown as ChartDataEntry)}
                                cursor="pointer"
                            >
                                {sortedData.map((entry: ChartDataEntry, index: number) => (
                                    <Cell key={`cell-${index}`} fill={UI_DEFAULTS.CHART_COLORS[index % UI_DEFAULTS.CHART_COLORS.length]} cursor="pointer" />
                                ))}
                            </Bar>
                        </BarChart>
                    </ResponsiveContainer>
                )}
            </div>
            <p className="text-center text-xs text-slate-400 mt-2">Funnel Visualization</p>
        </div>
    );
};
