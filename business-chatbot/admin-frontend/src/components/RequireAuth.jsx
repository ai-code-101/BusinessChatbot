import { Navigate } from "react-router-dom";
import { isAuthenticated } from "../api/client.js";

export default function RequireAuth({ children }) {
  if (!isAuthenticated()) {
    return <Navigate to="/login" replace />;
  }
  return children;
}
