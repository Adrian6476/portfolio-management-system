// This file will contain the shared API client setup using Axios.
// Developer A will set up the base Axios instance and helper functions here.
import axios from 'axios';

const getBaseURL = () => {
  const apiUrl = process.env.NEXT_PUBLIC_API_URL;
  
  if (!apiUrl) {
    throw new Error('NEXT_PUBLIC_API_URL environment variable is not set. Please check your environment configuration.');
  }
  
  return apiUrl;
};

const apiClient = axios.create({
  baseURL: getBaseURL(),
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 10000,
  maxRedirects: 5, // Allow up to 5 redirects
  validateStatus: function (status) {
    return status < 400; // Accept any status code less than 400
  },
});



export default apiClient;
