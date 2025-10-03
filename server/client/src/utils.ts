const dtFormat = new Intl.DateTimeFormat(undefined, {
    dateStyle: 'short',
    timeStyle: 'short',
});

export function ensureDate(dt: string | Date): Date {
    if (typeof (dt) === 'string') {
        return new Date(dt);
    }
    return dt;
}

export function formatDate(dt: string | Date): string {
    return dtFormat.format(ensureDate(dt));
}
