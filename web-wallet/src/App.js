import React from "react";
import { Link } from "react-router-dom";

function App() {
    return (
        <div className="min-h-screen bg-gray-900 text-white">
            <h1 className="text-4xl font-bold text-center py-4">TrueRepublic Web Wallet</h1>
            <nav className="flex justify-center gap-4 py-4">
                <Link to="/wallet" className="px-4 py-2 bg-blue-500 rounded hover:bg-blue-600">Wallet</Link>
                <Link to="/governance" className="px-4 py-2 bg-blue-500 rounded hover:bg-blue-600">Governance</Link>
                <Link to="/dex" className="px-4 py-2 bg-blue-500 rounded hover:bg-blue-600">DEX</Link>
            </nav>
        </div>
    );
}

export default App;
