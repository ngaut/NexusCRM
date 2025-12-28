
import { ChatMessage } from '../../infrastructure/api/agent';

export interface AIAssistantProps {
    isOpen: boolean;
    onClose: () => void;
}

export interface ProcessStep {
    id: string;
    type: 'thinking' | 'tool_call' | 'tool_result';
    content: string;
    toolName?: string;
    toolArgs?: string;
    toolResult?: string;
    isError?: boolean;
    isDone?: boolean;
    timestamp: Date;
}

export interface ToolBlock {
    type: 'tool_block';
    id: string;
    tools: {
        call: ProcessStep;
        result?: ProcessStep;
    }[];
    timestamp: Date;
}

export interface SummaryBlock {
    type: 'summary_block';
    id: string;
    summary: string;
    stats?: string;
    compactedAt?: string;
    timestamp: Date;
}

export type DisplayItem = ChatMessage | ToolBlock | SummaryBlock;
