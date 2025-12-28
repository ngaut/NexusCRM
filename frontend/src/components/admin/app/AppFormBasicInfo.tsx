import React from 'react';
import type { AppConfig } from '../../../types';

interface AppFormBasicInfoProps {
    formData: AppConfig;
    isEditMode: boolean;
    onChange: (updates: Partial<AppConfig>) => void;
}

export const AppFormBasicInfo: React.FC<AppFormBasicInfoProps> = ({ formData, isEditMode, onChange }) => {
    return (
        <div className="space-y-5">
            <div className="grid grid-cols-2 gap-4">
                {/* ID */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1">
                        App ID <span className="text-red-500">*</span>
                    </label>
                    <input
                        type="text"
                        value={formData.id}
                        onChange={(e) => onChange({ id: e.target.value })}
                        className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="sales"
                        required
                        disabled={isEditMode}
                    />
                </div>

                {/* Label */}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1">
                        Label <span className="text-red-500">*</span>
                    </label>
                    <input
                        type="text"
                        value={formData.label}
                        onChange={(e) => onChange({ label: e.target.value })}
                        className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                        placeholder="Sales"
                        required
                    />
                </div>
            </div>

            {/* Description */}
            <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">
                    Description
                </label>
                <textarea
                    value={formData.description}
                    onChange={(e) => onChange({ description: e.target.value })}
                    className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    rows={2}
                    placeholder="Sales Cloud application"
                />
            </div>
        </div>
    );
};
