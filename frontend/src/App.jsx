import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import ModelManagement from './components/ModelManagement';
import Chat from './components/Chat';

function App() {
  return (
    <Router>
      <div className="app">
        <Routes>
          <Route path="/" element={<ModelManagement />} />
          <Route path="/chat" element={<Chat />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;
