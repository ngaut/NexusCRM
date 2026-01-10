import React from 'react';
import { Trash2, GripVertical, CheckCircle, XCircle, ArrowRight, AlertCircle } from 'lucide-react';
import { FlowStep } from '../../infrastructure/api/flows';

export interface StepCardProps {
    step: FlowStep;
    index: number;
    totalSteps: number;
    objects: { api_name: string; label: string }[];
    triggerObject: string;
    onUpdate: (updates: Partial<FlowStep>) => void;
    onRemove: () => void;
    onMoveUp: () => void;
    onMoveDown: () => void;
    allSteps: FlowStep[];
    activeProcessName?: string;
}

const STEP_TYPES = [
    { value: 'action', label: 'Action', icon: ArrowRight, color: 'bg-blue-500' },
    { value: 'approval', label: 'Approval', icon: CheckCircle, color: 'bg-amber-500' },
    { value: 'decision', label: 'Decision', icon: XCircle, color: 'bg-purple-500' },
];

const ACTION_TYPES = [
    { value: 'updateRecord', label: 'Update Record' },
    { value: 'createRecord', label: 'Create Record' },
    { value: 'sendEmail', label: 'Send Email' },
    { value: 'callWebhook', label: 'Call Webhook' },
];

export const FlowStepCard: React.FC<StepCardProps> = ({
    step,
    index,
    objects,
    onUpdate,
    onRemove,
    onMoveUp,
    allSteps,
    activeProcessName,
}) => {
    const stepTypeInfo = STEP_TYPES.find(t => t.value === step.step_type) || STEP_TYPES[0];
    const StepIcon = stepTypeInfo.icon;

    return (
        <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 
            rounded-xl shadow-sm overflow-hidden">
            {/* Step Header */}
            <div className={`flex items-center gap-3 px-4 py-3 ${stepTypeInfo.color} bg-opacity-10 
                dark:bg-opacity-20 border-b border-gray-200 dark:border-gray-700`}>
                <div className="flex items-center gap-2">
                    <button
                        type="button"
                        onClick={onMoveUp}
                        disabled={index === 0}
                        className="p-1 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 
                            disabled:opacity-30 disabled:cursor-not-allowed"
                        title="Move up"
                    >
                        <GripVertical className="w-4 h-4" />
                    </button>
                    <span className="w-6 h-6 flex items-center justify-center bg-gray-200 dark:bg-gray-700 
                        rounded-full text-xs font-bold text-gray-600 dark:text-gray-300">
                        {index + 1}
                    </span>
                </div>

                <div className={`p-1.5 rounded-lg ${stepTypeInfo.color}`}>
                    <StepIcon className="w-4 h-4 text-white" />
                </div>

                <input
                    type="text"
                    value={step.step_name}
                    onChange={(e) => onUpdate({ step_name: e.target.value })}
                    className="flex-1 bg-transparent border-none text-sm font-medium text-gray-900 
                        dark:text-white focus:outline-none focus:ring-0"
                    placeholder="Step name..."
                />

                <button
                    type="button"
                    onClick={onRemove}
                    className="p-1.5 text-red-400 hover:text-red-600 hover:bg-red-50 
                        dark:hover:bg-red-900/20 rounded-lg transition-colors"
                    title="Remove step"
                >
                    <Trash2 className="w-4 h-4" />
                </button>
            </div>

            {/* Step Body */}
            <div className="p-4 space-y-4">
                {/* Step Type Selection */}
                <div className="flex gap-2">
                    {STEP_TYPES.map((type) => {
                        const Icon = type.icon;
                        const isSelected = step.step_type === type.value;
                        return (
                            <button
                                key={type.value}
                                type="button"
                                onClick={() => onUpdate({ step_type: type.value as FlowStep['step_type'] })}
                                className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm 
                                    transition-all ${isSelected
                                        ? `${type.color} text-white shadow-md`
                                        : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'
                                    }`}
                            >
                                <Icon className="w-3.5 h-3.5" />
                                {type.label}
                            </button>
                        );
                    })}
                </div>

                {/* Action Step Config */}
                {step.step_type === 'action' && (
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
                                Action Type
                            </label>
                            <select
                                value={step.action_type || 'updateRecord'}
                                onChange={(e) => onUpdate({ action_type: e.target.value })}
                                className="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 
                                    rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                            >
                                {ACTION_TYPES.map(t => (
                                    <option key={t.value} value={t.value}>{t.label}</option>
                                ))}
                            </select>
                        </div>
                        {step.action_type === 'createRecord' && (
                            <div>
                                <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
                                    Target Object
                                </label>
                                <select
                                    value={(step.action_config?.target_object as string) || ''}
                                    onChange={(e) => onUpdate({
                                        action_config: { ...step.action_config, target_object: e.target.value }
                                    })}
                                    className="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 
                                        rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                                >
                                    <option value="">Select object...</option>
                                    {objects.map(obj => (
                                        <option key={obj.api_name} value={obj.api_name}>{obj.label}</option>
                                    ))}
                                </select>
                            </div>
                        )}
                    </div>
                )}

                {/* Approval Step Config */}
                {step.step_type === 'approval' && (
                    <div>
                        <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
                            Submission Comments (Optional)
                        </label>
                        <textarea
                            value={(step.action_config?.comments as string) || ''}
                            onChange={(e) => onUpdate({
                                action_config: { ...step.action_config, comments: e.target.value }
                            })}
                            placeholder="Comments to include with approval request..."
                            rows={2}
                            className="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 
                                rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                        />
                        <div className={`mt-2 p-2 rounded-lg border flex items-start gap-2 text-xs
                            ${activeProcessName
                                ? 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800 text-blue-800 dark:text-blue-300'
                                : 'bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800 text-yellow-800 dark:text-yellow-300'
                            }`}>
                            <AlertCircle className="w-3 h-3 mt-0.5 shrink-0" />
                            <div>
                                {activeProcessName ? (
                                    <p>
                                        Will submit to active process: <strong>{activeProcessName}</strong>
                                    </p>
                                ) : (
                                    <p>
                                        <strong>Warning:</strong> No active approval process found.
                                    </p>
                                )}
                            </div>
                        </div>
                    </div>
                )}

                {/* Routing (for approval and decision steps) */}
                {(step.step_type === 'approval' || step.step_type === 'decision') && allSteps.length > 1 && (
                    <div className="grid grid-cols-2 gap-4 pt-2 border-t border-gray-100 dark:border-gray-700">
                        <div>
                            <label className="block text-xs font-medium text-green-600 dark:text-green-400 mb-1">
                                On Success → Go to
                            </label>
                            <select
                                value={step.on_success_step || ''}
                                onChange={(e) => onUpdate({ on_success_step: e.target.value || undefined })}
                                className="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 
                                    rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                            >
                                <option value="">Next step (default)</option>
                                {allSteps.filter(s => s.id !== step.id).map(s => (
                                    <option key={s.id} value={s.id}>
                                        Step {s.step_order}: {s.step_name}
                                    </option>
                                ))}
                                <option value="__END__">End Flow</option>
                            </select>
                        </div>
                        <div>
                            <label className="block text-xs font-medium text-red-600 dark:text-red-400 mb-1">
                                On Failure → Go to
                            </label>
                            <select
                                value={step.on_failure_step || ''}
                                onChange={(e) => onUpdate({ on_failure_step: e.target.value || undefined })}
                                className="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 
                                    rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                            >
                                <option value="">End Flow (default)</option>
                                {allSteps.filter(s => s.id !== step.id).map(s => (
                                    <option key={s.id} value={s.id}>
                                        Step {s.step_order}: {s.step_name}
                                    </option>
                                ))}
                            </select>
                        </div>
                    </div>
                )}

                {/* Entry Condition */}
                <div>
                    <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
                        Entry Condition (optional)
                    </label>
                    <input
                        type="text"
                        value={step.entry_condition || ''}
                        onChange={(e) => onUpdate({ entry_condition: e.target.value })}
                        placeholder="e.g., amount > 10000 (skip step if false)"
                        className="w-full px-3 py-2 text-sm font-mono border border-gray-300 dark:border-gray-600 
                            rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    />
                </div>
            </div>
        </div>
    );
};
