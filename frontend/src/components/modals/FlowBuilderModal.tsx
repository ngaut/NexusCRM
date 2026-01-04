import React, { useState, useEffect, useMemo } from 'react';
import { createPortal } from 'react-dom';
import { X, Zap, AlertCircle, Save, Layers } from 'lucide-react';
import { Flow } from '../../infrastructure/api/flows';
import { dataAPI } from '../../infrastructure/api/data';
import { SObject } from '../../types';
import { SYSTEM_TABLE_NAMES } from '../../generated-schema';
import { useSchemas } from '../../core/hooks/useMetadata';
import { actionAPI, ActionMetadata } from '../../infrastructure/api/actions';
import { FLOW_STATUS, FlowStatus } from '../../core/constants/FlowConstants';
import { FlowStepEditor, FlowStep } from '../flows/FlowStepEditor';

// Sub-components
import { FlowGeneralInfo } from '../flows/builder/FlowGeneralInfo';
import { TriggerConfigPanel } from '../flows/builder/TriggerConfigPanel';
import { ActionConfigPanel } from '../flows/builder/ActionConfigPanel';

interface FlowBuilderModalProps {
    flow: Flow | null;
    objects: { api_name: string; label: string }[];
    onSave: (flowData: Partial<Flow>) => Promise<void>;
    onClose: () => void;
}

const FlowBuilderModal: React.FC<FlowBuilderModalProps> = ({
    flow,
    objects,
    onSave,
    onClose,
}) => {
    const [name, setName] = useState(flow?.name || '');
    const [triggerObject, setTriggerObject] = useState(flow?.trigger_object || '');
    const [triggerCondition, setTriggerCondition] = useState(flow?.trigger_condition || '');
    const [triggerType, setTriggerType] = useState(flow?.trigger_type || 'afterCreate');
    const [actionType, setActionType] = useState(flow?.action_type || 'createRecord');
    const [actionConfig, setActionConfig] = useState<Record<string, any>>(flow?.action_config || {});
    const [status, setStatus] = useState<FlowStatus>(flow?.status || FLOW_STATUS.DRAFT);
    const [flowType, setFlowType] = useState<'simple' | 'multistep'>(flow?.flow_type || 'simple');
    const [steps, setSteps] = useState<FlowStep[]>((flow?.steps as FlowStep[]) || []);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [availableActions, setAvailableActions] = useState<ActionMetadata[]>([]);
    const [loadingActions, setLoadingActions] = useState(false);
    const { schemas } = useSchemas();
    const [activeProcess, setActiveProcess] = useState<SObject | null>(null);

    // Get fields for the selected trigger object to power the Formula Editor
    const triggerFields = useMemo(() => {
        if (!triggerObject) return [];
        const schema = schemas.find(s => s.api_name === triggerObject);
        return schema?.fields || [];
    }, [schemas, triggerObject]);

    useEffect(() => {
        const fetchApprovalProcess = async () => {
            if (!triggerObject) {
                setActiveProcess(null);
                return;
            }
            try {
                const records = await dataAPI.query({
                    objectApiName: SYSTEM_TABLE_NAMES.SYSTEM_APPROVALPROCESS,
                    filterExpr: `object_api_name == '${triggerObject}' && is_active == true`,
                    limit: 1
                });
                setActiveProcess(records[0] || null);
            } catch (err) {
                console.warn("Failed to fetch active approval process", err);
                setActiveProcess(null);
            }
        };
        fetchApprovalProcess();
    }, [triggerObject]);

    useEffect(() => {
        const fetchActions = async () => {
            if (actionType === 'action' && triggerObject) {
                setLoadingActions(true);
                try {
                    const actions = await actionAPI.getActions(triggerObject);
                    setAvailableActions(actions);
                } catch {
                    setAvailableActions([]);
                } finally {
                    setLoadingActions(false);
                }
            } else {
                setAvailableActions([]);
            }
        };
        fetchActions();
    }, [actionType, triggerObject]);

    const handleSave = async () => {
        if (!name.trim()) {
            setError('Flow name is required');
            return;
        }
        if (!triggerObject) {
            setError('Trigger object is required');
            return;
        }
        if (flowType === 'simple' && !actionType) {
            setError('Action type is required');
            return;
        }

        // Sanitize field mappings
        const cleanActionConfig = { ...actionConfig };
        if (cleanActionConfig.field_mappings && typeof cleanActionConfig.field_mappings === 'object') {
            const cleanMappings: Record<string, unknown> = {};
            const mappings = cleanActionConfig.field_mappings as Record<string, unknown>;
            Object.entries(mappings).forEach(([k, v]) => {
                if (k && k.trim()) {
                    cleanMappings[k] = v;
                }
            });
            cleanActionConfig.field_mappings = cleanMappings;
        }

        try {
            setSaving(true);
            setError(null);
            const flowData: Partial<Flow> & { flow_type?: string; steps?: FlowStep[]; action_type?: string; action_config?: Record<string, unknown> } = {
                name: name.trim(),
                trigger_object: triggerObject,
                trigger_condition: triggerCondition.trim() || 'true',
                trigger_type: triggerType,
                status,
                flow_type: flowType,
            };
            // For multistep flows, include steps; for simple, include action
            if (flowType === 'multistep') {
                flowData.steps = steps;
            } else {
                flowData.action_type = actionType;
                flowData.action_config = cleanActionConfig;
            }
            await onSave(flowData);
        } catch (err) {
            const msg = err instanceof Error ? err.message : 'Failed to save flow';
            // Try to extract more specific API error if possible, but keep it type safe
            if (typeof err === 'object' && err !== null && 'response' in err) {
                const response = (err as { response: { data?: { error?: string; message?: string } } }).response;
                if (response?.data?.error) setError(response.data.error);
                else if (response?.data?.message) setError(response.data.message);
                else setError(msg);
            } else {
                setError(msg);
            }
            setSaving(false);
        }
    };

    const updateConfig = (key: string, value: unknown) => {
        setActionConfig((prev) => ({ ...prev, [key]: value }));
    };

    return createPortal(
        <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/50 backdrop-blur-sm">
            <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-hidden">
                {/* Header */}
                <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
                    <div className="flex items-center gap-3">
                        <div className="p-2 bg-gradient-to-br from-purple-500 to-indigo-600 rounded-xl">
                            <Zap className="w-5 h-5 text-white" />
                        </div>
                        <div>
                            <h2 className="text-xl font-bold text-gray-900 dark:text-white">
                                {flow ? 'Edit Flow' : 'New Flow'}
                            </h2>
                            <p className="text-sm text-gray-500 dark:text-gray-400">
                                Configure automation trigger and action
                            </p>
                        </div>
                    </div>
                    <button
                        onClick={onClose}
                        className="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 
              hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
                    >
                        <X className="w-5 h-5" />
                    </button>
                </div>

                {/* Body */}
                <div className="p-6 overflow-y-auto max-h-[calc(90vh-180px)] space-y-6">
                    {/* Error */}
                    {error && (
                        <div className="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 
              rounded-lg flex items-center gap-2 text-red-700 dark:text-red-400 text-sm">
                            <AlertCircle className="w-4 h-4" />
                            {error}
                        </div>
                    )}

                    {/* Basic Info */}
                    <FlowGeneralInfo
                        name={name}
                        setName={setName}
                        status={status}
                        setStatus={setStatus}
                    />

                    {/* Trigger Section */}
                    <TriggerConfigPanel
                        triggerObject={triggerObject}
                        setTriggerObject={setTriggerObject}
                        triggerType={triggerType}
                        setTriggerType={setTriggerType}
                        triggerCondition={triggerCondition}
                        setTriggerCondition={setTriggerCondition}
                        triggerFields={triggerFields}
                        objects={objects}
                    />

                    {/* Flow Type Toggle */}
                    <div className="flex items-center gap-4 p-4 bg-gradient-to-r from-purple-50 to-indigo-50 
                        dark:from-purple-900/20 dark:to-indigo-900/20 rounded-xl border border-purple-200 dark:border-purple-800">
                        <Layers className="w-5 h-5 text-purple-600 dark:text-purple-400" />
                        <span className="text-sm font-medium text-gray-700 dark:text-gray-300">Flow Type:</span>
                        <div className="flex gap-2">
                            <button
                                type="button"
                                onClick={() => setFlowType('simple')}
                                className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${flowType === 'simple'
                                    ? 'bg-purple-600 text-white shadow-md'
                                    : 'bg-white dark:bg-gray-700 text-gray-600 dark:text-gray-300 border border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-600'
                                    }`}
                            >
                                Simple (Single Action)
                            </button>
                            <button
                                type="button"
                                onClick={() => setFlowType('multistep')}
                                className={`px-4 py-2 rounded-lg text-sm font-medium transition-all ${flowType === 'multistep'
                                    ? 'bg-purple-600 text-white shadow-md'
                                    : 'bg-white dark:bg-gray-700 text-gray-600 dark:text-gray-300 border border-gray-300 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-600'
                                    }`}
                            >
                                Multi-Step
                            </button>
                        </div>
                    </div>

                    {/* Multi-Step Flow Editor */}
                    {flowType === 'multistep' && (
                        <div className="p-4 bg-gray-50 dark:bg-gray-700/50 rounded-xl">
                            <FlowStepEditor
                                steps={steps}
                                onStepsChange={setSteps}
                                triggerObject={triggerObject}
                                objects={objects}
                                activeProcessName={activeProcess?.name as string | undefined}
                            />
                        </div>
                    )}

                    {/* Action Section (only for simple flows) */}
                    {flowType === 'simple' && (
                        <ActionConfigPanel
                            actionType={actionType}
                            setActionType={(val) => {
                                setActionType(val);
                                setActionConfig({});
                            }}
                            actionConfig={actionConfig}
                            updateConfig={updateConfig}
                            objects={objects}
                            availableActions={availableActions}
                            loadingActions={loadingActions}
                            activeProcess={activeProcess}
                            triggerObject={triggerObject}
                            schemas={schemas}
                        />
                    )}
                </div>

                {/* Footer */}
                <div className="flex items-center justify-end gap-3 p-6 border-t border-gray-200 dark:border-gray-700">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 text-gray-700 dark:text-gray-300 hover:bg-gray-100 
              dark:hover:bg-gray-700 rounded-lg transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSave}
                        disabled={saving}
                        className="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-purple-500 to-indigo-600 
              text-white rounded-lg hover:from-purple-600 hover:to-indigo-700 transition-all 
              shadow-md disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        <Save className="w-4 h-4" />
                        {saving ? 'Saving...' : 'Save Flow'}
                    </button>
                </div>
            </div>
        </div>,
        document.body
    );
};

export default FlowBuilderModal;
