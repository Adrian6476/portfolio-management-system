// This file will contain the shared API client setup using Axios.
// Developer A will set up the base Axios instance and helper functions here.
import axios from 'axios';

const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

export default apiClient;
