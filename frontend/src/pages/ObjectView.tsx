import React, { useState, useEffect } from 'react';
import { useParams, useSearchParams, useNavigate } from 'react-router-dom';
import { useObjectMetadata, useLayout } from '../core/hooks/useMetadata';
import { MetadataRecordList } from '../components/MetadataRecordList';
import { MetadataRecordDetail } from '../components/MetadataRecordDetail';
import { MetadataRecordForm } from '../components/MetadataRecordForm';
import { MetadataAwareSkeleton } from '../components/ui/LoadingSkeleton';
import { EmptyState } from '../components/ui/EmptyState';
import { TableObject } from '../constants';
import { Button } from '../components/ui/Button';
import { ArrowLeft } from 'lucide-react';
import { dataAPI } from '../infrastructure/api/data';
import { SYSTEM_TABLE_NAMES } from '../generated-schema';
import { formatApiError } from '../core/utils/errorHandling';

export const ObjectView: React.FC = () => {
    const { objectApiName, recordId } = useParams<{ objectApiName: string; recordId: string }>();
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();

    // Local state for editing
    const [isEditing, setIsEditing] = useState(false);

    // Sync isEditing state with URL query param
    useEffect(() => {
        const editParam = searchParams.get('edit');
        if (editParam === 'true') {
            setIsEditing(true);
        } else {
            setIsEditing(false);
        }
    }, [searchParams]);

    // Fetch Metadata
    const { metadata, loading: metaLoading, error: metaError } = useObjectMetadata(objectApiName || '');

    // Fetch Record Data for Edit / Detail
    const [recordData, setRecordData] = useState<Record<string, unknown> | null>(null);
    const [recordLoading, setRecordLoading] = useState(false);
    const [recordError, setRecordError] = useState<string | null>(null);

    useEffect(() => {
        // Only fetch record data here if we are editing, as the form needs initial data.
        // For detail view, the MetadataRecordDetail component handles its own fetching.
        if (recordId && recordId !== 'new' && isEditing) {
            setRecordLoading(true);
            setRecordData(null);
            dataAPI.getRecord(objectApiName || '', recordId)
                .then(data => setRecordData(data))
                .catch(err => setRecordError(formatApiError(err).message))
                .finally(() => setRecordLoading(false));
        } else if (recordId === 'new') {
            setRecordData(null);
            setRecordLoading(false);
        }
    }, [objectApiName, recordId, isEditing]);

    // Fetch Layout
    const isCreate = recordId === 'new';
    const mode = isCreate ? 'Create' : (isEditing ? 'Edit' : 'Detail');
    const { layout } = useLayout(objectApiName || '', mode);

    if (metaLoading) return (
        <div className="max-w-7xl mx-auto p-6">
            <MetadataAwareSkeleton fieldCount={5} layout="list" />
        </div>
    );
    if (metaError || !metadata) return (
        <EmptyState
            variant="error"
            title="Error Loading Object"
            description={metaError?.message || 'Object not found'}
            action={{ label: 'Retry', onClick: () => window.location.reload() }}
        />
    );

    // Redirect system objects to their dedicated Setup pages
    const systemObjectRedirects: Record<string, string> = {
        [SYSTEM_TABLE_NAMES.SYSTEM_FLOW]: '/setup/flows',
        [SYSTEM_TABLE_NAMES.SYSTEM_SHARINGRULE]: '/setup/sharing-rules',
        [SYSTEM_TABLE_NAMES.SYSTEM_GROUP]: '/setup/groups',
    };

    if (isCreate && objectApiName && systemObjectRedirects[objectApiName]) {
        // Redirect to dedicated setup page 
        navigate(systemObjectRedirects[objectApiName], { replace: true });
        return <div className="p-8 text-center text-slate-500">Redirecting to setup page...</div>;
    }

    // Only block for record loading if we are in Edit/Create mode (where we need data to render form)
    if (recordLoading && (isCreate || isEditing)) return <div className="p-8 text-center text-slate-500">Loading record...</div>;
    if (recordError) return <div className="p-8 text-center text-red-500">Error: {recordError}</div>;

    // View Mode Determination
    const isDetail = !!recordId && !isCreate;
    const isList = !recordId;

    // Handlers
    const handleCreateSuccess = (record: Record<string, unknown>) => {
        navigate(`/object/${objectApiName}/${record.id as string}`);
    };

    const handleUpdateSuccess = () => {
        setIsEditing(false);
        // Clear the edit query param from URL
        navigate(`/object/${objectApiName}/${recordId}`, { replace: true });
    };

    const handleCancel = () => {
        if (isCreate) {
            navigate(`/object/${objectApiName}`);
        } else {
            setIsEditing(false);
            // Clear the edit query param from URL
            navigate(`/object/${objectApiName}/${recordId}`, { replace: true });
        }
    };

    return (
        <div className="max-w-7xl mx-auto p-6">
            {/* List View */}
            {isList && (
                <MetadataRecordList
                    objectMetadata={metadata}
                    onCreateNew={() => navigate(`/object/${objectApiName}/new`)}
                />
            )}

            {/* Create / Edit Form */}
            {(isCreate || isEditing) && (
                <div className="bg-white rounded-xl shadow-lg border border-slate-200 overflow-hidden">
                    <div className="px-6 py-4 border-b border-slate-100 flex items-center gap-4">
                        <Button variant="ghost" size="sm" onClick={handleCancel} icon={<ArrowLeft className="w-4 h-4" />}>
                            Back
                        </Button>
                        <h1 className="text-xl font-bold text-slate-900">
                            {isCreate ? `New ${metadata.label}` : `Edit ${metadata.label}`}
                        </h1>
                    </div>
                    <div className="p-6">
                        <MetadataRecordForm
                            objectMetadata={metadata}
                            recordId={isEditing && recordId ? recordId : undefined}
                            initialData={isCreate ? Object.fromEntries(searchParams.entries()) : recordData}
                            onSuccess={isCreate ? handleCreateSuccess : handleUpdateSuccess}
                            onCancel={handleCancel}
                        />
                    </div>
                </div>
            )}

            {/* Detail View */}
            {isDetail && !isEditing && (
                <MetadataRecordDetail
                    objectMetadata={metadata}
                    recordId={recordId || ''}
                    layout={layout}
                    onBack={() => navigate(`/object/${objectApiName}`)}
                />
            )}
        </div>
    );
};
