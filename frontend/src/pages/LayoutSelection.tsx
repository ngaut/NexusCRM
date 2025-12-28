import React, { useState } from 'react';
import { useSchemas } from '../core/hooks/useMetadata';
import { useNavigate } from 'react-router-dom';
import { Layout, Search, Edit } from 'lucide-react';

export const LayoutSelection: React.FC = () => {
    const { schemas, loading, error } = useSchemas();
    const [searchQuery, setSearchQuery] = useState('');
    const navigate = useNavigate();

    const filteredSchemas = schemas.filter(schema =>
        schema.label.toLowerCase().includes(searchQuery.toLowerCase()) ||
        schema.api_name.toLowerCase().includes(searchQuery.toLowerCase())
    );

    if (loading) return <div className="p-8 text-center">Loading objects...</div>;
    if (error) return <div className="p-8 text-center text-red-500">Error: {error.message}</div>;

    return (
        <div className="max-w-7xl mx-auto p-6">
            <div className="flex items-center justify-between mb-8">
                <div className="flex items-center gap-4">
                    <div className="p-3 bg-green-100 rounded-lg">
                        <Layout className="text-green-600" size={32} />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold text-slate-800">Layout Editor</h1>
                        <p className="text-slate-500">Customize page layouts and record details for each object.</p>
                    </div>
                </div>
            </div>

            <div className="mb-6">
                <div className="relative max-w-md">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" size={18} />
                    <input
                        type="text"
                        placeholder="Search objects..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="w-full pl-10 pr-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-green-500 focus:border-green-500"
                    />
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                {filteredSchemas.map(schema => (
                    <button
                        key={schema.api_name}
                        onClick={() => navigate(`/setup/objects/${schema.api_name}/layout`)}
                        className="p-6 bg-white rounded-lg border-2 border-slate-200 hover:border-green-500 hover:shadow-lg transition-all text-left group"
                    >
                        <div className="flex items-start justify-between mb-3">
                            <div className="flex items-center gap-3">
                                <div className="p-2 bg-slate-100 rounded group-hover:bg-green-50 transition-colors">
                                    <Layout className="text-slate-600 group-hover:text-green-600" size={20} />
                                </div>
                                <div>
                                    <h3 className="font-semibold text-slate-900">{schema.label}</h3>
                                    <p className="text-xs text-slate-500 font-mono">{schema.api_name}</p>
                                </div>
                            </div>
                            <Edit className="text-slate-400 group-hover:text-green-600" size={18} />
                        </div>
                        {schema.description && (
                            <p className="text-sm text-slate-600 line-clamp-2">{schema.description}</p>
                        )}
                        {schema.is_system && (
                            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-slate-100 text-slate-700 mt-2">
                                Standard
                            </span>
                        )}
                    </button>
                ))}
            </div>

            {filteredSchemas.length === 0 && (
                <div className="text-center py-12 text-slate-500">
                    <Layout size={48} className="mx-auto text-slate-300 mb-4" />
                    <p>No objects found matching your search.</p>
                </div>
            )}
        </div>
    );
};
