import {createFileRoute} from '@tanstack/react-router'
import {useSuspenseQuery} from "@tanstack/react-query";
import {getChart} from "@/actions/rest.ts";
import Chart, {ChartColors} from "@/components/Chart.tsx";

export const Route = createFileRoute('/')({
  component: App,
  pendingComponent: () => <div>Loading...</div>,
  errorComponent: () => <div>Error</div>,
  loader: async ({ context: { queryClient }}) => {
    await Promise.all([
      await queryClient.prefetchQuery({
        queryKey: ['chart'],
        queryFn: () => getChart("BTCUSD", 1440)
      }),
    ]);
  }
});

/*
{
"type":"kraken",
"payload":{
  "channel":"ohlc",
  "data":[{
    "symbol":"BTC/USD",
    "open":107360.0,
    "high":108289.0,
    "low":106677.4,
    "close":107323.1,
    "trades":22607,
    "volume":759.97804600,
    "vwap":107493.5,
    "interval_begin":"2025-06-26T00:00:00.000000000Z",
    "interval":1440,
    "timestamp":"2025-06-27T00:00:00.000000Z"
  }],
  "type":"update",
  "timestamp":"2025-06-26T21:29:29.817268970Z"
  }
}
 */

function App() {
  const chartQuery = useSuspenseQuery({
    queryKey: ['chart'],
    queryFn: () => getChart("BTCUSD", 1440)
  })
  return (
    <>
      {
        chartQuery.data ?
            <Chart chartData={chartQuery.data.XXBTZUSD} chartColors={new ChartColors()} />
        : null
      }
    </>
  )
}
