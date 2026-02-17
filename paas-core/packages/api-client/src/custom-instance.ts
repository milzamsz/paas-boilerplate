import axios, { AxiosRequestConfig } from "axios";

const AXIOS_INSTANCE = axios.create({
    baseURL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080",
    withCredentials: true,
    headers: {
        "Content-Type": "application/json",
    },
});

export const customInstance = <T>(config: AxiosRequestConfig): Promise<T> => {
    const source = axios.CancelToken.source();
    const promise = AXIOS_INSTANCE({
        ...config,
        cancelToken: source.token,
    }).then(({ data }) => data);

    // @ts-ignore
    promise.cancel = () => {
        source.cancel("Query was cancelled");
    };

    return promise;
};

export default customInstance;
