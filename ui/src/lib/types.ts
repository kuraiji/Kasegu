export type ChartData = [time: number, open: string, high: string, low: string, close: string, vwap: string, volume: string, count: number];
export interface CDDestructured {
    time: number;
    open: string;
    high: string;
    low: string;
    close: string;
    vwap: string;
    volume: string;
    count: number;
}
export function DestructureChartData(data: ChartData): CDDestructured {
    return {
        time: data[0],
        open: data[1],
        high: data[2],
        low: data[3],
        close: data[4],
        vwap: data[5],
        volume: data[6],
        count: data[7],
    }
}