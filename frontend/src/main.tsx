import React from "react";
import { createRoot } from "react-dom/client";

function App() {
  return <div>Chat App</div>;
}

const rootEl = document.getElementById("root");
if (rootEl) {
  createRoot(rootEl).render(
    <React.StrictMode>
      <App />
    </React.StrictMode>
  );
}
