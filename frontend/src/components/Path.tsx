import React from 'react';
import { Check } from 'lucide-react';
import { dataAPI } from '../infrastructure/api/data';
import { useNotification } from '../contexts/NotificationContext';
import type { SObject, FieldMetadata } from '../types';
import { COMMON_FIELDS } from '../core/constants';

interface PathProps {
    objectApiName: string;
    record: SObject;
    pathField: string;
    fields: FieldMetadata[];
    onUpdate: () => void;
}

export const Path: React.FC<PathProps> = ({
    objectApiName,
    record,
    pathField,
    fields,
    onUpdate
}) => {
    const { success, error: showError } = useNotification();
    const fieldMetadata = fields.find(f => f.api_name === pathField);

    if (!fieldMetadata || fieldMetadata.type !== 'Picklist' || !fieldMetadata.options) {
        return null;
    }

    const currentStatus = record[pathField] as string;
    const currentIndex = fieldMetadata.options.indexOf(currentStatus);

    const handleStepClick = async (status: string) => {
        if (status === currentStatus) return;

        try {
            await dataAPI.updateRecord(objectApiName, record[COMMON_FIELDS.ID] as string, {
                [pathField]: status
            });
            success('Status Updated', `Moved to "${status}"`);
            onUpdate();
        } catch {
            showError('Update Failed', 'Could not update status. Please try again.');
        }
    };

    return (
        <div className="bg-white border-b border-slate-200 px-6 py-4">
            <div className="flex items-center w-full overflow-x-auto pb-2">
                <div className="flex w-full min-w-max">
                    {fieldMetadata.options.map((option, index) => {
                        const isCompleted = index < currentIndex;
                        const isCurrent = index === currentIndex;
                        const isFuture = index > currentIndex;

                        let bgClass = 'bg-slate-100 text-slate-500';
                        let arrowClass = 'text-slate-100'; // Color of the arrow pointing right
                        let borderClass = 'border-l-white'; // Color of the left border (arrow tail)

                        if (isCompleted) {
                            bgClass = 'bg-green-500 text-white hover:bg-green-600 cursor-pointer';
                            arrowClass = 'text-green-500';
                            borderClass = 'border-l-white';
                        } else if (isCurrent) {
                            bgClass = 'bg-blue-600 text-white';
                            arrowClass = 'text-blue-600';
                            borderClass = 'border-l-white';
                        } else {
                            bgClass = 'bg-slate-100 text-slate-600 hover:bg-slate-200 cursor-pointer';
                            arrowClass = 'text-slate-100';
                            borderClass = 'border-l-white';
                        }

                        return (
                            <div
                                key={option}
                                onClick={() => !isCurrent && handleStepClick(option)}
                                className={`
                                    relative flex-1 flex items-center justify-center h-10 px-8 
                                    first:pl-4 transition-colors select-none
                                    ${bgClass}
                                `}
                                style={{
                                    clipPath: 'polygon(0% 0%, calc(100% - 1rem) 0%, 100% 50%, calc(100% - 1rem) 100%, 0% 100%, 1rem 50%)',
                                    marginLeft: index === 0 ? 0 : '-1rem'
                                }}
                            >
                                <span className="text-sm font-medium whitespace-nowrap z-10 flex items-center gap-2">
                                    {isCompleted && <Check size={14} />}
                                    {option}
                                </span>
                            </div>
                        );
                    })}
                </div>
            </div>

            {/* Mark as Current Button (Optional - for when we implement "Select then Mark") */}
            {/* Currently we just click the step to update */}
        </div>
    );
};
