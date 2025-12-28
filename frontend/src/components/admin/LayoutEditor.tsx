import React from 'react';
import { useParams, Link } from 'react-router-dom';
import { useObjectMetadata, useLayout } from '../../core/hooks/useMetadata';
import { ArrowLeft, Save, Layout as LayoutIcon, Trash2 } from 'lucide-react';
import { ConfirmationModal } from '../modals/ConfirmationModal';

// Sub-components
import { LayoutPalette } from './layout/LayoutPalette';
import { LayoutCanvas } from './layout/LayoutCanvas';
import { ListViewEditor } from './layout/ListViewEditor';
import { CompactViewEditor } from './layout/CompactViewEditor';
import { RelatedListEditor } from './layout/RelatedListEditor';
import { useLayoutState } from './layout/useLayoutState';

export const LayoutEditor: React.FC = () => {
    const { objectApiName } = useParams<{ objectApiName: string }>();
    const { metadata } = useObjectMetadata(objectApiName || '');
    const { layout: existingLayout } = useLayout(objectApiName || '');

    const {
        layout,
        setLayout,
        listFields,
        isSaving,
        draggedField,
        dragOverSection,
        setDragOverSection,

        // Modals
        showRemoveSectionModal,
        setShowRemoveSectionModal,
        sectionToRemove,
        showRemoveRelatedListModal,
        setShowRemoveRelatedListModal,
        relatedListToRemove,
        showSaveModal,
        setShowSaveModal,
        saveStatus,
        saveMessage,

        // Tab state
        activeTab,
        setActiveTab,

        // Handlers
        handleRemoveRelatedList,
        confirmRemoveRelatedList,
        handleDragStart,
        handleDrop,
        handleRemoveField,
        handleMoveField,
        handleAddSection,
        handleRemoveSection,
        confirmRemoveSection,
        handleUpdateSection,
        handleSave,
        toggleListField,
        moveListField,
        openVisibilityEditor,
        onEditSection,
    } = useLayoutState({
        objectApiName: objectApiName || '',
        metadata,
        existingLayout
    });

    // Helpers for Drag and Drop Palette
    const usedFieldNames = new Set(layout?.sections.flatMap(s => s.fields) || []);
    const availableFields = metadata?.fields.filter(f => !usedFieldNames.has(f.api_name)) || [];

    if (!metadata || !layout) return <div className="p-8 text-center">Loading editor...</div>;

    return (
        <div className="h-[calc(100vh-4rem)] flex flex-col bg-slate-50">
            {/* Header */}
            <div className="bg-white border-b border-slate-200 px-6 py-4 flex items-center justify-between shadow-sm z-10">
                <div className="flex items-center gap-4">
                    <Link to={`/setup/objects/${objectApiName}`} className="text-slate-500 hover:text-slate-700">
                        <ArrowLeft size={20} />
                    </Link>
                    <div>
                        <h1 className="text-xl font-bold text-slate-800 flex items-center gap-2">
                            <LayoutIcon className="text-blue-600" size={24} />
                            Layout Editor: {metadata.label}
                        </h1>
                        <p className="text-sm text-slate-500">Drag fields to move them. Use arrows to reorder.</p>
                    </div>
                </div>
                <div className="flex gap-3">
                    <button
                        onClick={handleSave}
                        disabled={isSaving}
                        className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium disabled:opacity-50 shadow-sm transition-colors"
                    >
                        <Save size={18} />
                        {isSaving ? 'Saving...' : 'Save Layout'}
                    </button>
                </div>
            </div>

            {/* Tabs Header */}
            <div className="bg-white border-b border-slate-200 px-6 flex items-center gap-6">
                {(['details', 'list', 'compact', 'related'] as const).map(tab => (
                    <button
                        key={tab}
                        onClick={() => setActiveTab(tab)}
                        className={`py-3 text-sm font-medium border-b-2 transition-colors ${activeTab === tab ? 'border-blue-600 text-blue-600' : 'border-transparent text-slate-500 hover:text-slate-700'
                            } capitalize`}
                    >
                        {tab === 'details' ? 'Detail Layout' : tab === 'list' ? 'List View' : tab === 'compact' ? 'Compact Layout' : 'Related Lists'}
                    </button>
                ))}
            </div>

            {/* Main Content */}
            <div className="flex-1 flex overflow-hidden">
                {activeTab === 'details' && (
                    <>
                        <LayoutPalette
                            availableFields={availableFields}
                            metadata={metadata}
                            usedFieldNames={usedFieldNames}
                            onDragStart={handleDragStart}
                        />
                        <LayoutCanvas
                            layout={layout}
                            metadata={metadata}
                            dragOverSection={dragOverSection}
                            onUpdateSection={handleUpdateSection}
                            onRemoveSection={handleRemoveSection}
                            onAddSection={handleAddSection}
                            onEditSection={onEditSection}
                            onOpenVisibilityEditor={openVisibilityEditor}
                            onDrop={handleDrop}
                            onDragOver={setDragOverSection}
                            onDragLeave={(e: React.DragEvent) => {
                                if (e.currentTarget.contains(e.relatedTarget as Node)) return;
                                setDragOverSection(null);
                            }}
                            onDragStart={handleDragStart}
                            onMoveField={handleMoveField}
                            onRemoveField={handleRemoveField}
                        />
                    </>
                )}

                {activeTab === 'list' && (
                    <ListViewEditor
                        metadata={metadata}
                        listFields={listFields}
                        onToggleField={toggleListField}
                        onMoveField={moveListField}
                    />
                )}

                {activeTab === 'compact' && (
                    <CompactViewEditor
                        metadata={metadata}
                        layout={layout}
                        onToggleField={toggleListField}
                        onMoveField={moveListField}
                    />
                )}

                {activeTab === 'related' && (
                    <div className="flex-1 overflow-y-auto p-8">
                        <RelatedListEditor
                            layout={layout}
                            metadata={metadata}
                            onUpdateLayout={setLayout}
                            onRemoveRelatedList={handleRemoveRelatedList}
                        />
                    </div>
                )}
            </div>

            {/* Modals */}
            <ConfirmationModal
                isOpen={showRemoveSectionModal}
                onClose={() => setShowRemoveSectionModal(false)}
                onConfirm={confirmRemoveSection}
                title="Remove Section"
                message="Are you sure you want to remove this section? All fields in it will be moved back to the palette."
                confirmLabel="Remove"
                icon={<Trash2 className="text-red-600" />}
            />

            <ConfirmationModal
                isOpen={showRemoveRelatedListModal}
                onClose={() => setShowRemoveRelatedListModal(false)}
                onConfirm={confirmRemoveRelatedList}
                title="Remove Related List"
                message="Are you sure you want to remove this related list?"
                confirmLabel="Remove"
                icon={<Trash2 className="text-red-600" />}
            />

            <ConfirmationModal
                isOpen={showSaveModal}
                onClose={() => setShowSaveModal(false)}
                onConfirm={() => setShowSaveModal(false)}
                title={saveStatus === 'success' ? 'Saved' : 'Error'}
                message={saveMessage}
                confirmLabel="OK"
                icon={saveStatus === 'success' ? <Save className="text-green-600" /> : <Trash2 className="text-red-600" />}
            />
        </div>
    );
};
