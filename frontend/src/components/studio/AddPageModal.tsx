import React, { useState } from 'react';
import { X, Database, LayoutDashboard, Globe, ArrowRight } from 'lucide-react';
import * as Icons from 'lucide-react';
import { IconPicker } from '../pickers/IconPicker';
import { UI_CONFIG } from '../../core/constants/EnvironmentConfig';

interface AddPageModalProps {
    onClose: () => void;
    onAddObject: (objectDef: { label: string; apiName: string; icon: string }) => Promise<boolean>;
    onAddDashboard: (dashboardDef: { label: string; icon: string }) => Promise<boolean>;
    onAddWebLink: (webLinkDef: { label: string; url: string; icon: string }) => Promise<boolean>;
}

type PageType = 'object' | 'dashboard' | 'web';

export const AddPageModal: React.FC<AddPageModalProps> = ({ onClose, onAddObject, onAddDashboard, onAddWebLink }) => {
    const [step, setStep] = useState<'type' | 'details'>('type');
    const [pageType, setPageType] = useState<PageType | null>(null);
    const [label, setLabel] = useState('');
    const [apiName, setApiName] = useState('');
    const [url, setUrl] = useState(''); // New state for URL
    const [icon, setIcon] = useState('LayoutDashboard');
    const [showIconPicker, setShowIconPicker] = useState(false);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleLabelChange = (value: string) => {
        setLabel(value);
        // Only auto-generate API name for Objects if enabled in config
        if (pageType === 'object' && UI_CONFIG.ENABLE_AUTO_FILL_API_NAME) {
            const generated = value
                .toLowerCase()
                .replace(/[^a-z0-9]+/g, '_')
                .replace(/^_|_$/g, '');
            setApiName(generated);
        }
    };

    const handleCreate = async () => {
        if (!label.trim()) {
            setError('Label is required');
            return;
        }
        if (pageType === 'object' && !apiName.trim()) {
            setError('API Name is required');
            return;
        }
        if (pageType === 'web' && !url.trim()) {
            setError('URL is required');
            return;
        }

        setSaving(true);
        setError(null);

        try {
            if (pageType === 'object') {
                await onAddObject({ label: label.trim(), apiName: apiName.trim(), icon });
            } else if (pageType === 'dashboard') {
                await onAddDashboard({ label: label.trim(), icon });
            } else if (pageType === 'web') {
                await onAddWebLink({ label: label.trim(), url: url.trim(), icon });
            }
            onClose();
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to create');
        } finally {
            setSaving(false);
        }
    };

    const IconComponent = Icons[icon as keyof typeof Icons] as React.ComponentType<{ size?: number | string; className?: string }> || Database;

    const pageTypes = [
        {
            type: 'object' as const,
            icon: Database,
            label: 'Object',
            description: 'Create a new data table with fields',
            color: 'blue',
        },
        {
            type: 'dashboard' as const,
            icon: LayoutDashboard,
            label: 'Dashboard',
            description: 'Add charts and metrics',
            color: 'purple',
            disabled: false,
        },
        {
            type: 'web' as const,
            icon: Globe,
            label: 'Web Link',
            description: 'Link to an external page',
            color: 'emerald',
            disabled: false, // Enabled
        },
    ];

    return (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-[100]">
            <div className="bg-white rounded-xl shadow-2xl w-full max-w-md overflow-hidden">
                {/* Header */}
                <div className="flex items-center justify-between px-5 py-4 border-b">
                    <h2 className="text-lg font-semibold text-slate-800">
                        {step === 'type' ? 'Add Page' : `New ${pageType === 'object' ? 'Object' : (pageType === 'web' ? 'Web Link' : 'Page')}`}
                    </h2>
                    <button onClick={onClose} className="text-slate-400 hover:text-slate-600">
                        <X size={20} />
                    </button>
                </div>

                {/* Content */}
                <div className="p-5">
                    {step === 'type' && (
                        <div className="space-y-3">
                            {pageTypes.map(pt => (
                                <button
                                    key={pt.type}
                                    onClick={() => {
                                        if (!pt.disabled) {
                                            setPageType(pt.type);
                                            setStep('details');
                                            // Set default icon
                                            if (pt.type === 'dashboard') setIcon('LayoutDashboard');
                                            if (pt.type === 'object') setIcon('Database');
                                            if (pt.type === 'web') setIcon('Globe');
                                        }
                                    }}
                                    disabled={pt.disabled}
                                    className={`w-full flex items-center gap-4 p-4 border rounded-xl text-left transition-all ${pt.disabled
                                        ? 'opacity-50 cursor-not-allowed bg-slate-50'
                                        : 'hover:border-blue-300 hover:bg-blue-50/50 cursor-pointer'
                                        }`}
                                >
                                    <div className={`w-12 h-12 rounded-xl flex items-center justify-center bg-${pt.color}-100`}>
                                        <pt.icon size={24} className={`text-${pt.color}-600`} />
                                    </div>
                                    <div className="flex-1">
                                        <div className="font-medium text-slate-800">{pt.label}</div>
                                        <div className="text-sm text-slate-500">{pt.description}</div>
                                    </div>
                                    {!pt.disabled && <ArrowRight size={18} className="text-slate-400" />}
                                    {pt.disabled && (
                                        <span className="text-xs px-2 py-1 bg-slate-200 text-slate-500 rounded">Coming Soon</span>
                                    )}
                                </button>
                            ))}
                        </div>
                    )}

                    {step === 'details' && (
                        <div className="space-y-4">
                            {error && (
                                <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
                                    {error}
                                </div>
                            )}

                            {/* Icon + Label Row */}
                            <div className="flex gap-3">
                                <button
                                    type="button"
                                    onClick={() => setShowIconPicker(true)}
                                    className="w-14 h-14 rounded-xl bg-slate-100 hover:bg-slate-200 flex items-center justify-center flex-shrink-0 transition-colors"
                                >
                                    <IconComponent size={24} className="text-slate-600" />
                                </button>
                                <div className="flex-1">
                                    <label className="block text-sm font-medium text-slate-700 mb-1">Label</label>
                                    <input
                                        type="text"
                                        value={label}
                                        onChange={(e) => handleLabelChange(e.target.value)}
                                        placeholder={pageType === 'dashboard' ? "e.g. Sales Overview" : "e.g. Project, Task"}
                                        className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                                        autoFocus
                                    />
                                </div>
                            </div>

                            {/* URL Input - Only for Web Links */}
                            {pageType === 'web' && (
                                <div>
                                    <label className="block text-sm font-medium text-slate-700 mb-1">URL</label>
                                    <input
                                        type="url"
                                        value={url}
                                        onChange={(e) => setUrl(e.target.value)}
                                        placeholder="https://example.com"
                                        className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
                                    />
                                    <p className="text-xs text-slate-500 mt-1">
                                        External link to open in a new tab.
                                    </p>
                                </div>
                            )}

                            {/* API Name - Only for Objects */}
                            {pageType === 'object' && (
                                <div>
                                    <label className="block text-sm font-medium text-slate-700 mb-1">
                                        API Name
                                        <span className="text-slate-400 font-normal"> (auto-generated)</span>
                                    </label>
                                    <input
                                        type="text"
                                        value={apiName}
                                        onChange={(e) => setApiName(e.target.value.toLowerCase().replace(/[^a-z0-9_]/g, ''))}
                                        placeholder="project"
                                        className="w-full px-3 py-2 border rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 font-mono text-sm"
                                    />
                                    <p className="text-xs text-slate-500 mt-1">
                                        Used in URLs and API. Cannot be changed later.
                                    </p>
                                </div>
                            )}
                        </div>
                    )}
                </div>

                {/* Footer */}
                <div className="flex justify-between items-center px-5 py-4 border-t bg-slate-50">
                    {step === 'details' ? (
                        <>
                            <button
                                onClick={() => setStep('type')}
                                className="px-4 py-2 text-slate-600 hover:bg-slate-200 rounded-lg text-sm"
                            >
                                Back
                            </button>
                            <button
                                onClick={handleCreate}
                                disabled={saving || !label.trim()}
                                className="px-5 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 text-sm font-medium"
                            >
                                {saving ? 'Creating...' : `Create ${pageType === 'object' ? 'Object' : 'Page'}`}
                            </button>
                        </>
                    ) : (
                        <>
                            <div />
                            <button
                                onClick={onClose}
                                className="px-4 py-2 text-slate-600 hover:bg-slate-200 rounded-lg text-sm"
                            >
                                Cancel
                            </button>
                        </>
                    )}
                </div>
            </div>

            {/* Icon Picker */}
            {showIconPicker && (
                <IconPicker
                    value={icon}
                    onChange={(newIcon) => {
                        setIcon(newIcon);
                        setShowIconPicker(false);
                    }}
                    onClose={() => setShowIconPicker(false)}
                />
            )}
        </div>
    );
};
