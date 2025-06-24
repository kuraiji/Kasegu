import axios from "axios";
import type {ChartData} from "@/lib/types.ts";

const api_url = import.meta.env.MODE === "development" ? "http://localhost:1323/api" : "/api";


export interface BTCChart{
    XXBTZUSD: ChartData[]
    last: number
}

export async function getChart(pair: string, interval: number): Promise<BTCChart | null> {
    try {
        return await axios.get(api_url + "/chart?pair=" + pair + "&interval=" + interval,
            {}).then(res => {
                if (res.status >= 400 && res.status <= 499) return null;
                return res.data;
            }
        )
    }
    catch(error) {
        return null;
    }
}