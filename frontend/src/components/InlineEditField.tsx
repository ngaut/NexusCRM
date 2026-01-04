
import React, { useState, useRef, useEffect } from 'react';

import { FieldType } from '../core/constants/SchemaDefinitions';
import { KEYS } from '../core/constants';
import { UIRegistry } from '../registries/UIRegistry';
import { dataAPI } from '../infrastructure/api/data';
import { useNotification } from '../contexts/NotificationContext';
import type { SObject, FieldMetadata } from '../types';
import { Pencil, Check, X, Loader2 } from 'lucide-react';
import { isSystemField } from '../core/constants/CommonFields';

interface InlineEditFieldProps {
    objectApiName: string;
    recordId: string;
    field: FieldMetadata;
    value: unknown;
    record: SObject;
    onUpdate?: (fieldName: string, newValue: unknown) => void;
    isEditable?: boolean;
    onNavigate?: (obj: string, id: string) => void;
}

export const InlineEditField: React.FC<InlineEditFieldProps> = ({
    objectApiName,
    recordId,
    field,
    value,
    record,
    onUpdate,
    isEditable: propIsEditable = true,
    onNavigate
}) => {
    const { success, error: showError } = useNotification();
    const [isEditing, setIsEditing] = useState(false);
    const [editValue, setEditValue] = useState(value);
    const [saving, setSaving] = useState(false);
    const inputRef = useRef<HTMLInputElement>(null);

    // Reset edit value when value prop changes
    useEffect(() => {
        setEditValue(value);
    }, [value]);

    // Focus input when editing starts
    useEffect(() => {
        if (isEditing && inputRef.current) {
            inputRef.current.focus();
            inputRef.current.select();
        }
    }, [isEditing]);

    // Don't allow editing system fields
    const isSystemEditable = !field.is_system && !isSystemField(field.api_name);

    const canEdit = isSystemEditable && propIsEditable;

    const handleStartEdit = () => {
        if (!canEdit) return;
        setEditValue(value);
        setIsEditing(true);
    };

    const handleCancel = () => {
        setEditValue(value);
        setIsEditing(false);
    };

    const handleSave = async () => {
        if (editValue === value) {
            setIsEditing(false);
            return;
        }

        setSaving(true);
        try {
            await dataAPI.updateRecord(objectApiName, recordId, {
                [field.api_name]: editValue
            });
            success('Field Updated', `${field.label} has been updated.`);
            onUpdate?.(field.api_name, editValue);
            setIsEditing(false);
        } catch (err: unknown) {
            const message = err instanceof Error ? err.message : 'Failed to update field';
            showError('Update Failed', message);
        } finally {
            setSaving(false);
        }
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === KEYS.ENTER && !e.shiftKey) {
            e.preventDefault();
            handleSave();
        } else if (e.key === KEYS.ESCAPE) {
            handleCancel();
        }
    };

    const Renderer = UIRegistry.getFieldRenderer(field.type);

    // Render edit mode
    if (isEditing) {
        const renderEditInput = () => {
            const InputComponent = UIRegistry.getFieldInput(field.type);

            return (
                <InputComponent
                    field={field}
                    value={editValue}
                    onChange={(newValue: unknown) => setEditValue(newValue)}
                    onKeyDown={handleKeyDown}
                    disabled={saving}
                    autoFocus={true} // Add this prop to UIRegistry inputs if needed, or handle via ref
                />
            );
        };

        return (
            <div className="flex items-start gap-2">
                <div className="flex-1">
                    {renderEditInput()}
                </div>
                <div className="flex items-center gap-1">
                    <button
                        onClick={handleSave}
                        disabled={saving}
                        className="p-1 text-white bg-blue-600 hover:bg-blue-700 rounded disabled:opacity-50"
                        title="Save (Enter)"
                    >
                        {saving ? <Loader2 size={14} className="animate-spin" /> : <Check size={14} />}
                    </button>
                    <button
                        onClick={handleCancel}
                        disabled={saving}
                        className="p-1 text-slate-600 bg-slate-100 hover:bg-slate-200 rounded disabled:opacity-50"
                        title="Cancel (Escape)"
                    >
                        <X size={14} />
                    </button>
                </div>
            </div>
        );
    }

    // Render view mode with edit hover indicator
    return (
        <div
            className={`group flex items - center gap - 2 min - h - [24px] ${canEdit ? 'cursor-pointer hover:bg-blue-50 -mx-2 px-2 py-1 rounded transition-colors' : ''} `}
            onClick={handleStartEdit}
            title={canEdit ? 'Click to edit' : undefined}
        >
            <div className="flex-1">
                <Renderer field={field} value={value} record={record} variant="detail" onNavigate={onNavigate} />
            </div>
            {canEdit && (
                <Pencil
                    size={14}
                    className="text-slate-400 opacity-0 group-hover:opacity-100 transition-opacity"
                />
            )}
        </div>
    );
};
