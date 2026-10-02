const dtFormat = new Intl.DateTimeFormat(undefined, {
    formatMatcher: 'best fit',
    localeMatcher: 'best fit',
    dateStyle: 'short',
    timeStyle: 'short',
});

const durationFormat = new Intl.DurationFormat(undefined, {
    style: 'short',
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

export function formatDuration(from: string | Date, to: string | Date): string {
    const durationSec = Math.floor((ensureDate(to).valueOf() - ensureDate(from).valueOf()) / 1000);
    const hours = Math.floor(durationSec / 3600)
    const minutes = Math.floor((durationSec - hours * 3600) / 60)
    const seconds = durationSec % 60;
    return durationFormat.format({ hours, minutes, seconds })
}

export function formatPercent(p: number): string {
    return percentFormat.format(p);
}
