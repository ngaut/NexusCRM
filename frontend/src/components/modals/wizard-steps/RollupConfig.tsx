import React from 'react';
import { ObjectMetadata } from '../../../types';

interface RollupConfigProps {
    rollupSummaryObject: string;
    setRollupSummaryObject: (value: string) => void;
    rollupRelationshipField: string;
    setRollupRelationshipField: (value: string) => void;
    rollupCalcType: "COUNT" | "SUM" | "MIN" | "MAX" | "AVG";
    setRollupCalcType: (value: "COUNT" | "SUM" | "MIN" | "MAX" | "AVG") => void;
    rollupSummaryField: string;
    setRollupSummaryField: (value: string) => void;
    schemas: ObjectMetadata[];
    objectApiName: string;
}

export const RollupConfig: React.FC<RollupConfigProps> = ({
    rollupSummaryObject,
    setRollupSummaryObject,
    rollupRelationshipField,
    setRollupRelationshipField,
    rollupCalcType,
    setRollupCalcType,
    rollupSummaryField,
    setRollupSummaryField,
    schemas,
    objectApiName
}) => {
    return (
        <div className="p-4 bg-indigo-50 rounded-lg border border-indigo-100 space-y-4">
            <div>
                <label className="block text-sm font-medium text-indigo-800 mb-1.5">
                    Summarized Object <span className="text-red-500">*</span>
                </label>
                <select
                    value={rollupSummaryObject}
                    onChange={(e) => {
                        setRollupSummaryObject(e.target.value);
                        setRollupRelationshipField(''); // Reset relationship field when object changes
                    }}
                    className="w-full px-3 py-2 border border-indigo-200 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-sm bg-white"
                    required
                >
                    <option value="">Select an object...</option>
                    {schemas.map(schema => (
                        <option key={schema.api_name} value={schema.api_name}>
                            {schema.label} ({schema.api_name})
                        </option>
                    ))}
                </select>
                <p className="mt-1 text-xs text-indigo-500">Object containing the records to aggregate</p>
            </div>

            <div>
                <label className="block text-sm font-medium text-indigo-800 mb-1.5">
                    Relationship Field <span className="text-red-500">*</span>
                </label>
                <select
                    value={rollupRelationshipField}
                    onChange={(e) => setRollupRelationshipField(e.target.value)}
                    className="w-full px-3 py-2 border border-indigo-200 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-sm bg-white"
                    required
                    disabled={!rollupSummaryObject}
                >
                    <option value="">Select a lookup field...</option>
                    {rollupSummaryObject && schemas.find(s => s.api_name === rollupSummaryObject)?.fields
                        .filter(f => f.type === 'Lookup' && f.reference_to?.includes(objectApiName))
                        .map(field => (
                            <option key={field.api_name} value={field.api_name}>
                                {field.label} ({field.api_name})
                            </option>
                        ))
                    }
                </select>
                <p className="mt-1 text-xs text-indigo-500">The lookup field on the child object pointing to this parent</p>
            </div>

            <div className="grid grid-cols-2 gap-4">
                <div>
                    <label className="block text-sm font-medium text-indigo-800 mb-1.5">
                        Roll-up Type <span className="text-red-500">*</span>
                    </label>
                    <select
                        value={rollupCalcType}
                        onChange={(e) => setRollupCalcType(e.target.value as "COUNT" | "SUM" | "MIN" | "MAX" | "AVG")}
                        className="w-full px-3 py-2 border border-indigo-200 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-sm"
                    >
                        <option value="COUNT">COUNT</option>
                        <option value="SUM">SUM</option>
                        <option value="MIN">MIN</option>
                        <option value="MAX">MAX</option>
                    </select>
                </div>

                {rollupCalcType !== 'COUNT' && (
                    <div>
                        <label className="block text-sm font-medium text-indigo-800 mb-1.5">
                            Field to Aggregate <span className="text-red-500">*</span>
                        </label>
                        <input
                            type="text"
                            value={rollupSummaryField}
                            onChange={(e) => setRollupSummaryField(e.target.value)}
                            placeholder="e.g., amount"
                            className="w-full px-3 py-2 border border-indigo-200 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent text-sm"
                            required
                        />
                    </div>
                )}
            </div>
        </div>
    );
};
