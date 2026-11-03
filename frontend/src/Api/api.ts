import axios from "axios";
import { API_URL } from "../Config/env";

export const api = axios.create({
  baseURL: API_URL,
  headers: {
    "Content-Type": "application/json",
  },
});

api.interceptors.request.use((config) => {
  const token = sessionStorage.getItem("admin_token");

  console.log("API request:", config.url);
  console.log("Token exists:", Boolean(token));

  if (token) {
    config.headers.set("Authorization", `Bearer ${token}`);
  }

  return config;
});