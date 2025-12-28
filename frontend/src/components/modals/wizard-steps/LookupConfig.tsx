import React from 'react';
import { ObjectMetadata } from '../../../types';

interface LookupConfigProps {
    lookupTarget: string;
    setLookupTarget: (value: string) => void;
    isMasterDetail: boolean;
    setIsMasterDetail: (value: boolean) => void;
    setRequired: (value: boolean) => void;
    schemas: ObjectMetadata[];
}

export const LookupConfig: React.FC<LookupConfigProps> = ({
    lookupTarget,
    setLookupTarget,
    isMasterDetail,
    setIsMasterDetail,
    setRequired,
    schemas
}) => {
    return (
        <div className="p-4 bg-purple-50 rounded-lg border border-purple-100 space-y-4">
            <div>
                <label className="block text-sm font-medium text-purple-800 mb-1.5">
                    Related Object <span className="text-red-500">*</span>
                </label>
                <select
                    value={lookupTarget}
                    onChange={(e) => setLookupTarget(e.target.value)}
                    className="w-full px-3 py-2 border border-purple-200 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent text-sm bg-white"
                    required
                >
                    <option value="">Select an object...</option>
                    {schemas.map(schema => (
                        <option key={schema.api_name} value={schema.api_name}>
                            {schema.label} ({schema.api_name})
                        </option>
                    ))}
                </select>
                <p className="mt-1 text-xs text-purple-500">Select the object to link to</p>
            </div>

            {/* Master-Detail Toggle */}
            <div className="flex items-start gap-3 p-3 bg-white/50 rounded-lg border border-purple-200">
                <input
                    type="checkbox"
                    id="isMasterDetail"
                    checked={isMasterDetail}
                    onChange={(e) => {
                        setIsMasterDetail(e.target.checked);
                        if (e.target.checked) setRequired(true); // Enforce required
                    }}
                    className="mt-1 w-4 h-4 text-purple-600 rounded border-purple-300 focus:ring-purple-500"
                />
                <label htmlFor="isMasterDetail" className="text-sm">
                    <span className="font-medium text-purple-900">Master-Detail Relationship</span>
                    <p className="text-xs text-purple-700 mt-0.5">
                        If checked, this field is <strong>Required</strong>. Deleting the parent record will automatically delete this record (Cascade Delete).
                    </p>
                </label>
            </div>
        </div>
    );
};
