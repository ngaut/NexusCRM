import React, { useMemo } from 'react';
import { FolderOpen, Plus, Trash2, FileCode, FileJson, FileText, FileType } from 'lucide-react';

interface ContextFile {
    path: string;
    tokenSize: number;
}

interface ContextFileListProps {
    files: ContextFile[];
    onAddFiles?: () => void;
    onRemoveFile: (path: string) => void;
}

// Common file suggestions
const SUGGESTED_FILES = [
    { name: 'AIAssistant.tsx', path: '/add AIAssistant.tsx' },
    { name: 'package.json', path: '/add package.json' },
    { name: 'types.ts', path: '/add types.ts' },
];

function getFileIcon(path: string) {
    const ext = path.split('.').pop()?.toLowerCase() || '';
    switch (ext) {
        case 'tsx':
        case 'jsx':
            return <FileCode size={14} className="text-blue-500" />;
        case 'ts':
        case 'js':
            return <FileCode size={14} className="text-yellow-500" />;
        case 'json':
            return <FileJson size={14} className="text-amber-500" />;
        case 'md':
        case 'txt':
            return <FileText size={14} className="text-slate-500" />;
        case 'go':
            return <FileCode size={14} className="text-cyan-500" />;
        default:
            return <FileType size={14} className="text-slate-400" />;
    }
}

export const ContextFileList: React.FC<ContextFileListProps> = ({
    files,
    onAddFiles,
    onRemoveFile
}) => {
    // Calculate max file size for heatmap
    const maxFileTokenSize = useMemo(() => {
        return Math.max(...files.map(f => f.tokenSize), 1);
    }, [files]);

    if (files.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center h-full px-6 py-8 text-center">
                <div className="w-12 h-12 rounded-xl bg-slate-100 flex items-center justify-center mb-3">
                    <FolderOpen size={24} className="text-slate-400" />
                </div>
                <h4 className="text-sm font-medium text-slate-700 mb-1">No files pinned</h4>
                <p className="text-xs text-slate-400 mb-4 max-w-[200px]">
                    Add files to give the AI context
                </p>

                {onAddFiles && (
                    <button
                        onClick={onAddFiles}
                        className="flex items-center gap-2 px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white text-sm font-medium rounded-lg shadow-sm hover:shadow transition-all"
                    >
                        <Plus size={16} />
                        Add Files
                    </button>
                )}

                <div className="mt-4 w-full">
                    <div className="text-[10px] text-slate-400 uppercase tracking-wider mb-2">Quick add</div>
                    <div className="flex flex-wrap justify-center gap-1.5">
                        {SUGGESTED_FILES.map((file) => (
                            <button
                                key={file.name}
                                onClick={() => onAddFiles?.()}
                                className="px-2 py-1 bg-slate-100 hover:bg-slate-200 text-slate-600 text-[11px] rounded-md transition-colors"
                            >
                                {file.name}
                            </button>
                        ))}
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="p-2 space-y-1.5">
            <div className="flex items-center justify-between px-2 py-1">
                <span className="text-[10px] text-slate-400 uppercase tracking-wider font-medium">
                    Files ({files.length})
                </span>
                {files.length > 0 && (
                    <button
                        onClick={() => files.forEach(f => onRemoveFile(f.path))}
                        className="text-[10px] text-slate-400 hover:text-red-500 transition-colors"
                        title="Remove all files"
                    >
                        Clear all
                    </button>
                )}
            </div>

            {files.map((file, idx) => (
                <div
                    key={idx}
                    className="relative flex items-center gap-2 px-3 py-2.5 bg-slate-50 hover:bg-slate-100 rounded-lg transition-colors group overflow-hidden"
                >
                    {/* Heatmap Bar */}
                    <div
                        className="absolute left-0 top-0 bottom-0 bg-emerald-100/40 pointer-events-none transition-all duration-500"
                        style={{ width: `${(file.tokenSize / maxFileTokenSize) * 100}%` }}
                    />

                    <div className="relative z-10 flex items-center gap-2 flex-1 min-w-0">
                        {getFileIcon(file.path)}
                        <div className="flex-1 min-w-0">
                            <div className="text-xs font-medium text-slate-700 truncate" title={file.path}>
                                {file.path.split('/').pop()}
                            </div>
                            <div className="text-[10px] text-slate-400">
                                ~{file.tokenSize.toLocaleString()} tokens
                            </div>
                        </div>
                    </div>

                    <button
                        onClick={() => onRemoveFile(file.path)}
                        className="relative z-10 p-1 rounded text-slate-300 hover:text-red-500 hover:bg-red-50 transition-all opacity-60 group-hover:opacity-100"
                        title="Remove from context"
                    >
                        <Trash2 size={14} />
                    </button>
                </div>
            ))}
        </div>
    );
};
