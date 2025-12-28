import React from 'react';

interface FormulaConfigProps {
    formulaExpression: string;
    setFormulaExpression: (value: string) => void;
    formulaReturnType: string;
    setFormulaReturnType: (value: string) => void;
}

export const FormulaConfig: React.FC<FormulaConfigProps> = ({
    formulaExpression,
    setFormulaExpression,
    formulaReturnType,
    setFormulaReturnType
}) => {
    return (
        <div className="p-4 bg-yellow-50 rounded-lg border border-yellow-100 space-y-4">
            <div>
                <label className="block text-sm font-medium text-yellow-800 mb-1.5">
                    Formula Expression <span className="text-red-500">*</span>
                </label>
                <textarea
                    value={formulaExpression}
                    onChange={(e) => setFormulaExpression(e.target.value)}
                    placeholder="e.g., price * quantity"
                    rows={3}
                    className="w-full px-3 py-2 border border-yellow-200 rounded-lg focus:ring-2 focus:ring-yellow-500 focus:border-transparent text-sm font-mono"
                    required
                />
                <p className="mt-1 text-xs text-yellow-600">
                    Reference other fields by their API name. Supports math operators (+, -, *, /), functions (IF, ROUND, LEN, etc.)
                </p>
            </div>

            <div>
                <label className="block text-sm font-medium text-yellow-800 mb-1.5">
                    Return Type <span className="text-red-500">*</span>
                </label>
                <select
                    value={formulaReturnType}
                    onChange={(e) => setFormulaReturnType(e.target.value)}
                    className="w-full px-3 py-2 border border-yellow-200 rounded-lg focus:ring-2 focus:ring-yellow-500 focus:border-transparent text-sm"
                >
                    <option value="Number">Number</option>
                    <option value="Text">Text</option>
                    <option value="Currency">Currency</option>
                    <option value="Percent">Percent</option>
                    <option value="Date">Date</option>
                    <option value="Boolean">Checkbox</option>
                </select>
            </div>
        </div>
    );
};
