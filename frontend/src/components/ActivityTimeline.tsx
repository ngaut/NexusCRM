import React from 'react';
import { useQuery } from '../core/hooks/useData';
import { CheckCircle2, Clock, Mail, Phone, Calendar } from 'lucide-react';
import { format } from 'date-fns';

interface ActivityTimelineProps {
    parentRecordId: string;
    parentObjectApiName: string;
}

export const ActivityTimeline: React.FC<ActivityTimelineProps> = ({
    parentRecordId,
    parentObjectApiName
}) => {
    // Fetch tasks related to this record
    // We filter by RelatedToId matching the parent record ID
    const { records: tasks, loading, error } = useQuery({
        objectApiName: 'task',
        filterExpr: `related_to_id == '${parentRecordId}'`,
        sortField: 'created_date', // Ideally activity_date, but created_date works for now
        sortDirection: 'DESC'
    });

    if (loading) return <div className="p-4 text-center text-slate-500 text-sm">Loading activities...</div>;
    // Gracefully handle when Task object doesn't exist in the system
    if (error || !tasks) {
        return (
            <div className="p-8 text-center border border-slate-200 rounded-lg bg-slate-50">
                <div className="inline-flex items-center justify-center w-10 h-10 rounded-full bg-slate-100 mb-3">
                    <Clock className="text-slate-400" size={20} />
                </div>
                <h3 className="text-sm font-medium text-slate-900">No activities yet</h3>
                <p className="text-xs text-slate-500 mt-1">
                    Activity tracking is not configured.
                </p>
            </div>
        );
    }

    if (!tasks || tasks.length === 0) {
        return (
            <div className="p-8 text-center border border-slate-200 rounded-lg bg-slate-50">
                <div className="inline-flex items-center justify-center w-10 h-10 rounded-full bg-slate-100 mb-3">
                    <Clock className="text-slate-400" size={20} />
                </div>
                <h3 className="text-sm font-medium text-slate-900">No activities yet</h3>
                <p className="text-xs text-slate-500 mt-1">
                    Log a call, create a task, or send an email to see it here.
                </p>
            </div>
        );
    }

    const getIcon = (type: string) => {
        switch (type?.toLowerCase()) {
            case 'email': return <Mail size={14} />;
            case 'call': return <Phone size={14} />;
            case 'meeting': return <Calendar size={14} />;
            default: return <CheckCircle2 size={14} />;
        }
    };

    const getIconColor = (type: string) => {
        switch (type?.toLowerCase()) {
            case 'email': return 'bg-blue-100 text-blue-600';
            case 'call': return 'bg-green-100 text-green-600';
            case 'meeting': return 'bg-purple-100 text-purple-600';
            default: return 'bg-slate-100 text-slate-600';
        }
    };

    return (
        <div className="space-y-6">
            <h3 className="text-sm font-bold text-slate-800 uppercase tracking-wide">Activity Timeline</h3>
            <div className="relative border-l-2 border-slate-200 ml-3 space-y-8 pb-4">
                {tasks.map((task) => (
                    <div key={task.id} className="relative pl-8">
                        {/* Icon */}
                        <div className={`absolute -left-[9px] top-0 w-5 h-5 rounded-full border-2 border-white flex items-center justify-center ${getIconColor(String(task.type))}`}>
                            {getIcon(String(task.type))}
                        </div>

                        {/* Content */}
                        <div className="bg-white rounded-lg border border-slate-200 p-4 shadow-sm hover:shadow-md transition-shadow">
                            <div className="flex justify-between items-start mb-1">
                                <h4 className="text-sm font-semibold text-slate-900">
                                    {String(task.subject || 'No Subject')}
                                </h4>
                                <span className="text-xs text-slate-500 whitespace-nowrap">
                                    {task.created_date ? format(new Date(String(task.created_date)), 'MMM d, h:mm a') : ''}
                                </span>
                            </div>

                            <div className="text-xs text-slate-500 mb-2 flex gap-2">
                                <span className="px-1.5 py-0.5 bg-slate-100 rounded text-slate-600 font-medium">
                                    {String(task.status || 'Open')}
                                </span>
                                {task.priority && (
                                    <span className={`px-1.5 py-0.5 rounded font-medium ${task.priority === 'High' ? 'bg-red-50 text-red-600' : 'bg-slate-50 text-slate-600'
                                        }`}>
                                        {String(task.priority)}
                                    </span>
                                )}
                            </div>

                            {task.description && (
                                <p className="text-sm text-slate-600 line-clamp-3">
                                    {String(task.description)}
                                </p>
                            )}
                        </div>
                    </div>
                ))}
            </div>
        </div>
    );
};
