import React, { useState, useEffect } from 'react';
import { Plus, Zap, AlertCircle, ShieldCheck } from 'lucide-react';
import { flowsApi } from '../infrastructure/api/flows';
import { metadataAPI } from '../infrastructure/api/metadata';
import { dataAPI } from '../infrastructure/api/data';
import { SYSTEM_TABLE_NAMES } from '../generated-schema';
import { useSchemas } from '../core/hooks/useMetadata';
import FlowBuilderModal from '../components/modals/FlowBuilderModal';
import { FlowExecutionModal } from '../components/modals/FlowExecutionModal';
import type { Flow } from '../infrastructure/api/flows';
import { ApprovalProcess } from '../types';
import { ConfirmationModal } from '../components/modals/ConfirmationModal';
import { FlowList } from '../components/flows/FlowList';
import { ApprovalProcessList } from '../components/flows/ApprovalProcessList';
import { ApprovalProcessModal } from '../components/flows/ApprovalProcessModal';

const FlowsPage: React.FC = () => {
    const [activeTab, setActiveTab] = useState<'flows' | 'approvals'>('flows');

    // --- Flow State ---
    const [flows, setFlows] = useState<Flow[]>([]);
    const [loadingFlows, setLoadingFlows] = useState(true);
    const [flowError, setFlowError] = useState<string | null>(null);
    const [showFlowModal, setShowFlowModal] = useState(false);
    const [editingFlow, setEditingFlow] = useState<Flow | null>(null);
    const [objects, setObjects] = useState<{ api_name: string, label: string }[]>([]);

    // Flow Delete State
    const [deleteFlowModalOpen, setDeleteFlowModalOpen] = useState(false);
    const [flowToDelete, setFlowToDelete] = useState<Flow | null>(null);
    const [deletingFlow, setDeletingFlow] = useState(false);

    // Flow Execution Modal State
    const [executeModalOpen, setExecuteModalOpen] = useState(false);
    const [flowToExecute, setFlowToExecute] = useState<Flow | null>(null);

    // --- Approval State ---
    const [processes, setProcesses] = useState<ApprovalProcess[]>([]);
    const [loadingApprovals, setLoadingApprovals] = useState(true);
    const [approvalError, setApprovalError] = useState<string | null>(null);
    const [showApprovalModal, setShowApprovalModal] = useState(false);
    const [editingProcess, setEditingProcess] = useState<ApprovalProcess | null>(null);

    // Approval Delete State
    const [deleteApprovalModalOpen, setDeleteApprovalModalOpen] = useState(false);
    const [processToDelete, setProcessToDelete] = useState<ApprovalProcess | null>(null);

    const { schemas } = useSchemas();

    // --- Effects ---

    useEffect(() => {
        loadFlows();
        loadObjects();
        loadApprovals();
    }, []);

    // --- Flow Logic ---

    const loadFlows = async () => {
        try {
            setLoadingFlows(true);
            const data = await flowsApi.getAll();
            setFlows(data || []);
            setFlowError(null);
        } catch (err) {
            setFlowError('Failed to load flows: ' + (err instanceof Error ? err.message : 'Unknown error'));
        } finally {
            setLoadingFlows(false);
        }
    };

    const loadObjects = async () => {
        try {
            const response = await metadataAPI.getSchemas();
            setObjects(response.schemas.map((s: { api_name: string; label: string }) => ({ api_name: s.api_name, label: s.label })));
        } catch {
            // Objects loading failure is non-critical
        }
    };

    const handleToggleStatus = async (flow: Flow) => {
        try {
            await flowsApi.toggleStatus(flow.id, flow.status);
            loadFlows();
        } catch (err) {
            setFlowError('Failed to toggle status: ' + (err instanceof Error ? err.message : 'Unknown error'));
        }
    };

    const confirmDeleteFlow = async () => {
        if (!flowToDelete) return;
        setDeletingFlow(true);
        try {
            await flowsApi.delete(flowToDelete.id);
            loadFlows();
            setDeleteFlowModalOpen(false);
            setFlowToDelete(null);
        } catch (err) {
            setFlowError('Failed to delete flow: ' + (err instanceof Error ? err.message : 'Unknown error'));
        } finally {
            setDeletingFlow(false);
        }
    };

    const handleSaveFlow = async (flowData: Partial<Flow>) => {
        try {
            if (editingFlow) {
                await flowsApi.update(editingFlow.id, flowData);
            } else {
                await flowsApi.create(flowData as Omit<Flow, 'id' | 'lastModified'>);
            }
            setShowFlowModal(false);
            loadFlows();
        } catch (err) {
            throw new Error('Failed to save flow: ' + (err instanceof Error ? err.message : 'Unknown error'));
        }
    };

    // --- Approval Logic ---

    const loadApprovals = async () => {
        try {
            setLoadingApprovals(true);
            const records = await dataAPI.query<ApprovalProcess>({
                objectApiName: SYSTEM_TABLE_NAMES.SYSTEM_APPROVALPROCESS,
                sortField: 'created_date',
                sortDirection: 'DESC'
            });
            setProcesses(records);
            setApprovalError(null);
        } catch (err) {
            setApprovalError('Failed to load approval processes: ' + (err instanceof Error ? err.message : 'Unknown error'));
        } finally {
            setLoadingApprovals(false);
        }
    };

    const handleCreateApproval = () => {
        setEditingProcess(null);
        setShowApprovalModal(true);
    };

    const handleEditApproval = (process: ApprovalProcess) => {
        setEditingProcess(process);
        setShowApprovalModal(true);
    };

    const handleSaveApproval = async (formData: Partial<ApprovalProcess>) => {
        try {
            setApprovalError(null);

            if (editingProcess) {
                await dataAPI.updateRecord(SYSTEM_TABLE_NAMES.SYSTEM_APPROVALPROCESS, editingProcess.id, formData);
            } else {
                await dataAPI.createRecord<ApprovalProcess>(SYSTEM_TABLE_NAMES.SYSTEM_APPROVALPROCESS, formData);
            }

            loadApprovals();
            // Modal closes itself on success callback if no error thrown
        } catch (err) {
            // Rethrow to let modal handle error state
            throw new Error('Failed to save process: ' + (err instanceof Error ? err.message : 'Unknown error'));
        }
    };

    const confirmDeleteApproval = async () => {
        if (!processToDelete) return;
        try {
            await dataAPI.deleteRecord(SYSTEM_TABLE_NAMES.SYSTEM_APPROVALPROCESS, processToDelete.id);
            setDeleteApprovalModalOpen(false);
            setProcessToDelete(null);
            loadApprovals();
        } catch (err) {
            setApprovalError('Failed to delete process: ' + (err instanceof Error ? err.message : 'Unknown error'));
        }
    };

    return (
        <div className="p-6 max-w-7xl mx-auto">
            {/* Header */}
            <div className="flex items-center justify-between mb-6">
                <div className="flex items-center gap-3">
                    <div className="p-2 bg-gradient-to-br from-purple-500 to-indigo-600 rounded-xl shadow-lg">
                        <Zap className="w-6 h-6 text-white" />
                    </div>
                    <div>
                        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Automation Studio</h1>
                        <p className="text-sm text-gray-500 dark:text-gray-400">
                            Manage flows and approval processes
                        </p>
                    </div>
                </div>

                {activeTab === 'flows' ? (
                    <button
                        onClick={() => { setEditingFlow(null); setShowFlowModal(true); }}
                        className="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-purple-500 to-indigo-600 
                        text-white rounded-lg hover:from-purple-600 hover:to-indigo-700 transition-all shadow-md"
                    >
                        <Plus className="w-4 h-4" />
                        New Flow
                    </button>
                ) : (
                    <button
                        onClick={handleCreateApproval}
                        className="flex items-center gap-2 px-4 py-2 bg-gradient-to-r from-green-500 to-emerald-600 
                        text-white rounded-lg hover:from-green-600 hover:to-emerald-700 transition-all shadow-md"
                    >
                        <Plus className="w-4 h-4" />
                        New Approval Process
                    </button>
                )}
            </div>

            {/* Tabs */}
            <div className="flex bg-gray-100 dark:bg-gray-800 p-1 rounded-xl w-fit mb-6">
                <button
                    onClick={() => setActiveTab('flows')}
                    className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all ${activeTab === 'flows'
                        ? 'bg-white dark:bg-gray-700 text-purple-600 dark:text-purple-400 shadow-sm'
                        : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
                        }`}
                >
                    <Zap className="w-4 h-4" />
                    Flows
                </button>
                <button
                    onClick={() => setActiveTab('approvals')}
                    className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-all ${activeTab === 'approvals'
                        ? 'bg-white dark:bg-gray-700 text-green-600 dark:text-green-400 shadow-sm'
                        : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
                        }`}
                >
                    <ShieldCheck className="w-4 h-4" />
                    Approval Processes
                </button>
            </div>

            {/* Error Message */}
            {(flowError || approvalError) && (
                <div className="mb-4 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 
                rounded-lg flex items-center gap-2 text-red-700 dark:text-red-400">
                    <AlertCircle className="w-5 h-5" />
                    {flowError || approvalError}
                </div>
            )}

            {/* Content: Flows */}
            {activeTab === 'flows' && (
                <FlowList
                    flows={flows}
                    loading={loadingFlows}
                    onToggleStatus={handleToggleStatus}
                    onEdit={(flow) => { setEditingFlow(flow); setShowFlowModal(true); }}
                    onDelete={(flow) => { setFlowToDelete(flow); setDeleteFlowModalOpen(true); }}
                    onExecute={(flow) => { setFlowToExecute(flow); setExecuteModalOpen(true); }}
                    onCreate={() => { setEditingFlow(null); setShowFlowModal(true); }}
                />
            )}

            {/* Content: Approvals */}
            {activeTab === 'approvals' && (
                <ApprovalProcessList
                    processes={processes}
                    schemas={schemas}
                    loading={loadingApprovals}
                    onEdit={handleEditApproval}
                    onDelete={(process) => { setProcessToDelete(process); setDeleteApprovalModalOpen(true); }}
                    onCreate={handleCreateApproval}
                />
            )}

            {/* Modals */}

            {showFlowModal && (
                <FlowBuilderModal
                    flow={editingFlow}
                    objects={objects}
                    onSave={handleSaveFlow}
                    onClose={() => setShowFlowModal(false)}
                />
            )}

            <ConfirmationModal
                isOpen={deleteFlowModalOpen}
                onClose={() => { setDeleteFlowModalOpen(false); setFlowToDelete(null); }}
                onConfirm={confirmDeleteFlow}
                title="Delete Flow"
                message={`Are you sure you want to delete the flow "${flowToDelete?.name}"?`}
                confirmLabel="Delete"
                cancelLabel="Cancel"
                variant="danger"
                loading={deletingFlow}
            />

            {flowToExecute && (
                <FlowExecutionModal
                    isOpen={executeModalOpen}
                    onClose={() => { setExecuteModalOpen(false); setFlowToExecute(null); }}
                    flow={flowToExecute}
                />
            )}

            <ApprovalProcessModal
                isOpen={showApprovalModal}
                onClose={() => setShowApprovalModal(false)}
                onSave={handleSaveApproval}
                editingProcess={editingProcess}
                schemas={schemas}
            />

            <ConfirmationModal
                isOpen={deleteApprovalModalOpen}
                onClose={() => setDeleteApprovalModalOpen(false)}
                onConfirm={confirmDeleteApproval}
                title="Delete Approval Process"
                message={`Are you sure you want to delete "${processToDelete?.name}"?`}
                confirmLabel="Delete"
                variant="danger"
            />
        </div>
    );
};

export default FlowsPage;
