import React, { useState, useEffect } from 'react';
import { WidgetRendererProps } from '../../types';
import { analyticsAPI } from '../../infrastructure/api/analytics';
import { ResponsiveContainer, BarChart, Bar, LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend } from 'recharts';
import { Loader2, AlertTriangle, Code, Play, Download, Database } from 'lucide-react';
import { UI_DEFAULTS } from '../../core/constants';
import { useRuntime } from '../../contexts/RuntimeContext';

export const SqlChartWidget: React.FC<WidgetRendererProps> = ({
    title,
    config,
    isEditing,
    isVisible,
    onToggle,
    globalFilters,
    onConfigUpdate
}) => {
    const { user } = useRuntime();
    const isAdmin = user?.profile_id === 'system_admin' || user?.role_id === 'admin';

    // SQL can be in config.sql (frontend direct) OR config.config.sql (from backend JSON)
    const configuredSql = config.sql || (config as { config?: { sql?: string } }).config?.sql || '';

    const [data, setData] = useState<Record<string, unknown>[]>([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [sqlInput, setSqlInput] = useState(configuredSql);

    // Execute Query Function
    const executeQuery = async () => {
        const sqlToExecute = configuredSql || sqlInput;
        if (!sqlToExecute) return;

        setLoading(true);
        setError(null);

        try {
            // Prepare parameters from global filters
            const params: (string | number | boolean)[] = [];
            // Basic binding logic: If SQL contains @startTime, we inject it?
            // Actually, for MVP backend uses standard '?' or '$1' bind vars?
            // User research said bind vars.
            // Let's assume the user writes "WHERE created_date > @startTime" and we regex replace it?
            // Or better, we pass named params if backend supports it.
            // Our backend `ExecuteRawSQL` takes `[]interface{}`. So it supports positional `?`.
            // We need to parse named params -> positional params here? Or just let user use `?`?
            // "Best practice" research said named args.
            // But `database/sql` uses `?` for standard replacement usually (or `$1` for postgres).
            // Let's stick to `?` for raw SQL to match Go's `sql` package behavior for now.

            // Wait, if we use `?`, we can't easily dynamic bind "startTime" vs "endTime".
            // Let's implement a simple pre-processor here to replace `@startTime` with `?` and add param.

            let finalSql = sqlToExecute;
            const finalParams: (string | number | boolean)[] = [];

            if (globalFilters) {
                // Simple regex replacement for bind variables
                if (finalSql.includes('@start_date')) {
                    finalSql = finalSql.replace(/@start_date/g, '?');
                    finalParams.push(globalFilters.startDate || '1970-01-01');
                }
                if (finalSql.includes('@end_date')) {
                    finalSql = finalSql.replace(/@end_date/g, '?');
                    finalParams.push(globalFilters.endDate || '2099-12-31');
                }
            }


            // Save the SQL back to config if it changed and we have the callback
            if (sqlInput !== config.sql && onConfigUpdate) {
                onConfigUpdate({ sql: sqlInput });
            }

            const result = await analyticsAPI.executeAdminQuery(finalSql, finalParams);
            if (result.success) {
                setData(result.results || []);
            } else {
                setError("Query failed");
            }
        } catch (err: unknown) {
            const message = err instanceof Error ? err.message : "Failed to execute query";
            setError(message);
        } finally {
            setLoading(false);
        }
    };

    // Export data to CSV
    const exportToCsv = () => {
        if (!data || data.length === 0) return;

        // Get headers from first row
        const headers = Object.keys(data[0]);

        // Create CSV content
        const csvRows = [
            headers.join(','),
            ...data.map(row =>
                headers.map(h => {
                    const val = row[h];
                    // Escape quotes and wrap in quotes if contains comma
                    const stringVal = String(val ?? '');
                    if (stringVal.includes(',') || stringVal.includes('"') || stringVal.includes('\n')) {
                        return `"${stringVal.replace(/"/g, '""')}"`;
                    }
                    return stringVal;
                }).join(',')
            )
        ];

        const csvContent = csvRows.join('\n');
        const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
        const url = URL.createObjectURL(blob);

        // Create download link
        const link = document.createElement('a');
        link.href = url;
        link.download = `${title || 'query_results'}_${new Date().toISOString().split('T')[0]}.csv`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        URL.revokeObjectURL(url);
    };

    // Auto-execute in view mode
    useEffect(() => {
        if (!isEditing && configuredSql) {
            executeQuery();
        }
    }, [isEditing, configuredSql, globalFilters]);

    // Handle SQL cleanup on save (passed via config prop update in real app)
    // Here we just local state. The parent DashboardEditor needs to capture `config.sql`.
    // The Registry passes `config` prop. Updating it requires `onUpdate` prop?
    // `WidgetRendererProps` doesn't have `onUpdate`. 
    // The `StudioDashboardEditor` manages state.
    // For now, in "Editing" mode inside the widget, we might not be able to save UP to the parent without a callback.
    // Looking at `MetricWidget`, it doesn't edit its own query.
    // The standard Dashboard editor has a property panel side-bar.
    // THIS is where the "Whole Picture" matters.
    // I cannot put the SQL Input *inside* the widget if the architecture expects editing in a sidebar.
    // However, for this "Advanced" widget, having an inline editor is very "Notebook" style and cool.
    // Let's stick to the plan: Config Mode = Textarea.

    return (
        <div className={`bg-white p-6 rounded-lg border shadow-sm min-h-[400px] flex flex-col relative ${isEditing ? 'border-dashed border-2 border-slate-300' : 'border-slate-200'} h-full`}>
            {isEditing && (
                <div className="absolute top-2 right-2 flex gap-2 z-10">
                    <button onClick={onToggle} className={`p-1.5 rounded-full ${isVisible ? 'bg-blue-100 text-blue-600' : 'bg-slate-200 text-slate-500'}`}>
                        {isVisible ? <Code size={16} /> : <Code size={16} />}
                    </button>
                </div>
            )}

            <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-bold text-slate-800 flex items-center gap-2">
                    {title}
                    {isAdmin && <span className="text-xs font-normal px-2 py-0.5 bg-yellow-100 text-yellow-800 rounded-full">SQL</span>}
                </h3>
                {!isEditing && data.length > 0 && (
                    <button
                        onClick={exportToCsv}
                        className="flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-slate-600 bg-slate-100 rounded-lg hover:bg-slate-200 transition-colors"
                        title="Export to CSV"
                    >
                        <Download size={14} />
                        Export
                    </button>
                )}
            </div>

            {isEditing ? (
                <div className="flex-1 flex flex-col gap-4">
                    <div className="bg-slate-50 p-3 rounded-md border border-slate-200 text-xs">
                        <p className="font-semibold mb-1">Available Bindings:</p>
                        <code className="bg-slate-200 px-1 rounded">@start_date</code>
                        <code className="bg-slate-200 px-1 rounded ml-2">@end_date</code>
                        <code className="bg-slate-200 px-1 rounded ml-2">@user_id</code>
                    </div>
                    <textarea
                        className="w-full flex-1 p-3 font-mono text-sm border border-slate-300 rounded focus:ring-2 focus:ring-blue-500 outline-none resize-none"
                        placeholder="SELECT status, COUNT(*) as value FROM Ticket__c GROUP BY status"
                        value={sqlInput}
                        onChange={(e) => setSqlInput(e.target.value)}
                    />
                    <div className="flex justify-end gap-2">
                        <button
                            onClick={executeQuery}
                            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50"
                            disabled={loading}
                        >
                            {loading ? <Loader2 className="animate-spin" size={16} /> : <Play size={16} />}
                            Run Query
                        </button>
                    </div>
                    {/* Note: Saving typically happens in the parent editor when 'Save Dashboard' is clicked. 
                         We need to ensure `sqlInput` is propagated back to `config.sql`. 
                         BUT `WidgetRendererProps` misses an `onChange` callback. 
                         The current architecture might rely on a side-panel. 
                         I will assume for now editing text here doesn't auto-save to config unless I find a way to bubble it up. 
                         Actually, standard widgets are edited via `StudioDashboardEditor` property panel.
                         I'll need to modify `StudioDashboardEditor` to allow editing `sql` prop if type is sql-chart.
                     */}
                </div>
            ) : (
                <div className="flex-1 min-h-[300px] relative">
                    {error ? (
                        <div className="absolute inset-0 flex items-center justify-center text-red-500 bg-red-50 rounded">
                            <AlertTriangle className="mr-2" /> {error}
                        </div>
                    ) : loading ? (
                        <div className="absolute inset-0 flex items-center justify-center text-slate-400">
                            <Loader2 className="animate-spin mr-2" /> Executing SQL...
                        </div>
                    ) : data.length === 0 ? (
                        <div className="absolute inset-0 flex flex-col items-center justify-center text-slate-400">
                            <Database size={48} className="mb-2 opacity-50" strokeWidth={1.5} />
                            <p className="font-medium">No results found</p>
                        </div>
                    ) : (
                        <ResponsiveContainer width="100%" height="100%">
                            <BarChart data={data}>
                                <CartesianGrid strokeDasharray="3 3" vertical={false} />
                                <XAxis dataKey="name" />
                                <YAxis />
                                <Tooltip />
                                <Bar dataKey="value" fill={UI_DEFAULTS.CHART_COLORS[0]} radius={[4, 4, 0, 0]} />
                            </BarChart>
                        </ResponsiveContainer>
                    )}
                </div>
            )}
        </div>
    );
};
