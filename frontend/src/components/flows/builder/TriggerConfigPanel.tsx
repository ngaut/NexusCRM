import React from 'react';
import { Settings } from 'lucide-react';
import { FormulaEditor } from '../../formula/FormulaEditor';
import { FieldMetadata } from '../../../types';

interface TriggerConfigPanelProps {
    triggerObject: string;
    setTriggerObject: (val: string) => void;
    triggerType: string;
    setTriggerType: (val: string) => void;
    triggerCondition: string;
    setTriggerCondition: (val: string) => void;
    triggerFields: FieldMetadata[];
    objects: { api_name: string; label: string }[];
}

export const TriggerConfigPanel: React.FC<TriggerConfigPanelProps> = ({
    triggerObject,
    setTriggerObject,
    triggerType,
    setTriggerType,
    triggerCondition,
    setTriggerCondition,
    triggerFields,
    objects
}) => {
    return (
        <div className="space-y-4 p-4 bg-gray-50 dark:bg-gray-700/50 rounded-xl">
            <h3 className="font-semibold text-gray-900 dark:text-white flex items-center gap-2">
                <Settings className="w-4 h-4" />
                Trigger Configuration
            </h3>

            <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Object *
                </label>
                <select
                    value={triggerObject}
                    onChange={(e) => setTriggerObject(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg 
                  bg-white dark:bg-gray-700 text-gray-900 dark:text-white 
                  focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                >
                    <option value="">Select object...</option>
                    {objects.map((obj) => (
                        <option key={obj.api_name} value={obj.api_name}>
                            {obj.label}
                        </option>
                    ))}
                </select>
            </div>

            <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Trigger Event *
                </label>
                <select
                    value={triggerType}
                    onChange={(e) => setTriggerType(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg 
                  bg-white dark:bg-gray-700 text-gray-900 dark:text-white 
                  focus:ring-2 focus:ring-purple-500 focus:border-transparent"
                >
                    <option value="beforeCreate">Before Create</option>
                    <option value="afterCreate">After Create</option>
                    <option value="beforeUpdate">Before Update</option>
                    <option value="afterUpdate">After Update</option>
                    <option value="beforeDelete">Before Delete</option>
                    <option value="afterDelete">After Delete</option>
                </select>
                <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    When should this flow trigger?
                </p>
            </div>

            <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                    Condition (Formula)
                </label>
                <div className="border border-gray-300 dark:border-gray-600 rounded-lg overflow-hidden">
                    {triggerObject ? (
                        <FormulaEditor
                            value={triggerCondition}
                            onChange={setTriggerCondition}
                            fields={triggerFields}
                        />
                    ) : (
                        <div className="p-4 text-sm text-gray-500 bg-gray-50 dark:bg-gray-800 text-center">
                            Please select an object first to edit criteria
                        </div>
                    )}
                </div>
                <p className="mt-1 text-xs text-gray-500 dark:text-gray-400">
                    Leave empty to run on every record
                </p>
            </div>
        </div>
    );
};
