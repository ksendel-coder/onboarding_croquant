import axios from "axios";
import { API_URL } from "../Config/env";

export const SUBJECT_KEY = "croquant_subject_id";
export const ATTEMPT_PREFIX = "croquant_attempt_";

export const AUTH_TOKEN_KEY = "admin_token";
export const AUTH_EXPIRED_EVENT = "croquant:auth-expired";

export const api = axios.create({
  baseURL: API_URL,
  headers: { "Content-Type": "application/json" },
});