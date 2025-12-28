import React from 'react';

export interface AuditLog {
    id: string;
    field_name: string;
    old_value: string;
    new_value: string;
    changed_by: string;
    changed_at: string;
}

interface AuditLogListProps {
    history: AuditLog[];
}

export const AuditLogList: React.FC<AuditLogListProps> = ({ history }) => {
    if (history.length === 0) {
        return <div className="text-center text-slate-400 text-sm py-4">No history available</div>;
    }

    return (
        <div className="space-y-4">
            {history.map(h => (
                <div key={h.id} className="flex gap-3 text-sm">
                    <div className="min-w-[40px] flex flex-col items-center">
                        <div className="w-2 h-2 rounded-full bg-slate-300 mt-2" />
                        <div className="w-0.5 bg-slate-100 h-full mt-1" />
                    </div>
                    <div className="pb-4">
                        <p className="text-slate-900">
                            <span className="font-medium">{h.changed_by}</span> changed
                            <span className="font-medium mx-1">{h.field_name}</span>
                        </p>
                        <div className="flex items-center gap-2 mt-1 text-slate-500 text-xs">
                            <span className="line-through">{h.old_value || '(empty)'}</span>
                            <span>→</span>
                            <span className="text-slate-700 bg-green-50 px-1 rounded">{h.new_value}</span>
                        </div>
                        <div className="text-xs text-slate-400 mt-1">{h.changed_at}</div>
                    </div>
                </div>
            ))}
        </div>
    );
};
