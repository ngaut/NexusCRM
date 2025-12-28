import React from 'react';
import { ShieldCheck, Edit, Trash2 } from 'lucide-react';
import { ApprovalProcess, ObjectMetadata } from '../../types';

interface ApprovalProcessListProps {
    processes: ApprovalProcess[];
    schemas: ObjectMetadata[];
    loading: boolean;
    onEdit: (process: ApprovalProcess) => void;
    onDelete: (process: ApprovalProcess) => void;
    onCreate: () => void;
}

export const ApprovalProcessList: React.FC<ApprovalProcessListProps> = ({
    processes,
    schemas,
    loading,
    onEdit,
    onDelete,
    onCreate
}) => {
    const getObjectLabel = (apiName: string) => {
        const schema = schemas.find(s => s.api_name === apiName);
        return schema?.label || apiName;
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center py-12">
                <div className="w-8 h-8 border-2 border-green-500 border-t-transparent rounded-full animate-spin" />
            </div>
        );
    }

    if (processes.length === 0) {
        return (
            <div className="text-center py-12 bg-gray-50 dark:bg-gray-800/50 rounded-xl border border-dashed border-gray-200 dark:border-gray-700">
                <ShieldCheck className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-2">No Approval Processes</h3>
                <p className="text-gray-500 dark:text-gray-400 mb-4">
                    Create your first approval process to enable the "Submit for Approval" button
                </p>
                <button
                    onClick={onCreate}
                    className="px-4 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 transition-colors"
                >
                    Create Process
                </button>
            </div>
        );
    }

    return (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {processes.map((process) => (
                <div
                    key={process.id}
                    className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-xl rounded-2xl border border-white/20 dark:border-gray-700 shadow-xl hover:shadow-2xl transition-all p-5 flex flex-col justify-between group"
                >
                    <div>
                        <div className="flex items-center justify-between mb-3">
                            <div className="flex items-center gap-3">
                                <div className={`p-2 rounded-lg ${process.is_active
                                    ? 'bg-green-100 dark:bg-green-900/30 text-green-600 dark:text-green-400'
                                    : 'bg-gray-100 dark:bg-gray-700 text-gray-400'
                                    }`}>
                                    <ShieldCheck className="w-5 h-5" />
                                </div>
                                <span className={`px-2 py-0.5 text-xs font-medium rounded-full ${process.is_active
                                    ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                                    : 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'
                                    }`}>
                                    {process.is_active ? 'Active' : 'Inactive'}
                                </span>
                            </div>
                            <div className="flex gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                <button
                                    onClick={() => onEdit(process)}
                                    className="p-1.5 text-blue-600 hover:bg-blue-50 dark:text-blue-400 dark:hover:bg-blue-900/20 rounded-lg transition-colors"
                                >
                                    <Edit className="w-4 h-4" />
                                </button>
                                <button
                                    onClick={() => onDelete(process)}
                                    className="p-1.5 text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20 rounded-lg transition-colors"
                                >
                                    <Trash2 className="w-4 h-4" />
                                </button>
                            </div>
                        </div>

                        <h3 className="text-lg font-bold text-gray-900 dark:text-white mb-1">
                            {process.name}
                        </h3>

                        <p className="text-sm text-gray-500 dark:text-gray-400 mb-4 line-clamp-2 h-10">
                            {process.description || "No description provided."}
                        </p>

                        <div className="space-y-2">
                            <div className="flex items-center justify-between text-sm p-2 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
                                <span className="text-gray-500 dark:text-gray-400">Target Object</span>
                                <span className="font-medium text-gray-900 dark:text-white bg-white dark:bg-gray-600 px-2 py-0.5 rounded shadow-sm border border-gray-100 dark:border-gray-500">
                                    {getObjectLabel(process.object_api_name)}
                                </span>
                            </div>
                            <div className="flex items-center justify-between text-sm p-2 bg-gray-50 dark:bg-gray-700/50 rounded-lg">
                                <span className="text-gray-500 dark:text-gray-400">Approver</span>
                                <div className="text-right">
                                    <span className="font-medium text-gray-900 dark:text-white block">
                                        {process.approver_type}
                                    </span>
                                    {process.approver_id && (
                                        <span className="text-xs text-gray-400 block font-mono">
                                            {process.approver_id.substring(0, 8)}...
                                        </span>
                                    )}
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            ))}
        </div>
    );
};
