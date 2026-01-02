
export function validString(val: unknown): string | '' {
    return typeof val === 'string' ? val : '';
}
