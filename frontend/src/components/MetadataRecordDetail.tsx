import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Lock, Share2, Send } from 'lucide-react';
import { ObjectMetadata, SObject, PageLayout, FieldMetadata } from '../types';
import { dataAPI } from '../infrastructure/api/data';
import { Button } from './ui/Button';
import { EmptyState, ErrorEmptyState, AccessDeniedEmptyState } from './ui/EmptyState';
import { MetadataAwareSkeleton } from './ui/LoadingSkeleton';
import { useToast, useSuccessToast, useErrorToast } from './ui/Toast';
import { formatApiError, getOperationErrorMessage, AppError } from '../core/utils/errorHandling';
import { getHighlightFields, getPathField } from '../core/utils/recordUtils';
import { usePermissions } from '../contexts/PermissionContext';
import { InlineEditField } from './InlineEditField';
import { ActivityFeed } from './ActivityFeed';
import { RelatedList } from './RelatedList';
import { HighlightsPanel } from './HighlightsPanel';
import { Path } from './Path';
import { ChangePasswordModal } from './modals/ChangePasswordModal';
import { SubmitApprovalModal } from './modals/SubmitApprovalModal';
import { StudioFieldEditor } from './studio/StudioFieldEditor';
import { TableObject, TableField, TableUser } from '../constants';
import { COMMON_FIELDS, APPROVAL_STATUS } from '../core/constants';
import { useActions } from '../core/hooks/useMetadata';
import { ActionRenderer } from './ActionRenderer';
import { LayoutRenderer } from './runtime/LayoutRenderer';
import { ShareRecordModal } from './modals/ShareRecordModal';
import { ApprovalHistory, useApprovalStatus } from './ApprovalHistory';
import { ApprovalBanner } from './ApprovalBanner';
import { approvalsAPI } from '../infrastructure/api/approvals';

interface MetadataRecordDetailProps {
    objectMetadata: ObjectMetadata;
    recordId: string;
    layout?: PageLayout | null;
    onBack?: () => void;
    extraActions?: React.ReactNode;
}

export function MetadataRecordDetail({
    objectMetadata,
    recordId,
    layout,
    onBack,
    extraActions,
}: MetadataRecordDetailProps) {
    const navigate = useNavigate();
    const showSuccess = useSuccessToast();
    const showError = useErrorToast();
    const { hasObjectPermission, hasFieldPermission } = usePermissions();
    const { actions, refresh: refreshActions } = useActions(objectMetadata.api_name);
    const { status: approvalStatus, pendingItem, loading: approvalStatusLoading, refresh: refreshApprovalStatus } = useApprovalStatus(objectMetadata.api_name, recordId);

    const [record, setRecord] = useState<SObject | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<AppError | null>(null);
    const [refreshKey, setRefreshKey] = useState(0);
    const [changePasswordModalOpen, setChangePasswordModalOpen] = useState(false);
    const [createFieldWizardOpen, setCreateFieldWizardOpen] = useState(false);
    const [fieldsRefreshKey, setFieldsRefreshKey] = useState(0);
    const [shareModalOpen, setShareModalOpen] = useState(false);
    const [approvalModalOpen, setApprovalModalOpen] = useState(false);
    const [hasApprovalProcess, setHasApprovalProcess] = useState(false);

    // Load record data
    const loadRecord = async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await dataAPI.getRecord(objectMetadata.api_name, recordId);
            setRecord(data);
        } catch (err) {
            const apiError = formatApiError(err);
            setError(apiError);
            showError(getOperationErrorMessage('fetch', objectMetadata.label, apiError));
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadRecord();
    }, [objectMetadata.api_name, recordId, refreshKey]);

    // Check if approval process exists for this object
    useEffect(() => {
        approvalsAPI.hasProcessForObject(objectMetadata.api_name)
            .then(setHasApprovalProcess)
            .catch(() => setHasApprovalProcess(false));
    }, [objectMetadata.api_name]);

    const handleUpdate = async (fieldName: string, newValue: unknown) => {
        // Optimistic update
        if (record) {
            setRecord({ ...record, [fieldName]: newValue });
        }
    };

    const handleNavigate = (obj: string, id: string) => {
        navigate(`/object/${obj}/${id}`);
    };

    if (loading && !record) {
        return (
            <div className="space-y-6">
                <div className="h-12 w-full bg-gray-100 rounded-lg animate-pulse mb-6" />
                <MetadataAwareSkeleton
                    fieldCount={objectMetadata.fields.length}
                    layout="detail"
                />
            </div>
        );
    }

    if (error || !record) {
        // Show access denied if 403 error
        if (error?.code === 'FORBIDDEN') {
            return <AccessDeniedEmptyState onGoBack={() => navigate(-1)} />;
        }
        return <ErrorEmptyState onRetry={loadRecord} />;
    }

    // Determine Highlighting fields (first 4 non-system fields usually)
    const highlightFields = getHighlightFields(objectMetadata);

    // Identify "Status" or "Stage" field for Path component
    const pathField = getPathField(objectMetadata);

    return (
        <div className="space-y-6 animate-fade-in">
            {/* Header */}
            <div className="bg-white border-b border-gray-200 -mx-6 -mt-6 px-6 py-4 sticky top-0 z-10 shadow-sm">
                <div className="flex justify-between items-start">
                    <div className="flex items-center gap-4">
                        <Button variant="ghost" size="sm" onClick={onBack || (() => navigate(-1))} icon={<ArrowLeft className="w-4 h-4" />}>
                            Back
                        </Button>
                        <div>
                            <div className="text-sm text-gray-500">{objectMetadata.label}</div>
                            <h1 className="text-2xl font-bold text-gray-900 flex items-center gap-2">
                                {(record[COMMON_FIELDS.NAME] as string | number) || (record[COMMON_FIELDS.SUBJECT] as string | number) || 'Untitled Record'}
                            </h1>
                        </div>
                    </div>
                    <div className="flex items-center gap-2">
                        {extraActions}
                        {actions.map(action => (
                            <ActionRenderer
                                key={action.id}
                                action={action}
                                record={record}
                                onActionComplete={() => {
                                    loadRecord(); // Refresh record
                                    refreshActions(); // Refresh actions if needed
                                }}
                            />
                        ))}
                        {objectMetadata.api_name === TableUser && hasObjectPermission(objectMetadata.api_name, 'edit') && (
                            <Button variant="secondary" onClick={() => setChangePasswordModalOpen(true)} icon={<Lock className="w-4 h-4" />}>
                                Change Password
                            </Button>
                        )}
                        {/* Submit for Approval - only shown when approval process exists */}
                        {/* Submit for Approval - only shown when approval process exists */}
                        {hasApprovalProcess && hasObjectPermission(objectMetadata.api_name, 'read') && (
                            <Button
                                variant="ghost"
                                onClick={() => setApprovalModalOpen(true)}
                                disabled={!!pendingItem}
                                icon={pendingItem ? <Lock className="w-4 h-4" /> : <Send className="w-4 h-4" />}
                            >
                                {pendingItem ? 'Pending Approval' : 'Submit for Approval'}
                            </Button>
                        )}
                        {hasObjectPermission(objectMetadata.api_name, 'read') && (
                            <Button variant="ghost" onClick={() => setShareModalOpen(true)} icon={<Share2 className="w-4 h-4" />}>
                                Share
                            </Button>
                        )}
                    </div>
                </div>

                {/* Path / Progress Bar */}
                {pathField && (
                    <div className="mt-6">
                        <Path
                            objectApiName={objectMetadata.api_name}
                            record={record}
                            pathField={pathField.api_name}
                            fields={objectMetadata.fields}
                            onUpdate={() => handleUpdate(pathField.api_name, record[pathField.api_name])}
                        />
                    </div>
                )}
            </div>

            {/* Approval Banner - shows when pending */}
            <ApprovalBanner
                objectApiName={objectMetadata.api_name}
                recordId={recordId}
                recordName={String(record?.[COMMON_FIELDS.NAME] || record?.[COMMON_FIELDS.SUBJECT] || 'Record')}
                pendingItem={pendingItem}
                onActionComplete={() => {
                    loadRecord();
                    refreshApprovalStatus();
                }}
            />

            {/* Highlights Panel */}
            <HighlightsPanel
                record={record}
                fields={highlightFields}
                layout={layout || null}
            />

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                {/* Main Details Column (2/3) */}
                <div className="lg:col-span-2 space-y-8">
                    {/* Details Section */}

                    <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
                        {layout ? (
                            <LayoutRenderer
                                layout={layout}
                                record={record}
                                objectMetadata={objectMetadata}
                                onUpdate={handleUpdate}
                                onNavigate={handleNavigate}
                            />
                        ) : (
                            // Fallback if no layout (shouldn't happen given backend defaults, but safe to keep)
                            <div className="p-8 text-center text-gray-500">
                                No page layout assigned.
                            </div>
                        )}
                    </div>

                    {/* Related Lists */}
                    {layout?.related_lists?.map(rl => (
                        <div key={rl.id} className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
                            <div className="px-6 py-4 border-b border-gray-100 bg-gray-50">
                                <h2 className="font-semibold text-gray-900">{rl.label}</h2>
                            </div>
                            <div className="p-6">
                                <RelatedList
                                    config={rl}
                                    parentRecordId={recordId}
                                    parentObjectApiName={objectMetadata.api_name}
                                />
                            </div>
                        </div>
                    ))}

                    {/* Special Schema Builder Related Lists for System Objects */}
                    {objectMetadata.api_name === TableObject && record && (
                        <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden mt-8">
                            <div className="px-6 py-4 border-b border-gray-100 bg-gray-50 flex justify-between items-center">
                                <h2 className="font-semibold text-gray-900">Fields</h2>
                                <Button
                                    variant="secondary"
                                    size="sm"
                                    onClick={() => setCreateFieldWizardOpen(true)}
                                >
                                    + Add Field
                                </Button>
                            </div>
                            <div className="p-6">
                                <RelatedList
                                    config={{
                                        id: 'fields',
                                        label: 'Fields',
                                        object_api_name: TableField,
                                        lookup_field: COMMON_FIELDS.OBJECT_ID,
                                        fields: [COMMON_FIELDS.API_NAME, COMMON_FIELDS.LABEL, COMMON_FIELDS.TYPE, COMMON_FIELDS.IS_REQUIRED]
                                    }}
                                    parentRecordId={recordId}
                                    parentObjectApiName={objectMetadata.api_name}
                                    refreshKey={fieldsRefreshKey}
                                />
                            </div>
                        </div>
                    )}
                </div>

                {/* Sidebar Column (1/3) */}
                <div className="lg:col-span-1 space-y-6">
                    {/* Approval History */}
                    <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
                        <div className="px-6 py-4 border-b border-gray-100 bg-gray-50 flex items-center justify-between">
                            <h2 className="font-semibold text-gray-900">Approvals</h2>
                            {approvalStatus && (
                                <span className={`px-2 py-1 text-xs font-medium rounded-full ${approvalStatus === APPROVAL_STATUS.PENDING ? 'bg-amber-100 text-amber-700' :
                                    approvalStatus === APPROVAL_STATUS.APPROVED ? 'bg-green-100 text-green-700' :
                                        'bg-red-100 text-red-700'
                                    }`}>
                                    {approvalStatus}
                                </span>
                            )}
                        </div>
                        <div className="p-6">
                            <ApprovalHistory
                                objectApiName={objectMetadata.api_name}
                                recordId={recordId}
                            />
                        </div>
                    </div>

                    {/* Activity Timeline */}
                    <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden sticky top-24">
                        <div className="px-6 py-4 border-b border-gray-100 bg-gray-50">
                            <h2 className="font-semibold text-gray-900">Activity</h2>
                        </div>
                        <div className="p-6">
                            <ActivityFeed
                                recordId={recordId}
                                objectApiName={objectMetadata.api_name}
                            />
                        </div>
                    </div>
                </div>
            </div>
            <ChangePasswordModal
                isOpen={changePasswordModalOpen}
                onClose={() => setChangePasswordModalOpen(false)}
                recordId={recordId}
                objectApiName={objectMetadata.api_name}
            />

            {/* Schema Builder: Field Creation Wizard */}
            {objectMetadata.api_name === TableObject && record && createFieldWizardOpen && (
                <StudioFieldEditor
                    objectApiName={String(record.api_name)}
                    field={null}
                    onSave={() => {
                        setFieldsRefreshKey(k => k + 1);
                        setCreateFieldWizardOpen(false);
                    }}
                    onClose={() => setCreateFieldWizardOpen(false)}
                />
            )}

            {/* Share Record Modal */}
            <ShareRecordModal
                isOpen={shareModalOpen}
                onClose={() => setShareModalOpen(false)}
                objectApiName={objectMetadata.api_name}
                recordId={recordId}
                recordName={String(record?.[COMMON_FIELDS.NAME] || record?.[COMMON_FIELDS.SUBJECT] || 'Record')}
            />

            {/* Submit for Approval Modal */}
            <SubmitApprovalModal
                isOpen={approvalModalOpen}
                onClose={() => setApprovalModalOpen(false)}
                objectApiName={objectMetadata.api_name}
                recordId={recordId}
                recordName={String(record?.[COMMON_FIELDS.NAME] || record?.[COMMON_FIELDS.SUBJECT] || undefined)}
                onSuccess={loadRecord}
            />
        </div>
    );
}
