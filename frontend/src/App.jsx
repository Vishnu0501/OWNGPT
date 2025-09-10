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
  const [modelProgress, setModelProgress] = useState(0);
  const [currentModel, setCurrentModel] = useState('');
  const [showModelSuggestions, setShowModelSuggestions] = useState(false);
  const chatHistoryRef = useRef(null);

  const popularModels = [
    { name: 'mistral', description: 'Fast and efficient 7B model' },
    { name: 'llama2', description: 'Meta\'s powerful language model' },
    { name: 'codellama', description: 'Specialized for code generation' },
    { name: 'vicuna', description: 'Fine-tuned for conversations' },
    { name: 'orca-mini', description: 'Compact and fast model' },
    { name: 'neural-chat', description: 'Optimized for chat interactions' }
  ];

  useEffect(() => {
    // Scroll to bottom when new messages are added
    if (chatHistoryRef.current) {
      chatHistoryRef.current.scrollTop = chatHistoryRef.current.scrollHeight;
    }
  }, [messages]);

  useEffect(() => {
    // Check health status on component mount
    checkHealthStatus();
  }, []);

  const checkHealthStatus = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL}/health`);
      if (response.data.model_running && response.data.model_name) {
        const modelNameFromContainer = response.data.model_name.replace('ollama-', '').replace('-container', '');
        setCurrentModel(modelNameFromContainer);
        setModelStatus({ 
          type: 'success', 
          message: `${modelNameFromContainer} model is ready! You can now start chatting.` 
        });
      }
    } catch (error) {
      console.log('Health check failed:', error);
    }
  };

  const createModel = async () => {
    if (!modelName.trim()) {
      setModelStatus({ type: 'error', message: 'Please enter a model name' });
      return;
    }

    setIsModelLoading(true);
    setModelProgress(0);
    setModelStatus({ type: 'loading', message: `Creating ${modelName} model container...` });

    // Simulate progress updates during model creation
    const progressInterval = setInterval(() => {
      setModelProgress(prev => {
        if (prev < 90) return prev + 10;
        return prev;
      });
    }, 2000);

    try {
      const response = await axios.post(`${API_BASE_URL}/create-dockerfile`, {
        model: modelName
      });

      clearInterval(progressInterval);
      setModelProgress(100);
      setCurrentModel(modelName);
      
      const isExisting = response.data.already_exists;
      setModelStatus({ 
        type: 'success', 
        message: isExisting 
          ? `${modelName} model was already available and is now ready! ⚡` 
          : `${modelName} model has been created and is ready! You can now start chatting. 🎉`
      });
      
      // Clear any existing messages when a new model is created (only if it's actually new)
      if (!isExisting) {
        setMessages([]);
      }
    } catch (error) {
      clearInterval(progressInterval);
      console.error('Error creating model:', error);
      setModelStatus({ 
        type: 'error', 
        message: error.response?.data?.error || 'Failed to create model' 
      });
      setModelProgress(0);
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
    const isUser = message.type === 'user';
    const isError = message.type === 'error';
    
    return (
      <div key={index} className={`message ${message.type}`}>
        <div className="message-avatar">
          {isUser ? '👤' : isError ? '⚠️' : '🤖'}
        </div>
        <div className="message-content">
          <div className="message-text">
            {message.content}
          </div>
          <div className="message-time">
            {new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
          </div>
        </div>
      </div>
    );
  };

  return (
    <div className="app">
      <div className="header">
        <div className="header-content">
          <div className="logo-section">
            <div className="logo">🤖</div>
            <div className="title-section">
              <h1>OwnGPT</h1>
              <p>Your Personal AI Assistant</p>
            </div>
          </div>
          
          {currentModel && (
            <div className="current-model-badge">
              <span className="model-indicator"></span>
              {currentModel}
            </div>
          )}
        </div>
        
        <div className="model-selector">
          <div className="input-group">
            <div className="input-container">
              <input
                type="text"
                value={modelName}
                onChange={(e) => setModelName(e.target.value)}
                onFocus={() => setShowModelSuggestions(true)}
                onBlur={() => setTimeout(() => setShowModelSuggestions(false), 200)}
                placeholder="Enter model name (e.g., mistral, llama2, codellama)"
                disabled={isModelLoading}
                className="model-input"
              />
              {showModelSuggestions && (
                <div className="model-suggestions">
                  {popularModels
                    .filter(model => model.name.toLowerCase().includes(modelName.toLowerCase()))
                    .map((model, index) => (
                      <div 
                        key={index}
                        className="suggestion-item"
                        onClick={() => {
                          setModelName(model.name);
                          setShowModelSuggestions(false);
                        }}
                      >
                        <div className="suggestion-name">{model.name}</div>
                        <div className="suggestion-desc">{model.description}</div>
                      </div>
                    ))
                  }
                </div>
              )}
            </div>
            <button 
              className={`btn-primary ${isModelLoading ? 'loading' : ''}`}
              onClick={createModel}
              disabled={isModelLoading}
            >
              {isModelLoading ? (
                <span className="button-content">
                  <div className="spinner"></div>
                  Creating...
                </span>
              ) : (
                'Create Model'
              )}
            </button>
          </div>
        </div>

        {modelStatus.message && (
          <div className={`status ${modelStatus.type}`}>
            <div className="status-content">
              {modelStatus.type === 'loading' && (
                <div className="progress-container">
                  <div className="progress-bar">
                    <div 
                      className="progress-fill" 
                      style={{ width: `${modelProgress}%` }}
                    ></div>
                  </div>
                  <span className="progress-text">{modelProgress}%</span>
                </div>
              )}
              <span className="status-message">{modelStatus.message}</span>
            </div>
          </div>
        )}
      </div>

      <div className="chat-container">
        <div className="chat-history" ref={chatHistoryRef}>
          {messages.length === 0 && !isLoading && (
            <div className="welcome-message">
              <div className="welcome-icon">💬</div>
              <h3>
                {modelStatus.type === 'success' 
                  ? `Ready to chat with ${currentModel}!` 
                  : "Create a model to get started"}
              </h3>
              <p>
                {modelStatus.type === 'success' 
                  ? "Ask me anything - I'm here to help!" 
                  : "Choose from models like mistral, llama2, codellama, or any Ollama model"}
              </p>
            </div>
          )}
          
          {messages.map(renderMessage)}
          
          {isLoading && (
            <div className="message assistant typing">
              <div className="message-avatar">🤖</div>
              <div className="message-content">
                <div className="loading-indicator">
                  <span>AI is thinking</span>
                  <div className="loading-dots">
                    <span></span>
                    <span></span>
                    <span></span>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>

        <div className="chat-input-container">
          <div className="chat-input">
            <div className="input-wrapper">
              <textarea
                value={inputMessage}
                onChange={(e) => setInputMessage(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder={
                  modelStatus.type === 'success' 
                    ? "Type your message here..." 
                    : "Create a model first to start chatting..."
                }
                disabled={isLoading || modelStatus.type !== 'success'}
                rows="1"
                className="message-input"
              />
              <button 
                className={`send-button ${isLoading ? 'loading' : ''}`}
                onClick={sendMessage}
                disabled={isLoading || !inputMessage.trim() || modelStatus.type !== 'success'}
              >
                {isLoading ? (
                  <div className="send-spinner"></div>
                ) : (
                  <svg viewBox="0 0 24 24" width="20" height="20" fill="currentColor">
                    <path d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"/>
                  </svg>
                )}
              </button>
            </div>
            {modelStatus.type === 'success' && (
              <div className="input-hint">
                Press Enter to send, Shift+Enter for new line
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;
