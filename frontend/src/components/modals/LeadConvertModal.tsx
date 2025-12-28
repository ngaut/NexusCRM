import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { actionAPI } from '../../infrastructure/api/actions'; // We need to create this or add to dataAPI
import { useNotification } from '../../contexts/NotificationContext';
import { Loader2, Check, X, ArrowRight } from 'lucide-react';
import type { SObject } from '../../types';

interface LeadConvertModalProps {
    isOpen: boolean;
    onClose: () => void;
    lead: SObject;
    onSuccess?: () => void;
}

export const LeadConvertModal: React.FC<LeadConvertModalProps> = ({
    isOpen,
    onClose,
    lead,
    onSuccess
}) => {
    const navigate = useNavigate();
    const { success, error: showError } = useNotification();
    const [loading, setLoading] = useState(false);

    if (!isOpen) return null;

    const handleConvert = async () => {
        setLoading(true);
        try {
            // Execute the system 'lead_convert' action
            // We pass the lead.id as the recordId
            await actionAPI.executeAction('lead_convert', {
                recordId: lead.id,
                objectApiName: 'lead',
                contextRecord: lead
            });

            success('Lead Converted', 'Lead has been successfully converted.');
            onSuccess?.();
            onClose();

            // Navigate to the new Account? Or stays on Lead?
            // Usually we go to the Account or the Contact.
            // Since we don't know the new IDs easily without parsing specific response (which is generic),
            // We'll just close and let the parent handle refresh.
            // But ideally, we should get the result. 
            // The executeAction returns a success message, but we might want to return the results map.
            // For now, let's just refresh current page or go to Leads list.
            navigate('/object/lead');

        } catch (err) {
            showError('Conversion Failed', err instanceof Error ? err.message : 'An error occurred during conversion.');
        } finally {
            setLoading(false);
        }
    };

    const companyName = (lead.company as string) || 'Unknown Company';
    const contactName = (lead.name as string) || `${lead.first_name || ''} ${lead.last_name || ''}`.trim();
    const opportunityName = `${companyName} - Deal`;

    return (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
            <div className="bg-white/95 backdrop-blur-2xl rounded-3xl shadow-2xl w-full max-w-lg overflow-hidden border border-white/40">
                <div className="px-6 py-4 border-b border-slate-200 bg-slate-50 flex justify-between items-center">
                    <h2 className="text-xl font-bold text-slate-800">Convert Lead</h2>
                    <button onClick={onClose} className="text-slate-400 hover:text-slate-600">
                        <X size={24} />
                    </button>
                </div>

                <div className="p-6 space-y-6">
                    <div className="bg-blue-50 border border-blue-100 rounded-lg p-4 text-sm text-blue-800">
                        Converting this lead will create the following records:
                    </div>

                    <div className="space-y-4">
                        {/* Account Preview */}
                        <div className="flex items-center gap-4 p-3 border border-slate-200 rounded-lg">
                            <div className="bg-purple-100 p-2 rounded-lg">
                                <span className="text-xl">🏢</span>
                            </div>
                            <div>
                                <div className="text-xs text-slate-500 uppercase font-semibold">Account</div>
                                <div className="font-medium text-slate-900">{companyName}</div>
                            </div>
                        </div>

                        {/* Contact Preview */}
                        <div className="flex items-center gap-4 p-3 border border-slate-200 rounded-lg">
                            <div className="bg-green-100 p-2 rounded-lg">
                                <span className="text-xl">👤</span>
                            </div>
                            <div>
                                <div className="text-xs text-slate-500 uppercase font-semibold">Contact</div>
                                <div className="font-medium text-slate-900">{contactName}</div>
                            </div>
                        </div>

                        {/* Opportunity Preview */}
                        <div className="flex items-center gap-4 p-3 border border-slate-200 rounded-lg">
                            <div className="bg-orange-100 p-2 rounded-lg">
                                <span className="text-xl">💰</span>
                            </div>
                            <div>
                                <div className="text-xs text-slate-500 uppercase font-semibold">Opportunity</div>
                                <div className="font-medium text-slate-900">{opportunityName}</div>
                            </div>
                        </div>
                    </div>
                </div>

                <div className="px-6 py-4 bg-slate-50 border-t border-slate-200 flex justify-end gap-3">
                    <button
                        onClick={onClose}
                        disabled={loading}
                        className="px-4 py-2 text-slate-700 border border-slate-300 rounded-lg hover:bg-slate-100 font-medium disabled:opacity-50"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleConvert}
                        disabled={loading}
                        className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium disabled:opacity-50 flex items-center gap-2 shadow-sm"
                    >
                        {loading ? (
                            <>
                                <Loader2 size={18} className="animate-spin" />
                                Converting...
                            </>
                        ) : (
                            <>
                                <span>Convert</span>
                                <ArrowRight size={18} />
                            </>
                        )}
                    </button>
                </div>
            </div>
        </div>
    );
};
