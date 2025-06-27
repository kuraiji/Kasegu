export function delay(ms: number) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

export function getUTCDate(date: Date) {
    const timezoneName = Intl.DateTimeFormat().resolvedOptions().timeZone
    let utcDate = new Date(date.toLocaleString('en-US', { timeZone: "UTC" }));
    let tzDate = new Date(date.toLocaleString('en-US', { timeZone: timezoneName }));
    let offset = utcDate.getTime() - tzDate.getTime();

    date.setTime( date.getTime() + offset );

    return date;
}