import React, { useState, useRef, useEffect } from 'react';
import axios from 'axios';

const API_BASE_URL = process.env.NODE_ENV === 'production' 
  ? 'http://localhost:8080' 
  : 'http://localhost:8080';

function App() {
  const [messages, setMessages] = useState([]);
  const [inputMessage, setInputMessage] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [modelName, setModelName] = useState('mistral');
  const [modelStatus, setModelStatus] = useState({ type: '', message: '' });
  const [isModelLoading, setIsModelLoading] = useState(false);
  const chatHistoryRef = useRef(null);

  useEffect(() => {
    // Scroll to bottom when new messages are added
    if (chatHistoryRef.current) {
      chatHistoryRef.current.scrollTop = chatHistoryRef.current.scrollHeight;
    }
  }, [messages]);

  const createModel = async () => {
    if (!modelName.trim()) {
      setModelStatus({ type: 'error', message: 'Please enter a model name' });
      return;
    }

    setIsModelLoading(true);
    setModelStatus({ type: 'loading', message: `Creating ${modelName} model container...` });

    try {
      const response = await axios.post(`${API_BASE_URL}/create-dockerfile`, {
        model: modelName
      });

      setModelStatus({ 
        type: 'success', 
        message: `${modelName} model is ready! You can now start chatting.` 
      });
      
      // Clear any existing messages when a new model is created
      setMessages([]);
    } catch (error) {
      console.error('Error creating model:', error);
      setModelStatus({ 
        type: 'error', 
        message: error.response?.data?.error || 'Failed to create model' 
      });
    } finally {
      setIsModelLoading(false);
    }
  };

  const sendMessage = async () => {
    if (!inputMessage.trim() || isLoading) return;

    const userMessage = inputMessage.trim();
    setInputMessage('');
    setIsLoading(true);

    // Add user message to chat
    setMessages(prev => [...prev, { type: 'user', content: userMessage }]);

    try {
      const response = await axios.post(`${API_BASE_URL}/chat`, {
        message: userMessage
      });

      // Add assistant response to chat
      setMessages(prev => [...prev, { 
        type: 'assistant', 
        content: response.data.response 
      }]);
    } catch (error) {
      console.error('Error sending message:', error);
      const errorMessage = error.response?.data?.error || 'Failed to get response from model';
      setMessages(prev => [...prev, { 
        type: 'error', 
        content: errorMessage 
      }]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  const renderMessage = (message, index) => {
    return (
      <div key={index} className={`message ${message.type}`}>
        {message.content}
      </div>
    );
  };

  return (
    <div className="app">
      <div className="header">
        <h1>🤖 OwnGPT</h1>
        <p>Your Personal AI Assistant</p>
        
        <div className="model-selector">
          <input
            type="text"
            value={modelName}
            onChange={(e) => setModelName(e.target.value)}
            placeholder="Enter model name (e.g., mistral, llama2)"
            disabled={isModelLoading}
          />
          <button 
            className="btn-primary"
            onClick={createModel}
            disabled={isModelLoading}
          >
            {isModelLoading ? 'Creating...' : 'Create Model'}
          </button>
        </div>

        {modelStatus.message && (
          <div className={`status ${modelStatus.type}`}>
            {modelStatus.message}
          </div>
        )}
      </div>

      <div className="chat-container">
        <div className="chat-history" ref={chatHistoryRef}>
          {messages.length === 0 && !isLoading && (
            <div className="welcome-message">
              {modelStatus.type === 'success' 
                ? "Start a conversation with your AI assistant!" 
                : "Create a model first, then start chatting!"}
            </div>
          )}
          
          {messages.map(renderMessage)}
          
          {isLoading && (
            <div className="message assistant">
              <div className="loading-indicator">
                <span>AI is thinking</span>
                <div className="loading-dots">
                  <span></span>
                  <span></span>
                  <span></span>
                </div>
              </div>
            </div>
          )}
        </div>

        <div className="chat-input-container">
          <div className="chat-input">
            <textarea
              value={inputMessage}
              onChange={(e) => setInputMessage(e.target.value)}
              onKeyPress={handleKeyPress}
              placeholder="Type your message here..."
              disabled={isLoading || modelStatus.type !== 'success'}
              rows="1"
            />
            <button 
              onClick={sendMessage}
              disabled={isLoading || !inputMessage.trim() || modelStatus.type !== 'success'}
            >
              Send
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;
