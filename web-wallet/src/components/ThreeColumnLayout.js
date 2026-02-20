import React from "react";

export default function ThreeColumnLayout({ left, center, right, header }) {
  return (
    <div className="min-h-screen bg-dark-900 text-dark-50">
      {/* Header */}
      {header && (
        <header className="border-b border-dark-700 bg-dark-850">
          {header}
        </header>
      )}

      {/* Three-column grid */}
      <div className="flex flex-col lg:flex-row min-h-[calc(100vh-64px)]">
        {/* Left sidebar */}
        <aside className="w-full lg:w-72 xl:w-80 border-b lg:border-b-0 lg:border-r border-dark-700 bg-dark-850 overflow-y-auto">
          <div className="p-4">{left}</div>
        </aside>

        {/* Center / Main content */}
        <main className="flex-1 overflow-y-auto">
          <div className="p-4 lg:p-6 max-w-3xl mx-auto">{center}</div>
        </main>

        {/* Right sidebar */}
        <aside className="w-full lg:w-72 xl:w-80 border-t lg:border-t-0 lg:border-l border-dark-700 bg-dark-850 overflow-y-auto">
          <div className="p-4">{right}</div>
        </aside>
      </div>
    </div>
  );
}
