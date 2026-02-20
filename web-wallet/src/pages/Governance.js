import React from "react";
import { Navigate } from "react-router-dom";

// Governance is now the main App page at "/"
// Redirect /governance to / for backwards compatibility
function Governance() {
  return <Navigate to="/" replace />;
}

export default Governance;
