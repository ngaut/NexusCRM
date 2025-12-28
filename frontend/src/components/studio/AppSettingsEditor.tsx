import React from 'react';
import type { AppConfig } from '../../types';
import { getColorClasses } from '../../core/utils/colorClasses';

interface AppSettingsEditorProps {
    app: AppConfig;
    isNewApp: boolean;
    onChange: (updates: Partial<AppConfig>) => void;
}

export const AppSettingsEditor: React.FC<AppSettingsEditorProps> = ({
    app,
    isNewApp,
    onChange,
}) => {
    const handleIdChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const value = e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, '');
        onChange({ id: value });
    };

    return (
        <div className="bg-white rounded-xl border border-slate-200 p-8 max-w-2xl mx-auto shadow-sm">
            <h2 className="text-xl font-bold text-slate-800 mb-6">App Settings</h2>
            <div className="space-y-4">
                {isNewApp && (
                    <div>
                        <label className="block text-sm font-medium text-slate-700 mb-1">App ID</label>
                        <input
                            type="text"
                            value={app.id}
                            onChange={handleIdChange}
                            className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 font-mono text-sm"
                            placeholder="my_app_id"
                        />
                        <p className="text-xs text-slate-500 mt-1">Unique identifier, cannot be changed later.</p>
                    </div>
                )}
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1">App Name</label>
                    <input
                        type="text"
                        value={app.label}
                        onChange={(e) => onChange({ label: e.target.value })}
                        className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                    />
                </div>
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1">Description</label>
                    <textarea
                        value={app.description}
                        onChange={(e) => onChange({ description: e.target.value })}
                        className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500"
                        rows={3}
                    />
                </div>
                <div>
                    <label className="block text-sm font-medium text-slate-700 mb-1">Theme Color</label>
                    <div className="flex gap-2">
                        {['blue', 'purple', 'emerald', 'orange', 'rose', 'slate'].map(color => (
                            <button
                                key={color}
                                onClick={() => onChange({ color })}
                                className={`w-8 h-8 rounded-full border-2 ${app.color === color ? 'border-slate-600' : 'border-transparent'
                                    } ${getColorClasses(color).bg}`}
                            />
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
};
