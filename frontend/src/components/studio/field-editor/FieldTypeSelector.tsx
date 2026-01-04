import { FIELD_TYPE_OPTIONS } from '../../../core/constants/ui/FieldUIConstants';
import { FieldType } from '../../../core/constants/SchemaDefinitions';

export type { FieldTypeOption } from '../../../core/constants/ui/FieldUIConstants';
export { FIELD_TYPE_OPTIONS }; // Re-export if needed or just use


interface FieldTypeSelectorProps {
    onSelect: (type: FieldType | 'MasterDetail') => void;
}

export const FieldTypeSelector: React.FC<FieldTypeSelectorProps> = ({ onSelect }) => {
    return (
        <div>
            <p className="text-sm text-slate-600 mb-4">Select the type of field to create:</p>
            <div className="grid grid-cols-3 gap-2">
                {FIELD_TYPE_OPTIONS.map(ft => (
                    <button
                        key={ft.type}
                        onClick={() => onSelect(ft.type === 'MasterDetail' ? 'MasterDetail' : ft.type)}
                        className={`flex flex-col items-center gap-2 p-4 border rounded-xl hover:border-blue-300 hover:bg-blue-50/50 transition-all text-center`}
                    >
                        <div className={`w-10 h-10 rounded-lg bg-${ft.color}-100 flex items-center justify-center`}>
                            <ft.icon size={20} className={`text-${ft.color}-600`} />
                        </div>
                        <div className="text-sm font-medium text-slate-800">{ft.label}</div>
                    </button>
                ))}
            </div>
        </div>
    );
};
