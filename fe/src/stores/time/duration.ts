export function formatDuration(durationString: string): string {
    const time = durationString.slice(11, -5);
    const parts = time.split(":");
    const hours = +parts[0];
    const minutes = +parts[1];
    const seconds = Math.round(+parts[2]);

    if (hours === 0) {
        return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
    }
    return `${hours}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
}
