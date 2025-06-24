import {createChart, CandlestickSeries, ColorType, type CandlestickData, type UTCTimestamp} from "lightweight-charts";
import { useEffect, useRef } from "react";
import {type ChartData, DestructureChartData} from "@/lib/types.ts";

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

export default function Chart({chartData, chartColors}: {chartData: ChartData[], chartColors: ChartColors }) {
    const chartContainerRef = useRef<HTMLDivElement>(null);

    useEffect(
        () => {
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
            console.log(chartContainerRef.current.clientHeight);
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

            window.addEventListener('resize', handleResize);

            return () => {
                window.removeEventListener('resize', handleResize);
                chart.remove();
            };
        },
        [chartData, chartColors, chartContainerRef.current]
    );

    return (
        <div ref={chartContainerRef}/>
    )
}