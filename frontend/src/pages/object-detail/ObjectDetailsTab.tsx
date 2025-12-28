import React from 'react';
import { ObjectMetadata } from '../../types';
import { metadataAPI } from '../../infrastructure/api/metadata';

interface ObjectDetailsTabProps {
    metadata: ObjectMetadata;
    refresh: () => Promise<void>;
}

export const ObjectDetailsTab: React.FC<ObjectDetailsTabProps> = ({ metadata, refresh }) => {
    return (
        <div className="space-y-6">
            <div className="bg-white/80 backdrop-blur-xl border border-white/20 rounded-2xl shadow-xl p-6">
                <h3 className="text-lg font-medium text-slate-900 mb-4">Object Details</h3>
                <dl className="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
                    <div>
                        <dt className="text-sm font-medium text-slate-500">Label</dt>
                        <dd className="mt-1 text-sm text-slate-900">{metadata.label}</dd>
                    </div>
                    <div>
                        <dt className="text-sm font-medium text-slate-500">Plural Label</dt>
                        <dd className="mt-1 text-sm text-slate-900">{metadata.plural_label}</dd>
                    </div>
                    <div>
                        <dt className="text-sm font-medium text-slate-500">API Name</dt>
                        <dd className="mt-1 text-sm text-slate-900 font-mono">{metadata.api_name}</dd>
                    </div>
                    <div>
                        <dt className="text-sm font-medium text-slate-500">Type</dt>
                        <dd className="mt-1 text-sm text-slate-900">{metadata.is_system ? 'Standard' : 'Custom'}</dd>
                    </div>
                    <div className="sm:col-span-2">
                        <dt className="text-sm font-medium text-slate-500">Description</dt>
                        <dd className="mt-1 text-sm text-slate-900">{metadata.description || '-'}</dd>
                    </div>
                </dl>
            </div>

            <div className="bg-white/80 backdrop-blur-xl border border-white/20 rounded-2xl shadow-xl p-6">
                <h3 className="text-lg font-medium text-slate-900 mb-4">Path Settings</h3>
                <div className="max-w-xl">
                    <label className="block text-sm font-medium text-slate-700 mb-1">
                        Path Field
                    </label>
                    <p className="text-sm text-slate-500 mb-3">
                        Select a picklist field to use for the Path component (visual progress bar).
                    </p>
                    <div className="flex gap-2">
                        <select
                            className="flex-1 px-3 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                            value={metadata.path_field || ''}
                            onChange={async (e) => {
                                const path_field = e.target.value;
                                try {
                                    await metadataAPI.updateSchema(metadata.api_name, { path_field });
                                    // Refresh metadata
                                    window.location.reload();
                                } catch {
                                    // Path field update failure - page will reload anyway
                                }
                            }}
                        >
                            <option value="">-- None --</option>
                            {(metadata.fields || [])
                                .filter(f => f.type === 'Picklist')
                                .map(f => (
                                    <option key={f.api_name} value={f.api_name}>
                                        {f.label} ({f.api_name})
                                    </option>
                                ))}
                        </select>
                    </div>
                </div>
            </div>
        </div>
    );
};
