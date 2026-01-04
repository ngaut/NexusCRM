import React, { useState } from 'react';
import { createPortal } from 'react-dom';
import { Play, X, AlertCircle, CheckCircle } from 'lucide-react';
import { flowsApi, ExecuteFlowRequest, ExecuteFlowResponse } from '../../infrastructure/api/flows';
import { Button } from '../ui/Button';
import { useSuccessToast, useErrorToast } from '../ui/Toast';
import { useSchemas } from '../../core/hooks/useMetadata';
import type { Flow } from '../../infrastructure/api/flows';

interface FlowExecutionModalProps {
    isOpen: boolean;
    onClose: () => void;
    flow: Flow;
}

export function FlowExecutionModal({
    isOpen,
    onClose,
    flow,
}: FlowExecutionModalProps) {
    const showSuccess = useSuccessToast();
    const showError = useErrorToast();

    const [recordId, setRecordId] = useState('');
    const [objectApiName, setObjectApiName] = useState(flow.trigger_object || '');
    const [executing, setExecuting] = useState(false);
    const [result, setResult] = useState<ExecuteFlowResponse | null>(null);
    const [error, setError] = useState<string | null>(null);
    const { schemas } = useSchemas();

    if (!isOpen) return null;

    const handleExecute = async () => {
        setExecuting(true);
        setError(null);
        setResult(null);

        try {
            const request: ExecuteFlowRequest = {};
            if (recordId.trim()) {
                request.record_id = recordId.trim();
            }
            if (objectApiName.trim()) {
                request.object_api_name = objectApiName.trim();
            }

            const response = await flowsApi.execute(flow.id, request);
            setResult(response);
            if (response.success) {
                showSuccess('Flow executed successfully');
            }
        } catch (err) {
            const message = err instanceof Error ? err.message : 'Unknown error';
            setError(message || 'Failed to execute flow');
            showError('Failed to execute flow: ' + message);
        } finally {
            setExecuting(false);
        }
    };

    const handleClose = () => {
        setRecordId('');
        setObjectApiName(flow.trigger_object || '');
        setResult(null);
        setError(null);
        onClose();
    };

    return createPortal(
        <div className="fixed inset-0 z-[100] flex items-center justify-center">
            {/* Backdrop */}
            <div
                className="absolute inset-0 bg-black/50 backdrop-blur-sm"
                onClick={handleClose}
            />

            {/* Modal */}
            <div className="relative bg-white rounded-xl shadow-2xl w-full max-w-lg mx-4 overflow-hidden animate-fade-in">
                {/* Header */}
                <div className="px-6 py-4 border-b border-gray-200 bg-gradient-to-r from-purple-500 to-indigo-600">
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-3">
                            <div className="p-2 bg-white/20 rounded-lg">
                                <Play className="w-5 h-5 text-white" />
                            </div>
                            <div>
                                <h2 className="text-lg font-semibold text-white">Execute Flow</h2>
                                <p className="text-sm text-white/80">{flow.name}</p>
                            </div>
                        </div>
                        <button
                            onClick={handleClose}
                            className="p-1 rounded hover:bg-white/20 transition-colors"
                        >
                            <X className="w-5 h-5 text-white" />
                        </button>
                    </div>
                </div>

                {/* Body */}
                <div className="p-6 space-y-4">
                    {!result ? (
                        <>
                            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 text-sm text-blue-700">
                                <p><strong>Flow Action:</strong> {flow.action_type}</p>
                                <p className="mt-1"><strong>Trigger Object:</strong> {flow.trigger_object}</p>
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">
                                    Record ID (optional)
                                </label>
                                <input
                                    type="text"
                                    value={recordId}
                                    onChange={(e) => setRecordId(e.target.value)}
                                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-purple-500 transition-colors"
                                    placeholder="Enter a record ID to use as context..."
                                    disabled={executing}
                                />
                                <p className="mt-1 text-xs text-gray-500">
                                    If provided, the flow will use this record's data for field mappings.
                                </p>
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-1">
                                    Object API Name (optional)
                                </label>
                                <select
                                    value={objectApiName}
                                    onChange={(e) => setObjectApiName(e.target.value)}
                                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-purple-500 transition-colors bg-white"
                                    disabled={executing}
                                >
                                    <option value="">Use Trigger Object ({flow.trigger_object})</option>
                                    {schemas.map(schema => (
                                        <option key={schema.api_name} value={schema.api_name}>
                                            {schema.label} ({schema.api_name})
                                        </option>
                                    ))}
                                </select>
                            </div>

                            {error && (
                                <div className="bg-red-50 border border-red-200 rounded-lg p-4 flex items-start gap-2 text-red-700">
                                    <AlertCircle className="w-5 h-5 flex-shrink-0 mt-0.5" />
                                    <span className="text-sm">{error}</span>
                                </div>
                            )}
                        </>
                    ) : (
                        /* Result Display */
                        <div className={`rounded-lg p-4 ${result.success ? 'bg-green-50 border border-green-200' : 'bg-red-50 border border-red-200'}`}>
                            <div className="flex items-start gap-3">
                                {result.success ? (
                                    <CheckCircle className="w-6 h-6 text-green-600 flex-shrink-0" />
                                ) : (
                                    <AlertCircle className="w-6 h-6 text-red-600 flex-shrink-0" />
                                )}
                                <div>
                                    <h3 className={`font-semibold ${result.success ? 'text-green-800' : 'text-red-800'}`}>
                                        {result.success ? 'Flow Executed Successfully' : 'Flow Execution Failed'}
                                    </h3>
                                    <p className={`text-sm mt-1 ${result.success ? 'text-green-700' : 'text-red-700'}`}>
                                        {result.message}
                                    </p>
                                    {result.result && (
                                        <div className="mt-3 bg-white rounded p-3 text-sm">
                                            <pre className="text-gray-700 overflow-x-auto">
                                                {JSON.stringify(result.result, null, 2)}
                                            </pre>
                                        </div>
                                    )}
                                </div>
                            </div>
                        </div>
                    )}
                </div>

                {/* Footer */}
                <div className="px-6 py-4 bg-gray-50 border-t border-gray-200 flex justify-end gap-3">
                    <Button
                        variant="ghost"
                        onClick={handleClose}
                    >
                        {result ? 'Close' : 'Cancel'}
                    </Button>
                    {!result && (
                        <Button
                            variant="primary"
                            onClick={handleExecute}
                            disabled={executing}
                            icon={<Play className="w-4 h-4" />}
                        >
                            {executing ? 'Executing...' : 'Execute Flow'}
                        </Button>
                    )}
                </div>
            </div>
        </div>,
        document.body
    );
}
