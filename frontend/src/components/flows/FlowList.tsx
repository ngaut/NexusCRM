import React from 'react';
import { Play, Pause, PlayCircle, Edit, Trash2, CheckCircle, Zap, Clock } from 'lucide-react';
import type { Flow } from '../../infrastructure/api/flows';
import { FLOW_STATUS } from '../../core/constants/FlowConstants';

interface FlowListProps {
    flows: Flow[];
    loading: boolean;
    onToggleStatus: (flow: Flow) => void;
    onEdit: (flow: Flow) => void;
    onDelete: (flow: Flow) => void;
    onExecute: (flow: Flow) => void;
    onCreate: () => void;
}

export const FlowList: React.FC<FlowListProps> = ({
    flows,
    loading,
    onToggleStatus,
    onEdit,
    onDelete,
    onExecute,
    onCreate
}) => {
    const getActionTypeLabel = (type: string) => {
        const labels: Record<string, string> = {
            'action': 'Execute Action',
            'createRecord': 'Create Record',
            'updateRecord': 'Update Record',
            'sendEmail': 'Send Email',
            'callWebhook': 'Call Webhook',
        };
        return labels[type] || type;
    };

    const getTriggerTypeLabel = (type: string) => {
        const labels: Record<string, string> = {
            'beforeCreate': 'Before Create',
            'afterCreate': 'After Create',
            'beforeUpdate': 'Before Update',
            'afterUpdate': 'After Update',
            'beforeDelete': 'Before Delete',
            'afterDelete': 'After Delete',
        };
        return labels[type] || type;
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center py-12">
                <div className="w-8 h-8 border-2 border-purple-500 border-t-transparent rounded-full animate-spin" />
            </div>
        );
    }

    if (flows.length === 0) {
        return (
            <div className="text-center py-12 bg-gray-50 dark:bg-gray-800/50 rounded-xl border border-dashed border-gray-200 dark:border-gray-700">
                <Zap className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">No Flows Yet</h3>
                <p className="text-gray-500 dark:text-gray-400 mb-4">
                    Create your first flow to automate record actions
                </p>
                <button
                    onClick={onCreate}
                    className="px-4 py-2 bg-purple-500 text-white rounded-lg hover:bg-purple-600 transition-colors"
                >
                    Create Flow
                </button>
            </div>
        );
    }

    return (
        <div className="space-y-4">
            {flows.map((flow) => (
                <div
                    key={flow.id}
                    className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-xl rounded-2xl border border-white/20 dark:border-gray-700 shadow-xl hover:shadow-2xl transition-all p-4"
                >
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-4">
                            <button
                                onClick={() => onToggleStatus(flow)}
                                className={`p-2 rounded-full transition-colors ${flow.status === FLOW_STATUS.ACTIVE
                                    ? 'bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400'
                                    : 'bg-gray-100 dark:bg-gray-700 text-gray-400'
                                    }`}
                                title={flow.status === FLOW_STATUS.ACTIVE ? 'Deactivate' : 'Activate'}
                            >
                                {flow.status === FLOW_STATUS.ACTIVE ? <Play className="w-5 h-5" /> : <Pause className="w-5 h-5" />}
                            </button>
                            <div>
                                <div className="flex items-center gap-2">
                                    <h3 className="font-semibold text-gray-900 dark:text-white">{flow.name}</h3>
                                    <span className={`px-2 py-0.5 text-xs font-medium rounded-full ${flow.status === FLOW_STATUS.ACTIVE
                                        ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                                        : 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'
                                        }`}>
                                        {flow.status}
                                    </span>
                                </div>
                                <div className="flex items-center gap-4 mt-1 text-sm text-gray-500 dark:text-gray-400">
                                    <span className="flex items-center gap-1">
                                        <CheckCircle className="w-4 h-4" />
                                        {flow.trigger_object} • {getTriggerTypeLabel(flow.trigger_type)}
                                    </span>
                                    <span>•</span>
                                    <span className="flex items-center gap-1">
                                        <Zap className="w-4 h-4" />
                                        {getActionTypeLabel(flow.action_type)}
                                    </span>
                                    {flow.last_modified && (
                                        <>
                                            <span>•</span>
                                            <span className="flex items-center gap-1">
                                                <Clock className="w-4 h-4" />
                                                {new Date(flow.last_modified).toLocaleDateString()}
                                            </span>
                                        </>
                                    )}
                                </div>
                            </div>
                        </div>
                        <div className="flex items-center gap-2">
                            {flow.status === FLOW_STATUS.ACTIVE && (
                                <button
                                    onClick={() => onExecute(flow)}
                                    className="p-2 text-gray-400 hover:text-green-500 hover:bg-green-50 dark:hover:bg-green-900/20 rounded-lg transition-colors"
                                    title="Run Flow"
                                >
                                    <PlayCircle className="w-4 h-4" />
                                </button>
                            )}
                            <button
                                onClick={() => onEdit(flow)}
                                className="p-2 text-gray-400 hover:text-blue-500 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-lg transition-colors"
                            >
                                <Edit className="w-4 h-4" />
                            </button>
                            <button
                                onClick={() => onDelete(flow)}
                                className="p-2 text-gray-400 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
                            >
                                <Trash2 className="w-4 h-4" />
                            </button>
                        </div>
                    </div>
                </div>
            ))}
        </div>
    );
};
