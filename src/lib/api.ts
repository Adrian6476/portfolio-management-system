import axios, { type AxiosRequestConfig } from 'axios';
import type { ApiResponse } from '@/types/portfolio';

const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json',
  },
});

// 请求拦截器
apiClient.interceptors.request.use(config => {
  const token = localStorage.getItem('authToken');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// 响应拦截器
apiClient.interceptors.response.use(
  response => ({
    ...response,
    success: true,
    timestamp: Date.now(),
  }),
  error => ({
    success: false,
    error: error.response?.data?.message || error.message,
    timestamp: Date.now(),
  })
);

// 强类型的API方法
export const fetchPortfolioSummary = async (): Promise<ApiResponse<PortfolioSummaryResponse>> => {
  return apiClient.get('/portfolio/summary');
};

export const fetchHoldings = async (): Promise<ApiResponse<Holding[]>> => {
  return apiClient.get('/portfolio/holdings');
};

export const fetchCurrentPrice = async (symbol: string): Promise<ApiResponse<CurrentPriceResponse>> => {
  return apiClient.get(`/market/price/${symbol}`);
};

export default apiClient;