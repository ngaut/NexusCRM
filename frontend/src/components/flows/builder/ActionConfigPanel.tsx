import React from 'react';
import { Zap, AlertCircle } from 'lucide-react';
import { ActionMetadata } from '../../../infrastructure/api/actions';
import { SObject, ObjectMetadata } from '../../../types';
import { FieldMappingBuilder } from '../../flows/FieldMappingBuilder';

interface ActionConfigPanelProps {
    actionType: string;
    setActionType: (val: string) => void;
    actionConfig: Record<string, any>;
    updateConfig: (key: string, value: unknown) => void;
    objects: { api_name: string; label: string }[];
    availableActions: ActionMetadata[];
    loadingActions: boolean;
    activeProcess: SObject | null;
    triggerObject: string;
    schemas: ObjectMetadata[];
}

const ACTION_TYPES = [
    { value: 'createRecord', label: 'Create Record', description: 'Create a new record in another object' },
    { value: 'updateRecord', label: 'Update Record', description: 'Update the triggering record' },
    { value: 'action', label: 'Execute Action', description: 'Run a predefined action by ID' },
    { value: 'sendEmail', label: 'Send Email', description: 'Send an email notification' },
    { value: 'callWebhook', label: 'Call Webhook', description: 'Make an HTTP request to external service' },
    { value: 'submitForApproval', label: 'Submit for Approval', description: 'Submit record for approval workflow' },
];

export const ActionConfigPanel: React.FC<ActionConfigPanelProps> = ({
    actionType,
    setActionType,
    actionConfig,
    updateConfig,
    objects,
    availableActions,
    loadingActions,
    activeProcess,
    triggerObject,
    schemas
}) => {
    return (
        <div className="space-y-4 p-4 bg-gray-50 dark:bg-gray-700/50 rounded-xl">
            <h3 className="font-semibold text-gray-900 dark:text-white flex items-center gap-2">
                <Zap className="w-4 h-4" />
                Action Configuration
            </h3>

            <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Action Type *
                </label>
                <select
                    value={actionType}
                    onChange={(e) => setActionType(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg 
                  bg-white dark:bg-gray-700 text-gray-900 dark:text-white 
                  focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                >
                    {ACTION_TYPES.map((type) => (
                        <option key={type.value} value={type.value}>
                            {type.label}
                        </option>
                    ))}
                </select>
                <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    {ACTION_TYPES.find((t) => t.value === actionType)?.description}
                </p>
            </div>

            {/* Dynamic Action Config Fields */}
            {actionType === 'createRecord' && (
                <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Target Object
                    </label>
                    <select
                        value={actionConfig.target_object || ''}
                        onChange={(e) => updateConfig('target_object', e.target.value)}
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg 
                    bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    >
                        <option value="">Select object...</option>
                        {objects.map((obj) => (
                            <option key={obj.api_name} value={obj.api_name}>
                                {obj.label}
                            </option>
                        ))}
                    </select>
                </div>
            )}

            {actionType === 'action' && (
                <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Action ID
                    </label>
                    {loadingActions ? (
                        <div className="text-sm text-gray-500 animate-pulse py-2">Loading actions...</div>
                    ) : (
                        <select
                            value={actionConfig.action_id || ''}
                            onChange={(e) => updateConfig('action_id', e.target.value)}
                            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg 
                                            bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                        >
                            <option value="">Select an action...</option>
                            {availableActions.map(action => (
                                <option key={action.id} value={action.id}>
                                    {action.label} ({action.name})
                                </option>
                            ))}
                        </select>
                    )}
                    <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                        Select the configured action to execute
                    </p>
                </div>
            )}

            {actionType === 'callWebhook' && (
                <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Webhook URL
                    </label>
                    <input
                        type="url"
                        value={actionConfig.url || ''}
                        onChange={(e) => updateConfig('url', e.target.value)}
                        placeholder="https://..."
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg 
                    bg-white dark:bg-gray-700 text-gray-900 dark:text-white font-mono text-sm"
                    />
                </div>
            )}

            {actionType === 'submitForApproval' && (
                <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Submission Comments (Optional)
                    </label>
                    <textarea
                        value={actionConfig.comments || ''}
                        onChange={(e) => updateConfig('comments', e.target.value)}
                        placeholder="Comments to include with approval request..."
                        rows={2}
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg 
                                        bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                    />
                    <div className={`mt-3 p-3 rounded-lg border flex items-start gap-2 text-sm
                                        ${activeProcess
                            ? 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800 text-blue-800 dark:text-blue-300'
                            : 'bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800 text-yellow-800 dark:text-yellow-300'
                        }`}>
                        <AlertCircle className="w-4 h-4 mt-0.5 shrink-0" />
                        <div>
                            {activeProcess ? (
                                <p>
                                    Will submit to active process: <strong>{activeProcess.name as string}</strong>
                                </p>
                            ) : (
                                <p>
                                    <strong>Warning:</strong> No active approval process found for {triggerObject || 'this object'}.
                                    This flow will fail until one is created and activated.
                                </p>
                            )}
                        </div>
                    </div>
                </div>
            )}

            {(actionType === 'createRecord' || actionType === 'updateRecord') && (
                <FieldMappingBuilder
                    targetObjectApiName={actionConfig.target_object || (actionType === 'updateRecord' ? triggerObject : '')}
                    schemas={schemas}
                    mapping={actionConfig.field_mappings || {}}
                    onChange={(newMapping) => updateConfig('field_mappings', newMapping)}
                />
            )}
        </div>
    );
};
