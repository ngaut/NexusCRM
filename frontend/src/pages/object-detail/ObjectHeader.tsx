import React from 'react';
import { Link } from 'react-router-dom';
import { ArrowLeft, Database, Edit } from 'lucide-react';
import { ObjectMetadata } from '../../types';

interface ObjectHeaderProps {
    metadata: ObjectMetadata;
    onEditObject: () => void;
}

export const ObjectHeader: React.FC<ObjectHeaderProps> = ({ metadata, onEditObject }) => {
    return (
        <div className="mb-6">
            <Link to="/setup/objects" className="text-blue-600 hover:underline flex items-center gap-2 mb-4">
                <ArrowLeft size={16} />
                Back to Object Manager
            </Link>

            <div className="flex items-center justify-between mb-6">
                <div className="flex items-center gap-4">
                    <div className="p-3 bg-blue-100 rounded-lg">
                        <Database className="text-blue-600" size={32} />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold text-slate-800">{metadata.label}</h1>
                        <div className="flex items-center gap-2 text-sm text-slate-500">
                            <span className="font-mono bg-slate-100 px-2 py-0.5 rounded">{metadata.api_name}</span>
                            <span>•</span>
                            <span>{metadata.is_system ? 'Standard Object' : 'Custom Object'}</span>
                        </div>
                    </div>
                </div>
                {!metadata.is_system && (
                    <button
                        onClick={onEditObject}
                        className="flex items-center gap-2 px-4 py-2 border border-slate-300 text-slate-700 rounded-lg hover:bg-slate-50 font-medium transition-colors"
                    >
                        <Edit size={16} />
                        Edit Object
                    </button>
                )}
            </div>
        </div>
    );
};
