import { createFileRoute } from '@tanstack/react-router'
import {useSuspenseQuery} from "@tanstack/react-query";
import {getChart} from "@/actions/rest.ts";

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
    ])
  }
});

function App() {
  const chartQuery = useSuspenseQuery({
    queryKey: ['chart'],
    queryFn: () => getChart("BTCUSD", 1440)
  })
  console.log(chartQuery.data)
  return (
    <>
      {
        chartQuery.data ?
            <p>Pee</p>
        : null
      }
    </>
  )
}
