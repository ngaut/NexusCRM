import React from 'react';
import { CheckCircle, Circle, Clock, AlertCircle } from 'lucide-react';

export interface StepInfo {
    id: string;
    name: string;
    order: number;
    type: 'action' | 'approval' | 'decision';
    status: 'pending' | 'completed' | 'current' | 'skipped';
}

interface StepProgressIndicatorProps {
    steps: StepInfo[];
    currentStepId?: string;
    className?: string;
}

/**
 * Visual step progress indicator for multi-step approval flows.
 * Shows completed, current, and pending steps.
 */
export function StepProgressIndicator({ steps, currentStepId, className = '' }: StepProgressIndicatorProps) {
    if (!steps || steps.length === 0) {
        return null;
    }

    const sortedSteps = [...steps].sort((a, b) => a.order - b.order);

    return (
        <div className={`flex items-center gap-1 ${className}`}>
            {sortedSteps.map((step, index) => {
                const isCurrent = step.id === currentStepId || step.status === 'current';
                const isCompleted = step.status === 'completed';
                const isPending = step.status === 'pending';
                const isSkipped = step.status === 'skipped';

                return (
                    <React.Fragment key={step.id}>
                        {/* Step indicator */}
                        <div className="flex flex-col items-center gap-1">
                            <div
                                className={`flex items-center justify-center w-6 h-6 rounded-full transition-all ${isCompleted
                                        ? 'bg-green-500 text-white'
                                        : isCurrent
                                            ? 'bg-amber-500 text-white ring-2 ring-amber-300 ring-offset-1'
                                            : isSkipped
                                                ? 'bg-gray-300 text-gray-500'
                                                : 'bg-gray-200 text-gray-400'
                                    }`}
                                title={`Step ${step.order}: ${step.name} (${step.type})`}
                            >
                                {isCompleted ? (
                                    <CheckCircle className="w-4 h-4" />
                                ) : isCurrent ? (
                                    <Clock className="w-4 h-4 animate-pulse" />
                                ) : isSkipped ? (
                                    <AlertCircle className="w-3 h-3" />
                                ) : (
                                    <Circle className="w-3 h-3" />
                                )}
                            </div>
                            <span className={`text-[10px] max-w-[60px] truncate text-center ${isCurrent ? 'font-medium text-amber-700'
                                    : isCompleted ? 'text-green-700'
                                        : 'text-gray-500'
                                }`}>
                                {step.name}
                            </span>
                        </div>

                        {/* Connector line */}
                        {index < sortedSteps.length - 1 && (
                            <div className={`flex-1 h-0.5 min-w-[16px] ${isCompleted ? 'bg-green-400' : 'bg-gray-200'
                                }`} />
                        )}
                    </React.Fragment>
                );
            })}
        </div>
    );
}

/**
 * Compact version showing just "Step X of Y" with progress bar
 */
export function StepProgressCompact({
    currentStep,
    totalSteps,
    currentStepName
}: {
    currentStep: number;
    totalSteps: number;
    currentStepName?: string;
}) {
    const progress = totalSteps > 0 ? ((currentStep - 1) / totalSteps) * 100 : 0;

    return (
        <div className="space-y-1">
            <div className="flex items-center justify-between text-xs">
                <span className="text-gray-600">
                    Step <span className="font-semibold">{currentStep}</span> of {totalSteps}
                    {currentStepName && (
                        <span className="text-gray-500 ml-1">• {currentStepName}</span>
                    )}
                </span>
            </div>
            <div className="w-full h-1.5 bg-gray-200 rounded-full overflow-hidden">
                <div
                    className="h-full bg-gradient-to-r from-amber-400 to-amber-500 rounded-full transition-all duration-300"
                    style={{ width: `${progress}%` }}
                />
                <div
                    className="h-full bg-amber-300 -mt-1.5 animate-pulse rounded-full"
                    style={{ width: `${(currentStep / totalSteps) * 100}%`, opacity: 0.5 }}
                />
            </div>
        </div>
    );
}

export default StepProgressIndicator;
