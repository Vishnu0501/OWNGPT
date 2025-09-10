import React, { useState, useRef, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import axios from 'axios';
import MessageFormatter from './MessageFormatter';

const API_BASE_URL = process.env.NODE_ENV === 'production' 
  ? 'http://localhost:8080' 
  : 'http://localhost:8080';

function Chat() {
  const navigate = useNavigate();
  const location = useLocation();
  const [messages, setMessages] = useState([]);
  const [inputMessage, setInputMessage] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [currentModel, setCurrentModel] = useState('');
  const [modelStatus, setModelStatus] = useState({ type: '', message: '' });
  const [responseTime, setResponseTime] = useState(null);
  const chatHistoryRef = useRef(null);

  useEffect(() => {
    // Get selected model from navigation state or check health
    const selectedModel = location.state?.selectedModel;
    if (selectedModel) {
      setCurrentModel(selectedModel);
      setModelStatus({ 
        type: 'success', 
        message: `Connected to ${selectedModel} model` 
      });
    } else {
      checkHealthStatus();
    }
  }, [location.state]);

  useEffect(() => {
    // Scroll to bottom when new messages are added
    if (chatHistoryRef.current) {
      chatHistoryRef.current.scrollTop = chatHistoryRef.current.scrollHeight;
    }
  }, [messages]);

  const checkHealthStatus = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL}/health`);
      if (response.data.model_running && response.data.model_name) {
        const modelNameFromContainer = response.data.model_name.replace('ollama-', '').replace('-container', '');
        setCurrentModel(modelNameFromContainer);
        setModelStatus({ 
          type: 'success', 
          message: `Connected to ${modelNameFromContainer} model` 
        });
      } else {
        setModelStatus({ 
          type: 'error', 
          message: 'No model is currently running. Please go to Model Management to start a model.' 
        });
      }
    } catch (error) {
      console.log('Health check failed:', error);
      setModelStatus({ 
        type: 'error', 
        message: 'Unable to connect to backend. Please check if the server is running.' 
      });
    }
  };

  const sendMessage = async () => {
    if (!inputMessage.trim() || isLoading) return;

    const userMessage = inputMessage.trim();
    setInputMessage('');
    setIsLoading(true);
    
    const startTime = Date.now();

    // Add user message to chat
    setMessages(prev => [...prev, { type: 'user', content: userMessage }]);

    // Add empty assistant message that will be updated with streaming content
    const assistantMessageIndex = Date.now();
    setMessages(prev => [...prev, { 
      type: 'assistant', 
      content: '',
      id: assistantMessageIndex,
      streaming: true
    }]);

    try {
      // Use streaming for faster response perception
      const response = await fetch(`${API_BASE_URL}/chat/stream`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ message: userMessage })
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let accumulatedContent = '';

      while (true) {
        const { value, done } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value);
        const lines = chunk.split('\n');

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const content = line.slice(6); // Remove 'data: ' prefix
            if (content && content !== '[DONE]') {
              accumulatedContent += content;
              
              // Update the streaming message
              setMessages(prev => prev.map(msg => 
                msg.id === assistantMessageIndex 
                  ? { ...msg, content: accumulatedContent }
                  : msg
              ));
            }
          } else if (line.startsWith('error: ')) {
            throw new Error(line.slice(7)); // Remove 'error: ' prefix
          }
        }
      }

      // Mark streaming as complete and calculate response time
      const endTime = Date.now();
      const responseTimeMs = endTime - startTime;
      setResponseTime(responseTimeMs);
      
      setMessages(prev => prev.map(msg => 
        msg.id === assistantMessageIndex 
          ? { ...msg, streaming: false, responseTime: responseTimeMs }
          : msg
      ));

    } catch (error) {
      console.error('Error sending message:', error);
      
      // Remove the empty streaming message and add error message
      setMessages(prev => prev.filter(msg => msg.id !== assistantMessageIndex));
      
      const errorMessage = error.message || 'Failed to get response from model';
      setMessages(prev => [...prev, { 
        type: 'error', 
        content: errorMessage 
      }]);
      
      // If no model is running, suggest going to model management
      if (errorMessage.includes('No model is currently running')) {
        setModelStatus({ 
          type: 'error', 
          message: 'No model is currently running. Please go to Model Management to start a model.' 
        });
      }
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

  const clearChat = () => {
    setMessages([]);
  };

  const goToModelManagement = () => {
    navigate('/');
  };

  const renderMessage = (message, index) => {
    const isUser = message.type === 'user';
    const isError = message.type === 'error';
    const isStreaming = message.streaming;
    
    return (
      <div key={message.id || index} className={`message ${message.type} ${isStreaming ? 'streaming' : ''}`}>
        <div className="message-avatar">
          {isUser ? '👤' : isError ? '⚠️' : '🤖'}
        </div>
        <div className="message-content">
          <div className="message-text">
            {isUser || isError ? (
              <>
                {message.content}
                {isStreaming && (
                  <span className="streaming-cursor">▊</span>
                )}
              </>
            ) : (
              <MessageFormatter 
                content={message.content} 
                isStreaming={isStreaming}
              />
            )}
            {isStreaming && !isUser && !isError && (
              <span className="streaming-cursor">▊</span>
            )}
          </div>
          <div className="message-time">
            {isStreaming ? 'Generating...' : new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
          </div>
        </div>
      </div>
    );
  };

  return (
    <div className="chat-page">
      <div className="header">
        <div className="header-content">
          <div className="header-left">
            <button className="back-button" onClick={goToModelManagement}>
              ← Back to Models
            </button>
            <div className="logo-section">
              <div className="logo">💬</div>
              <div className="title-section">
                <h1>Chat</h1>
                {currentModel && <p>with {currentModel}</p>}
              </div>
            </div>
          </div>
          
          <div className="header-actions">
            {messages.length > 0 && (
              <button className="btn-secondary" onClick={clearChat}>
                🗑️ Clear Chat
              </button>
            )}
            <button className="btn-secondary" onClick={goToModelManagement}>
              🔧 Switch Model
            </button>
            {currentModel && (
              <div className="current-model-badge">
                <span className="model-indicator"></span>
                {currentModel}
                {responseTime && (
                  <span className={`performance-indicator ${responseTime < 6000 ? 'fast' : 'slow'}`}>
                    {(responseTime/1000).toFixed(1)}s
                  </span>
                )}
              </div>
            )}
          </div>
        </div>

        {modelStatus.message && (
          <div className={`status ${modelStatus.type}`}>
            <div className="status-content">
              <span className="status-message">{modelStatus.message}</span>
              {modelStatus.type === 'error' && (
                <button className="btn-primary" onClick={goToModelManagement}>
                  Go to Model Management
                </button>
              )}
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
                  : "No model available"}
              </h3>
              <p>
                {modelStatus.type === 'success' 
                  ? "Ask me anything - I'm here to help!" 
                  : "Please go to Model Management to select and start a model"}
              </p>
              {modelStatus.type !== 'success' && (
                <button className="btn-primary" onClick={goToModelManagement}>
                  Go to Model Management
                </button>
              )}
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
                    : "Select a model first to start chatting..."
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

export default Chat;
