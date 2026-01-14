import axios, { type AxiosInstance, type AxiosResponse, type AxiosError } from 'axios';
import { toast } from 'sonner';
import { useAuthStore } from '@/store/useAuthStore';

// Define the standard response structure from backend
interface ApiResponse<T = any> {
  code: number;
  msg: string;
  data: T;
}

const request: AxiosInstance = axios.create({
  baseURL: '/api', // Vite proxy will handle this
  timeout: 10000,
});

// Request Interceptor
request.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response Interceptor
request.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const { code, msg, data } = response.data;

    // 200 means success in business logic
    if (code === 200) {
      return data as any; // Return only the data part
    } else {
      // Business error
      toast.error(msg || 'Request failed');
      return Promise.reject(new Error(msg || 'Request failed'));
    }
  },
  (error: AxiosError) => {
    // HTTP errors
    let message = 'Something went wrong';
    if (error.response) {
      const { status } = error.response;
      switch (status) {
        case 401:
          message = 'Unauthorized. Please login again.';
          useAuthStore.getState().logout();
          break;
        case 403:
          message = 'Forbidden.';
          break;
        case 404:
          message = 'Resource not found.';
          break;
        case 500:
          message = 'Internal server error.';
          break;
        default:
          message = `Error: ${status}`;
      }
    } else if (error.request) {
      message = 'No response received from server.';

    } else {
      message = error.message;
    }

    toast.error(message);
    return Promise.reject(error);
  }
);

export default request;
