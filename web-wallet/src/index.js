import React from "react";
import ReactDOM from "react-dom";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import App from "./App";
import Wallet from "./pages/Wallet";
import Governance from "./pages/Governance";
import Dex from "./pages/Dex";

ReactDOM.render(
    <Router>
        <Routes>
            <Route path="/" element={<App />} />
            <Route path="/wallet" element={<Wallet />} />
            <Route path="/governance" element={<Governance />} />
            <Route path="/dex" element={<Dex />} />
        </Routes>
    </Router>,
    document.getElementById("root")
);
