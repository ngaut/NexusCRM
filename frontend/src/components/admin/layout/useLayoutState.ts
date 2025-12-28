import { useState, useEffect } from 'react';
import type { PageLayout, FieldMetadata, PageSection, ObjectMetadata } from '../../../types';
import { useErrorToast } from '../../ui/Toast';
import { metadataAPI } from '../../../infrastructure/api/metadata';

interface DragData {
    field: FieldMetadata;
    sourceSectionId?: string;
}

interface UseLayoutStateProps {
    objectApiName: string;
    metadata: ObjectMetadata | null;
    existingLayout: PageLayout | null;
}

export function useLayoutState({ objectApiName, metadata, existingLayout }: UseLayoutStateProps) {
    const showError = useErrorToast();
    const [layout, setLayout] = useState<PageLayout | null>(null);
    const [listFields, setListFields] = useState<string[]>([]);
    const [draggedField, setDraggedField] = useState<FieldMetadata | null>(null);
    const [dragOverSection, setDragOverSection] = useState<string | null>(null);
    const [isSaving, setIsSaving] = useState(false);

    // Modal states
    const [sectionToRemove, setSectionToRemove] = useState<string | null>(null);
    const [showRemoveSectionModal, setShowRemoveSectionModal] = useState(false);
    const [relatedListToRemove, setRelatedListToRemove] = useState<string | null>(null);
    const [showRemoveRelatedListModal, setShowRemoveRelatedListModal] = useState(false);
    const [showSaveModal, setShowSaveModal] = useState(false);
    const [saveStatus, setSaveStatus] = useState<'success' | 'error' | null>(null);
    const [saveMessage, setSaveMessage] = useState('');

    // Active Tab
    const [activeTab, setActiveTab] = useState<'details' | 'list' | 'compact' | 'related'>('details');

    // Visibility Editor state
    const [editingSectionId, setEditingSectionId] = useState<string | null>(null);
    const [visibilityEditorOpen, setVisibilityEditorOpen] = useState(false);
    const [selectedSectionForVisibility, setSelectedSectionForVisibility] = useState<PageSection | null>(null);
    const [tempVisibilityCondition, setTempVisibilityCondition] = useState('');

    // Initialize layout
    useEffect(() => {
        if (existingLayout) {
            setLayout(existingLayout);
        } else if (metadata) {
            setLayout({
                id: crypto.randomUUID(),
                object_api_name: metadata.api_name,
                layout_name: 'Default Layout',
                type: 'Detail',
                compact_layout: [],
                sections: [
                    {
                        id: crypto.randomUUID(),
                        label: 'Information',
                        columns: 2,
                        fields: []
                    }
                ],
                related_lists: [],
                header_actions: [],
                quick_actions: []
            });
        }

        if (metadata?.list_fields) {
            setListFields(metadata.list_fields);
        } else if (metadata) {
            setListFields(metadata.fields.filter(f => !f.is_system || f.api_name === 'name').slice(0, 5).map(f => f.api_name));
        }
    }, [existingLayout, metadata]);

    // Actions
    const handleRemoveRelatedList = (id: string) => {
        setRelatedListToRemove(id);
        setShowRemoveRelatedListModal(true);
    };

    const confirmRemoveRelatedList = () => {
        if (!relatedListToRemove) return;
        setLayout(prev => {
            if (!prev) return null;
            return {
                ...prev,
                related_lists: (prev.related_lists || []).filter(r => r.id !== relatedListToRemove)
            };
        });
        setShowRemoveRelatedListModal(false);
        setRelatedListToRemove(null);
    };

    const handleDragStart = (e: React.DragEvent, field: FieldMetadata, sourceSectionId?: string) => {
        setDraggedField(field);
        e.dataTransfer.effectAllowed = 'copyMove';
        e.dataTransfer.setData('application/json', JSON.stringify({ field, sourceSectionId }));
    };

    const handleDrop = (e: React.DragEvent, targetSectionId: string) => {
        e.preventDefault();
        setDragOverSection(null);

        try {
            const rawData = e.dataTransfer.getData('application/json');
            if (!rawData) return;

            const data = JSON.parse(rawData) as DragData;
            const droppedField = data.field;
            const sourceSectionId = data.sourceSectionId;

            if (!droppedField) return;

            setLayout(prev => {
                if (!prev) return null;
                if (sourceSectionId === targetSectionId) return prev;

                let newSections = [...prev.sections];

                if (sourceSectionId) {
                    newSections = newSections.map(section => {
                        if (section.id === sourceSectionId) {
                            return {
                                ...section,
                                fields: section.fields.filter(f => f !== droppedField.api_name)
                            };
                        }
                        return section;
                    });
                }

                newSections = newSections.map(section => {
                    if (section.id === targetSectionId) {
                        if (section.fields.includes(droppedField.api_name)) return section;
                        return {
                            ...section,
                            fields: [...section.fields, droppedField.api_name]
                        };
                    }
                    return section;
                });

                return { ...prev, sections: newSections };
            });
        } catch (e) {
            console.warn('Failed to parse drag data:', e);
        }

        setDraggedField(null);
    };

    const handleRemoveField = (sectionId: string, fieldApiName: string) => {
        setLayout(prev => {
            if (!prev) return null;
            return {
                ...prev,
                sections: prev.sections.map(section => {
                    if (section.id === sectionId) {
                        return {
                            ...section,
                            fields: section.fields.filter(f => f !== fieldApiName)
                        };
                    }
                    return section;
                })
            };
        });
    };

    const handleMoveField = (sectionId: string, fieldApiName: string, direction: 'up' | 'down') => {
        setLayout(prev => {
            if (!prev) return null;
            return {
                ...prev,
                sections: prev.sections.map(section => {
                    if (section.id === sectionId) {
                        const fields = [...section.fields];
                        const index = fields.indexOf(fieldApiName);
                        if (index === -1) return section;

                        if (direction === 'up' && index > 0) {
                            [fields[index], fields[index - 1]] = [fields[index - 1], fields[index]];
                        } else if (direction === 'down' && index < fields.length - 1) {
                            [fields[index], fields[index + 1]] = [fields[index + 1], fields[index]];
                        }

                        return { ...section, fields };
                    }
                    return section;
                })
            };
        });
    };

    const handleAddSection = () => {
        setLayout(prev => {
            if (!prev) return null;
            return {
                ...prev,
                sections: [
                    ...prev.sections,
                    {
                        id: crypto.randomUUID(),
                        label: 'New Section',
                        columns: 2,
                        fields: []
                    }
                ]
            };
        });
    };

    const handleRemoveSection = (sectionId: string) => {
        setSectionToRemove(sectionId);
        setShowRemoveSectionModal(true);
    };

    const confirmRemoveSection = () => {
        if (!sectionToRemove) return;
        setLayout(prev => {
            if (!prev) return null;
            return {
                ...prev,
                sections: prev.sections.filter(s => s.id !== sectionToRemove)
            };
        });
        setShowRemoveSectionModal(false);
        setSectionToRemove(null);
    };

    const handleUpdateSection = (sectionId: string, updates: Partial<PageSection>) => {
        setLayout(prev => {
            if (!prev) return null;
            return {
                ...prev,
                sections: prev.sections.map(s => s.id === sectionId ? { ...s, ...updates } : s)
            };
        });
    };

    const handleSave = async () => {
        if (!layout || !metadata) return;
        setIsSaving(true);
        try {
            await metadataAPI.saveLayout(layout);
            await metadataAPI.updateSchema(metadata.api_name, {
                list_fields: listFields
            });

            setSaveStatus('success');
            setSaveMessage('Layout and List View settings saved successfully!');
            setShowSaveModal(true);
        } catch {
            setSaveStatus('error');
            setSaveMessage('Failed to save layout. Please try again.');
            setShowSaveModal(true);
        } finally {
            setIsSaving(false);
        }
    };

    const toggleListField = (fieldApiName: string, targetList: 'list' | 'compact') => {
        if (targetList === 'list') {
            setListFields(prev => {
                if (prev.includes(fieldApiName)) return prev.filter(f => f !== fieldApiName);
                return [...prev, fieldApiName];
            });
        } else {
            setLayout(prev => {
                if (!prev) return null;
                const current = prev.compact_layout || [];
                if (current.includes(fieldApiName)) {
                    return { ...prev, compact_layout: current.filter(f => f !== fieldApiName) };
                }
                if (current.length >= 6) {
                    showError('Maximum 6 fields allowed in Compact Layout');
                    return prev;
                }
                return { ...prev, compact_layout: [...current, fieldApiName] };
            });
        }
    };

    const moveListField = (fieldApiName: string, direction: 'up' | 'down', targetList: 'list' | 'compact') => {
        const update = (current: string[]) => {
            const index = current.indexOf(fieldApiName);
            if (index === -1) return current;
            const newArr = [...current];
            if (direction === 'up' && index > 0) {
                [newArr[index], newArr[index - 1]] = [newArr[index - 1], newArr[index]];
            } else if (direction === 'down' && index < newArr.length - 1) {
                [newArr[index], newArr[index + 1]] = [newArr[index + 1], newArr[index]];
            }
            return newArr;
        };

        if (targetList === 'list') {
            setListFields(prev => update(prev));
        } else {
            setLayout(prev => {
                if (!prev) return null;
                return { ...prev, compact_layout: update(prev.compact_layout || []) };
            });
        }
    };

    const openVisibilityEditor = (e: React.MouseEvent, section: PageSection) => {
        e.stopPropagation();
        setSelectedSectionForVisibility(section);
        setTempVisibilityCondition(section.visibility_condition || '');
        setVisibilityEditorOpen(true);
    };

    const onEditSection = (sectionId: string) => {
        setEditingSectionId(sectionId);
    };

    return {
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

        // Visibility Editor
        editingSectionId,
        visibilityEditorOpen,
        setVisibilityEditorOpen,
        selectedSectionForVisibility,
        setSelectedSectionForVisibility,
        tempVisibilityCondition,
        setTempVisibilityCondition,

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
    };
}
