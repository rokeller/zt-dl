const dtFormat = new Intl.DateTimeFormat(undefined, {
    formatMatcher: 'best fit',
    localeMatcher: 'best fit',
    dateStyle: 'short',
    timeStyle: 'short',
});

const percentFormat = new Intl.NumberFormat(undefined, {
    localeMatcher: 'best fit',
    style: 'percent',
    minimumFractionDigits: 1,
    maximumFractionDigits: 1,
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

export function formatPercent(p: number): string {
    return percentFormat.format(p);
}
