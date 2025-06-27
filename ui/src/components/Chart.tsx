import {createChart, CandlestickSeries, ColorType, type CandlestickData, type UTCTimestamp} from "lightweight-charts";
import {useEffect, useRef} from "react";
import {type ChartData, DestructureChartData, type LiveChartData} from "@/lib/types.ts";
import {getUTCDate} from "@/lib/helpers.ts";

const ws_url = import.meta.env.MODE === "development" ? "ws://localhost:1323/ws" : "/ws";

export class ChartColors {
    backgroundColor = 'white';
    lineColor = '#2962FF'
    textColor = 'black';
    areaTopColor = '#2962FF';
    areaBottomColor = 'rgba(41, 98, 255, 0.28)';
}

function ToCandlestickData(data: ChartData[]): CandlestickData[] {
    return data.map((e)=> {
        const d = DestructureChartData(e)
        return {
            time: d.time as UTCTimestamp,
            open: Number(d.open),
            high: Number(d.high),
            low: Number(d.low),
            close: Number(d.close),
        }
    })
}

export default function Chart({chartData, chartColors}: {chartData: ChartData[], chartColors: ChartColors}) {
    const chartContainerRef = useRef<HTMLDivElement>(null);
    useEffect(
        () => {
            const socket = new WebSocket(ws_url);
            socket.onopen = async () => {
                socket.send(JSON.stringify({
                    "type": "kraken",
                    "payload": {
                        "method": "subscribe",
                        "params": {
                            "channel": "ohlc",
                            "symbol": [
                                "BTC/USD"
                            ],
                            "interval": 1440
                        }
                    }
                }));
            }
            const handleResize = () => {
                if (!chartContainerRef.current) return;
                chart.applyOptions(
                    {
                        width: chartContainerRef.current.clientWidth,
                        height: window.innerHeight - 40
                    }
                );
            };
            if (!chartContainerRef.current) return;
            const chart = createChart(chartContainerRef.current, {
                layout: {
                    background: { type: ColorType.Solid, color: chartColors.backgroundColor },
                    textColor: chartColors.textColor,
                },
                width: chartContainerRef.current.clientWidth,
                height: window.innerHeight - 40
            });
            chart.timeScale().fitContent();
            const newSeries = chart.addSeries(CandlestickSeries, { baseLineColor: chartColors.lineColor });
            newSeries.setData(ToCandlestickData(chartData));
            console.log(newSeries);
            window.addEventListener('resize', handleResize);
            socket.onmessage = (e: MessageEvent) => {
                const data = JSON.parse(e.data) as LiveChartData;
                if (data.payload.channel !== 'ohlc') return;
                console.log(data)
                const index = data.payload.data.length - 1
                //TODO: The Date takes TZ into account, undo that so it is UTC+0
                const d = getUTCDate(new Date(data.payload.data[index].interval_begin))
                console.log(d)
                const ds = `${d.getFullYear()}-${(d.getMonth() + 1).toString().padStart(2, '0')}-${d.getDate().toString().padStart(2, '0')}`;
                console.log(ds)
                newSeries.update({
                    time: ds,
                    open: Number(data.payload.data[index].open),
                    high: Number(data.payload.data[index].high),
                    low: Number(data.payload.data[index].low),
                    close: Number(data.payload.data[index].close),
                })
            };
            return () => {
                window.removeEventListener('resize', handleResize);
                chart.remove();
                socket.onmessage = null;
                if (socket.readyState !== WebSocket.OPEN) return;
                socket.send(JSON.stringify({
                    "type": "kraken",
                    "payload": {
                        "method": "unsubscribe",
                        "params": {
                            "channel": "ohlc",
                            "symbol": [
                                "BTC/USD"
                            ],
                            "interval": 1440
                        }
                    }
                }));
                socket.close()
            };
        },
        [chartData, chartColors, chartContainerRef.current]
    );

    return (
        <div ref={chartContainerRef}/>
    )
}