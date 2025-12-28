import React, { useState, useEffect, useMemo } from 'react';
import { BarChart2, TrendingUp, DollarSign, Hash, PieChart, ChevronDown, ChevronUp, Loader2 } from 'lucide-react';
import { ObjectMetadata, FieldMetadata } from '../types';
import { dataAPI } from '../infrastructure/api/data';
import { UI_DEFAULTS } from '../core/constants/ApplicationDefaults';

interface ListViewChartsProps {
    objectMetadata: ObjectMetadata;
    filterExpr?: string;
}

interface QuickStat {
    label: string;
    value: string | number;
    icon: React.ReactNode;
    color: string;
    trend?: number;
}

/**
 * ListViewCharts - Quick analytics panel above record lists
 * Shows aggregate stats: count, sums of currency fields, distributions
 */
export function ListViewCharts({ objectMetadata, filterExpr }: ListViewChartsProps) {
    const [isExpanded, setIsExpanded] = useState(true);
    const [loading, setLoading] = useState(true);
    const [stats, setStats] = useState<QuickStat[]>([]);
    const [distribution, setDistribution] = useState<{ label: string; value: number; color: string }[]>([]);
    const [distributionField, setDistributionField] = useState<string>('');

    // Find currency and number fields for aggregation
    const currencyFields = useMemo(() =>
        objectMetadata.fields?.filter(f =>
            f.type === 'Currency' || f.type === 'Number' || f.type === 'Decimal'
        ) || []
        , [objectMetadata.fields]);

    // Find picklist fields for distribution
    const picklistFields = useMemo(() =>
        objectMetadata.fields?.filter(f =>
            f.type === 'Picklist' || f.type === 'Status' || f.api_name === 'status'
        ) || []
        , [objectMetadata.fields]);

    // Set default distribution field
    useEffect(() => {
        if (picklistFields.length > 0 && !distributionField) {
            // Prefer 'status' or 'stage' fields
            const preferred = picklistFields.find(f =>
                f.api_name === 'status' || f.api_name === 'stage'
            );
            setDistributionField(preferred?.api_name || picklistFields[0]?.api_name || '');
        }
    }, [picklistFields, distributionField]);

    // Fetch aggregate stats
    useEffect(() => {
        const fetchStats = async () => {
            if (!objectMetadata.api_name) return;

            setLoading(true);
            try {
                // Fetch all records for aggregation
                const records = await dataAPI.query({
                    objectApiName: objectMetadata.api_name,
                    filterExpr,
                    limit: 10000 // Get enough for meaningful stats
                });

                const newStats: QuickStat[] = [];

                // Total Record Count
                newStats.push({
                    label: 'Total Records',
                    value: records.length.toLocaleString(),
                    icon: <Hash size={20} />,
                    color: 'blue'
                });

                // Sum/Avg of currency fields (take first 2)
                currencyFields.slice(0, 2).forEach(field => {
                    const sum = records.reduce((acc, r) => {
                        const val = parseFloat(String(r[field.api_name])) || 0;
                        return acc + val;
                    }, 0);

                    const displayValue = field.type === 'Currency'
                        ? `$${sum.toLocaleString(undefined, { minimumFractionDigits: 0, maximumFractionDigits: 0 })}`
                        : sum.toLocaleString(undefined, { maximumFractionDigits: 2 });

                    newStats.push({
                        label: `Total ${field.label}`,
                        value: displayValue,
                        icon: field.type === 'Currency' ? <DollarSign size={20} /> : <TrendingUp size={20} />,
                        color: 'emerald'
                    });
                });

                // Calculate distribution for picklist field
                if (distributionField) {
                    const counts: Record<string, number> = {};
                    records.forEach(r => {
                        const val = String(r[distributionField] || 'Unset');
                        counts[val] = (counts[val] || 0) + 1;
                    });

                    const colors = UI_DEFAULTS.CHART_COLORS;
                    const dist = Object.entries(counts)
                        .sort((a, b) => b[1] - a[1])
                        .map(([label, value], i) => ({
                            label,
                            value,
                            color: colors[i % colors.length]
                        }));

                    setDistribution(dist);
                }

                setStats(newStats);
            } catch {
                // Stats fetch failure - handled via empty state
            } finally {
                setLoading(false);
            }
        };

        fetchStats();
    }, [objectMetadata.api_name, filterExpr, currencyFields, distributionField]);

    // Don't show if object has no useful data fields
    if (currencyFields.length === 0 && picklistFields.length === 0) {
        return null;
    }

    const totalDistribution = distribution.reduce((sum, d) => sum + d.value, 0);

    return (
        <div className="bg-white border border-slate-200 rounded-xl shadow-sm mb-6 overflow-hidden">
            {/* Header */}
            <button
                onClick={() => setIsExpanded(!isExpanded)}
                className="w-full px-5 py-3 flex items-center justify-between hover:bg-slate-50 transition-colors"
            >
                <div className="flex items-center gap-3">
                    <div className="p-2 bg-blue-50 rounded-lg">
                        <BarChart2 size={18} className="text-blue-600" />
                    </div>
                    <span className="font-semibold text-slate-800">Quick Analytics</span>
                    {loading && <Loader2 size={16} className="animate-spin text-slate-400" />}
                </div>
                {isExpanded ? (
                    <ChevronUp size={20} className="text-slate-400" />
                ) : (
                    <ChevronDown size={20} className="text-slate-400" />
                )}
            </button>

            {/* Content */}
            {isExpanded && (
                <div className="px-5 pb-5 pt-2">
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                        {/* Stat Cards */}
                        {stats.map((stat, i) => (
                            <div
                                key={i}
                                className={`p-4 rounded-xl bg-gradient-to-br ${stat.color === 'blue' ? 'from-blue-50 to-blue-100/50' :
                                    stat.color === 'emerald' ? 'from-emerald-50 to-emerald-100/50' :
                                        'from-slate-50 to-slate-100/50'
                                    }`}
                            >
                                <div className="flex items-center justify-between mb-2">
                                    <span className="text-sm font-medium text-slate-600">{stat.label}</span>
                                    <div className={`${stat.color === 'blue' ? 'text-blue-600' :
                                        stat.color === 'emerald' ? 'text-emerald-600' :
                                            'text-slate-600'
                                        }`}>
                                        {stat.icon}
                                    </div>
                                </div>
                                <div className="text-2xl font-bold text-slate-900">{stat.value}</div>
                            </div>
                        ))}

                        {/* Distribution Chart (Mini Horizontal Bars) */}
                        {distribution.length > 0 && (
                            <div className="p-4 rounded-xl bg-gradient-to-br from-violet-50 to-violet-100/50 col-span-1 lg:col-span-1">
                                <div className="flex items-center justify-between mb-3">
                                    <span className="text-sm font-medium text-slate-600">
                                        By {picklistFields.find(f => f.api_name === distributionField)?.label || distributionField}
                                    </span>
                                    <PieChart size={18} className="text-violet-600" />
                                </div>
                                <div className="space-y-2">
                                    {distribution.slice(0, 4).map((d, i) => (
                                        <div key={i} className="flex items-center gap-2">
                                            <div
                                                className="h-2 rounded-full"
                                                style={{
                                                    width: `${Math.max((d.value / totalDistribution) * 100, 5)}%`,
                                                    backgroundColor: d.color,
                                                    minWidth: '8px'
                                                }}
                                            />
                                            <span className="text-xs text-slate-600 truncate flex-1">{d.label}</span>
                                            <span className="text-xs font-semibold text-slate-700">{d.value}</span>
                                        </div>
                                    ))}
                                    {distribution.length > 4 && (
                                        <div className="text-xs text-slate-400">+{distribution.length - 4} more</div>
                                    )}
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            )}
        </div>
    );
}
