import { useState, useEffect } from 'react';
import { agentApi } from '../../infrastructure/api/agent';

const STORAGE_KEY_FILES = 'nexus_ai_active_files';
const STORAGE_KEY_TOKENS = 'nexus_ai_total_tokens';

export function useAIContext() {
    const [activeFiles, setActiveFiles] = useState<{ path: string, tokenSize: number }[]>([]);
    const [totalTokens, setTotalTokens] = useState(0);

    // Load persisted state
    useEffect(() => {
        try {
            const savedFiles = localStorage.getItem(STORAGE_KEY_FILES);
            const savedTokens = localStorage.getItem(STORAGE_KEY_TOKENS);

            if (savedFiles) setActiveFiles(JSON.parse(savedFiles));
            if (savedTokens) setTotalTokens(parseInt(savedTokens, 10));
        } catch (e) {
            console.error('Failed to load persisted AI context:', e);
        }
    }, []);

    // Sync with backend on mount
    useEffect(() => {
        const loadContext = async () => {
            try {
                const context = await agentApi.getContext();
                // We prefer backend state for files as it ensures session validity
                const files = context.items.map(item => ({
                    path: item.path,
                    tokenSize: item.token_size
                }));
                setActiveFiles(files); // This will trigger the persistence effect below
                setTotalTokens(context.total_tokens);
            } catch (err) {
                console.error("Failed to load context:", err);
            }
        };
        loadContext();
    }, []);

    // Persist state
    useEffect(() => {
        localStorage.setItem(STORAGE_KEY_FILES, JSON.stringify(activeFiles));
    }, [activeFiles]);

    useEffect(() => {
        localStorage.setItem(STORAGE_KEY_TOKENS, totalTokens.toString());
    }, [totalTokens]);

    const refreshContext = async () => {
        try {
            const context = await agentApi.getContext();
            setActiveFiles(context.items.map(item => ({
                path: item.path,
                tokenSize: item.token_size
            })));
            setTotalTokens(context.total_tokens);
        } catch (err) {
            console.error('Failed to refresh context:', err);
        }
    };

    // Explicit mutators if needed from UI side (though usually driven by backend tools)
    const updateFilesFromToolResult = (result: string) => {
        // Parse heuristics: "Active Context (1 files, ~50 tokens): ... - /path (~10 tokens)"
        const regex = /- (.*?) \(~(\d+) tokens\)/g;
        let match;
        const newFiles = [];
        let newTotal = 0;
        while ((match = regex.exec(result)) !== null) {
            const tokenSize = parseInt(match[2], 10);
            newFiles.push({ path: match[1], tokenSize });
            newTotal += tokenSize;
        }
        if (newFiles.length > 0 || result.includes("0 files")) {
            setActiveFiles(newFiles);
            setTotalTokens(newTotal);
        }
    };

    return {
        activeFiles,
        totalTokens,
        refreshContext,
        setActiveFiles,
        setTotalTokens,
        updateFilesFromToolResult
    };
}
