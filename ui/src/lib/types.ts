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

export interface LiveChartData {
    type: string;
    payload: {
        channel: string;
        type: string;
        timestamp: string;
        data: LiveChartPayload[]
    }
}

export interface LiveChartPayload {
    symbol: string;
    open: number;
    high: number;
    low: number;
    close: number;
    trades: number;
    volume: number;
    vwap: number;
    interval_begin: string;
    interval: number;
    timestamp: string;
}

export function LiveChartDataToChartData(liveData: LiveChartPayload): ChartData {
    return [
        (new Date(liveData.interval_begin).getTime()) / 1000,
        String(liveData.open),
        String(liveData.high),
        String(liveData.low),
        String(liveData.close),
        String(liveData.vwap),
        String(liveData.volume),
        liveData.trades
    ];
}