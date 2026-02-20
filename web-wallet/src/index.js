import React from "react";
import { createRoot } from "react-dom/client";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import "./index.css";
import App from "./App";
import Wallet from "./pages/Wallet";
import Governance from "./pages/Governance";
import Dex from "./pages/Dex";

const root = createRoot(document.getElementById("root"));
root.render(
    <Router>
        <Routes>
            <Route path="/" element={<App />} />
            <Route path="/wallet" element={<Wallet />} />
            <Route path="/governance" element={<Governance />} />
            <Route path="/dex" element={<Dex />} />
        </Routes>
    </Router>
);
