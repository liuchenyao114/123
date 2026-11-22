import { showError } from './utils';
import axios from 'axios';

export const API = axios.create({
  baseURL: process.env.REACT_APP_SERVER ? process.env.REACT_APP_SERVER : '',
});

API.interceptors.response.use(
  (response) => response,
  (error) => {
    showError(error);
    // 必须把错误继续向下传递，否则调用方拿到 undefined 会在 res.data 处抛 TypeError
    return Promise.reject(error);
  }
);
