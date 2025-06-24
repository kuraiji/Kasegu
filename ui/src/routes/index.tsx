import {createFileRoute} from '@tanstack/react-router'
import {useSuspenseQuery} from "@tanstack/react-query";
import {getChart} from "@/actions/rest.ts";
import Chart, {ChartColors} from "@/components/Chart.tsx";
import {useEffect} from "react";
import {delay} from "@/lib/helpers.ts"

export const Route = createFileRoute('/')({
  component: App,
  pendingComponent: () => <div>Loading...</div>,
  errorComponent: () => <div>Error</div>,
  loader: async ({ context: { queryClient, socket }}) => {
    await Promise.all([
      await queryClient.prefetchQuery({
        queryKey: ['chart'],
        queryFn: () => getChart("BTCUSD", 1440)
      }),
    ]);
    return socket;
  }
});

function App() {
  const chartQuery = useSuspenseQuery({
    queryKey: ['chart'],
    queryFn: () => getChart("BTCUSD", 1440)
  })
  const socket = Route.useLoaderData();
  useEffect(() => {
    socket.onmessage = (e: MessageEvent) => {
      console.log("readMessage: ", e.data);
    };
    const test = async () => {
      await delay(1000);
      socket.send(JSON.stringify("Poop"));
    }
    test()
    return () => {
      socket.onmessage = null;
    };
  }, []);

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
