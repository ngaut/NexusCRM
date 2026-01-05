import React, { useState, useEffect } from 'react';
import { X, Sparkles, Database, PanelLeftClose, PanelLeft } from 'lucide-react';
import { ContextPanel } from './ai/ContextPanel';
import { InputArea } from './ai/InputArea';
import { MessageList } from './ai/MessageList';
import { ConversationSidebar } from './ai/ConversationSidebar';
import { Z_LAYERS } from '../core/constants/zIndex';
import { useResizable } from '../core/hooks/useResizable';
import { useAIStream } from '../core/hooks/useAIStream';
import { useAIContext } from '../core/hooks/useAIContext';
import { ChatMessage } from '../infrastructure/api/agent';

export interface AIAssistantProps {
  isOpen: boolean;
  onClose: () => void;
}

export function AIAssistant({ isOpen, onClose }: AIAssistantProps) {
  const [input, setInput] = useState('');
  const [isContextPanelOpen, setIsContextPanelOpen] = useState(false);

  const {
    width: panelWidth,
    startResizing: handleResizeStart
  } = useResizable({
    initialWidth: 440,
    minWidth: 360,
    maxWidth: window.innerWidth - 100,
    direction: 'left'
  });

  const {
    width: contextPanelWidth,
    startResizing: handleContextResizeStart
  } = useResizable({
    initialWidth: 320,
    minWidth: 320,
    maxWidth: 800,
    direction: 'left'
  });

  const {
    messages,
    setMessages,
    processSteps,
    isLoading,
    streamingContent,
    isProcessExpanded,
    setIsProcessExpanded,
    sendMessage,
    cancelStream,
    clearChat,
    compactMessages,
    conversationTokens,
    summaryInfo,
    currentConversationId,
    conversations,
    isSidebarOpen,
    setIsSidebarOpen,
    newChat,
    selectConversation,
    deleteConversation,
  } = useAIStream();

  const {
    activeFiles,
    totalTokens,
    refreshContext,
    updateFilesFromToolResult
  } = useAIContext();

  useEffect(() => {
    const lastStep = [...processSteps].reverse().find(s => s.type === 'tool_result' && !s.isError);
    if (lastStep && lastStep.toolName === 'context_list' && lastStep.toolResult) {
      updateFilesFromToolResult(lastStep.toolResult);
    }
  }, [processSteps, updateFilesFromToolResult]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim() || isLoading) return;

    const command = input.trim();
    if (command.startsWith('/')) {
      const [cmd, ...args] = command.split(' ');

      switch (cmd) {
        case '/clear':
          clearChat();
          setInput('');
          return;
        case '/add':
          sendMessage(`Please add these files to the context: ${args.join(' ')}. Use the context_add tool. After adding, please list the context.`);
          setInput('');
          return;
        case '/remove':
          sendMessage(`Please remove these files from the context: ${args.join(' ')}. Use the context_remove tool. After removing, please list the context.`);
          setInput('');
          return;
        case '/list':
          sendMessage('Please list the files currently in the context using context_list.');
          setInput('');
          return;
        case '/compact':
          if (messages.length === 0) {
            setMessages(prev => [...prev, {
              role: 'assistant',
              content: '⚠️ No conversation history to compact.'
            } as ChatMessage]);
          } else {
            await compactMessages(args.join(' '));
          }
          setInput('');
          return;
        case '/help':
          setMessages(prev => [...prev, {
            role: 'user', content: command
          } as ChatMessage, {
            role: 'assistant', content: `**Available Commands:**
  - \`/add <file>\`: Add files to context
  - \`/remove <file>\`: Remove files from context
  - \`/list\`: List active files
  - \`/compact [keep X]\`: Compact conversation history
  - \`/clear\`: Clear chat history
  - \`/help\`: Show this help`
          } as ChatMessage]);
          setInput('');
          return;
        default:
          break;
      }
    }

    sendMessage(input);
    setInput('');
  };

  const handlePromptSelect = (prompt: string) => {
    setInput(prompt);
  };

  if (!isOpen) return null;

  const sidebarWidth = isSidebarOpen ? 220 : 36;
  const totalWidth = sidebarWidth + panelWidth + (isContextPanelOpen ? contextPanelWidth : 0);

  return (
    <div
      className="fixed right-0 top-0 h-full bg-white/98 backdrop-blur-xl shadow-2xl border-l border-slate-200/50 flex animate-fade-in"
      style={{ width: `${totalWidth}px`, zIndex: Z_LAYERS.PANEL }}
    >
      {/* 1. Conversation Sidebar (left-most) */}
      <ConversationSidebar
        conversations={conversations}
        currentConversationId={currentConversationId}
        isOpen={isSidebarOpen}
        onToggle={() => setIsSidebarOpen(!isSidebarOpen)}
        onSelectConversation={selectConversation}
        onNewChat={newChat}
        onDeleteConversation={deleteConversation}
        isLoading={isLoading}
      />

      {/* 2. Main Chat Area (center) */}
      <div className="flex flex-col h-full flex-1 relative border-l border-slate-200 max-w-full overflow-hidden">
        <div
          onMouseDown={handleResizeStart}
          className="absolute left-0 top-0 h-full w-2 cursor-ew-resize hover:bg-indigo-500/20 transition-colors z-10 group -ml-1"
        >
          <div className="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-12 bg-slate-300 rounded-full opacity-0 group-hover:opacity-100 transition-opacity" />
        </div>

        <div className="flex items-center justify-between px-4 py-3 border-b border-slate-200/80 bg-white">
          <div className="flex items-center gap-3">
            <button
              onClick={() => setIsSidebarOpen(!isSidebarOpen)}
              className="p-1.5 rounded-lg hover:bg-slate-100 text-slate-500 hover:text-slate-700 transition-colors"
              title={isSidebarOpen ? 'Collapse history' : 'Expand history'}
            >
              {isSidebarOpen ? <PanelLeftClose size={16} /> : <PanelLeft size={16} />}
            </button>
            <div className="bg-gradient-to-br from-indigo-600 via-purple-600 to-pink-500 p-2 rounded-xl shadow-md shadow-purple-500/20">
              <Sparkles size={20} className="text-white" />
            </div>
            <div>
              <h3 className="font-semibold text-slate-800">Nexus AI</h3>
              <p className="text-[11px] text-slate-400">Your CRM assistant</p>
            </div>
          </div>
          <div className="flex items-center gap-1 flex-shrink-0">
            <button
              onClick={() => setIsContextPanelOpen(!isContextPanelOpen)}
              className={`p-2 rounded-lg transition-all ${isContextPanelOpen ? 'bg-indigo-100 text-indigo-600' : 'hover:bg-slate-100 text-slate-500 hover:text-slate-700'}`}
              title="Toggle Context Panel"
            >
              <Database size={18} />
            </button>

            <button
              onClick={onClose}
              className="p-2 hover:bg-slate-100 rounded-lg transition-all text-slate-500 hover:text-slate-700 hover:bg-red-50 hover:text-red-600"
              title="Close"
            >
              <X size={18} />
            </button>
          </div>
        </div>

        <MessageList
          messages={messages}
          isLoading={isLoading}
          streamingContent={streamingContent}
          processSteps={processSteps}
          isProcessExpanded={isProcessExpanded}
          setIsProcessExpanded={setIsProcessExpanded}
          onPromptSelect={handlePromptSelect}
        />

        <InputArea
          input={input}
          isLoading={isLoading}
          setInput={setInput}
          handleSubmit={handleSubmit}
          handleCancel={cancelStream}
        />
      </div>

      {/* 3. Context Panel (right-most, when open) */}
      {isContextPanelOpen && (
        <div
          onMouseDown={handleContextResizeStart}
          className="absolute right-0 top-0 h-full w-2 cursor-ew-resize hover:bg-indigo-500/20 transition-colors group"
          style={{ zIndex: Z_LAYERS.MODAL, right: contextPanelWidth - 4 }}
        >
          <div className="absolute left-0 top-1/2 -translate-y-1/2 w-1 h-12 bg-indigo-300 rounded-full opacity-0 group-hover:opacity-100 transition-opacity shadow-sm" />
        </div>
      )}

      <ContextPanel
        width={contextPanelWidth}
        files={activeFiles}
        messages={messages}
        toolsList={[
          'list_objects', 'describe_object', 'query_object',
          'create_record', 'update_record', 'delete_record',
          'create_dashboard', 'context_add', 'context_remove',
          'context_list', 'context_clear'
        ]}
        systemPromptTokens={200}
        toolsTokens={1800}
        totalTokens={totalTokens}
        conversationTokens={conversationTokens}
        maxTokens={128000}
        isOpen={isContextPanelOpen}
        hasSummary={summaryInfo.hasSummary}
        tokensSaved={summaryInfo.tokensSaved}
        isLoading={isLoading}
        onClose={() => setIsContextPanelOpen(false)}
        onRemoveFile={(path) => setInput(`/remove ${path}`)}
        onRemoveMessage={(index) => setMessages(prev => prev.filter((_, i) => i !== index))}
        onAddFiles={() => setInput('/add ')}
        onCompact={() => compactMessages()}
        onRefresh={refreshContext}
      />
    </div>
  );
}
