import React, { useState } from 'react';
import { PageLayout, SObject, ObjectMetadata } from '../../types';
import { InlineEditField } from '../InlineEditField';
import { usePermissions } from '../../contexts/PermissionContext';
import { ChevronDown, ChevronRight } from 'lucide-react';
import { evaluateExpression } from '../../core/utils/expressionEvaluator';

interface LayoutRendererProps {
    layout: PageLayout;
    record: SObject;
    objectMetadata: ObjectMetadata;
    onUpdate: (field: string, value: unknown) => Promise<void>;
    isReadOnly?: boolean;
    onNavigate?: (obj: string, id: string) => void;
}

export const LayoutRenderer: React.FC<LayoutRendererProps> = ({
    layout,
    record,
    objectMetadata,
    onUpdate,
    isReadOnly = false,
    onNavigate
}) => {
    const { hasFieldPermission } = usePermissions();

    // Section collapse state
    const [collapsedSections, setCollapsedSections] = useState<Record<string, boolean>>({});

    const toggleSection = (sectionId: string) => {
        setCollapsedSections(prev => ({
            ...prev,
            [sectionId]: !prev[sectionId]
        }));
    };

    return (
        <div className="space-y-8">
            {layout.sections.map(section => {
                const isCollapsed = collapsedSections[section.id];

                // Visibility Check
                if (section.visibility_condition && !evaluateExpression(section.visibility_condition, record)) {
                    return null;
                }

                return (
                    <div key={section.id} className="bg-white rounded-lg overflow-hidden">
                        {/* Section Header */}
                        <div
                            className="bg-gray-50 px-6 py-3 border-b border-gray-100 flex items-center gap-2 cursor-pointer hover:bg-gray-100 transition-colors"
                            onClick={() => toggleSection(section.id)}
                        >
                            <button className="text-gray-400">
                                {isCollapsed ? <ChevronRight size={16} /> : <ChevronDown size={16} />}
                            </button>
                            <h3 className="text-sm font-medium text-gray-700 uppercase tracking-wider">
                                {section.label}
                            </h3>
                        </div>

                        {/* Section Content */}
                        {!isCollapsed && (
                            <div className="p-6">
                                <div className={`grid gap-x-12 gap-y-6 ${section.columns === 2 ? 'grid-cols-1 md:grid-cols-2' : 'grid-cols-1'}`}>
                                    {section.fields.map(fieldName => {
                                        const field = objectMetadata.fields.find(f => f.api_name === fieldName);

                                        // Skip invalid fields (deleted or not in schema)
                                        if (!field) return null;

                                        // Skip fields user doesn't have Read permission for
                                        if (!hasFieldPermission(objectMetadata.api_name, field.api_name, 'readable')) {
                                            return null;
                                        }

                                        const canEdit = !isReadOnly &&
                                            hasFieldPermission(objectMetadata.api_name, field.api_name, 'editable');

                                        return (
                                            <div key={fieldName} className="group">
                                                <div className="text-xs text-gray-500 font-medium mb-1.5 ml-1">
                                                    {field.label}
                                                    {field.required && <span className="text-red-500 ml-0.5">*</span>}
                                                </div>
                                                <div className="min-h-[30px]">
                                                    <InlineEditField
                                                        objectApiName={objectMetadata.api_name}
                                                        recordId={record.id}
                                                        field={field}
                                                        value={record[fieldName]}
                                                        record={record}
                                                        onUpdate={onUpdate}
                                                        isEditable={canEdit}
                                                        onNavigate={onNavigate}
                                                    />
                                                </div>
                                            </div>
                                        );
                                    })}
                                </div>
                            </div>
                        )}
                    </div>
                );
            })}
        </div>
    );
};
