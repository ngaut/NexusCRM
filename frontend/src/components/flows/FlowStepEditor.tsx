import React from 'react';
import { Plus } from 'lucide-react';
import { FlowStepCard } from './FlowStepCard';
import { FlowStep } from '../../infrastructure/api/flows';
import { COMMON_FIELDS } from '../../core/constants';


interface FlowStepEditorProps {
    steps: FlowStep[];
    onStepsChange: (steps: FlowStep[]) => void;
    triggerObject: string;
    objects: { api_name: string; label: string }[];
    activeProcessName?: string;
}

export const FlowStepEditor: React.FC<FlowStepEditorProps> = ({
    steps,
    onStepsChange,
    triggerObject,
    objects,
    activeProcessName,
}) => {
    const generateId = () => `step-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;

    const addStep = () => {
        const newStep: FlowStep = {
            [COMMON_FIELDS.ID]: generateId(),
            step_order: steps.length + 1,
            step_name: `Step ${steps.length + 1}`,
            step_type: 'action',
            action_type: 'updateRecord',
            action_config: {},
        };
        onStepsChange([...steps, newStep]);
    };

    const removeStep = (stepId: string) => {
        const filtered = steps.filter(s => s[COMMON_FIELDS.ID] !== stepId);
        // Re-order remaining steps
        const reordered = filtered.map((s, idx) => ({ ...s, step_order: idx + 1 }));
        onStepsChange(reordered);
    };

    const updateStep = (stepId: string, updates: Partial<FlowStep>) => {
        onStepsChange(steps.map(s => s[COMMON_FIELDS.ID] === stepId ? { ...s, ...updates } : s));
    };

    const moveStep = (fromIndex: number, toIndex: number) => {
        if (toIndex < 0 || toIndex >= steps.length) return;
        const newSteps = [...steps];
        const [moved] = newSteps.splice(fromIndex, 1);
        newSteps.splice(toIndex, 0, moved);
        // Re-order
        const reordered = newSteps.map((s, idx) => ({ ...s, step_order: idx + 1 }));
        onStepsChange(reordered);
    };

    return (
        <div className="space-y-4">
            <div className="flex items-center justify-between">
                <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300">
                    Flow Steps ({steps.length})
                </h4>
                <button
                    type="button"
                    onClick={addStep}
                    className="flex items-center gap-1 px-3 py-1.5 text-sm bg-purple-100 dark:bg-purple-900/30 
                        text-purple-700 dark:text-purple-300 rounded-lg hover:bg-purple-200 
                        dark:hover:bg-purple-900/50 transition-colors"
                >
                    <Plus className="w-4 h-4" />
                    Add Step
                </button>
            </div>

            {steps.length === 0 ? (
                <div className="text-center py-8 text-gray-500 dark:text-gray-400 border-2 border-dashed 
                    border-gray-200 dark:border-gray-700 rounded-xl">
                    <p className="text-sm">No steps defined yet</p>
                    <p className="text-xs mt-1">Click "Add Step" to create your first step</p>
                </div>
            ) : (
                <div className="space-y-3">
                    {steps.map((step, index) => (
                        <FlowStepCard
                            key={step[COMMON_FIELDS.ID]}
                            step={step}
                            index={index}
                            totalSteps={steps.length}
                            objects={objects}
                            triggerObject={triggerObject}
                            onUpdate={(updates) => updateStep(step[COMMON_FIELDS.ID], updates)}
                            onRemove={() => removeStep(step[COMMON_FIELDS.ID])}
                            onMoveUp={() => moveStep(index, index - 1)}
                            onMoveDown={() => moveStep(index, index + 1)}
                            allSteps={steps}
                            activeProcessName={activeProcessName}
                        />
                    ))}
                </div>
            )}
        </div>
    );
};

export default FlowStepEditor;
