import React from 'react';

interface PicklistConfigProps {
    picklistValues: string;
    onChange: (value: string) => void;
}

export const PicklistConfig: React.FC<PicklistConfigProps> = ({ picklistValues, onChange }) => {
    return (
        <div className="p-4 bg-amber-50 rounded-lg border border-amber-100">
            <label className="block text-sm font-medium text-amber-800 mb-1.5">
                Picklist Values <span className="text-red-500">*</span>
            </label>
            <textarea
                value={picklistValues}
                onChange={(e) => onChange(e.target.value)}
                placeholder="Enter each option on a new line:&#10;Option 1&#10;Option 2&#10;Option 3"
                rows={4}
                className="w-full px-3 py-2 border border-amber-200 rounded-lg focus:ring-2 focus:ring-amber-500 focus:border-transparent text-sm"
                required
            />
        </div>
    );
};
