import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

const API_BASE_URL = process.env.NODE_ENV === 'production' 
  ? 'http://localhost:8080' 
  : 'http://localhost:8080';

function ModelManagement() {
  const navigate = useNavigate();
  const [availableModels, setAvailableModels] = useState([]);
  const [installedModels, setInstalledModels] = useState([]);
  const [customModelName, setCustomModelName] = useState('');
  const [isCreating, setIsCreating] = useState(false);
  const [createProgress, setCreateProgress] = useState(0);
  const [status, setStatus] = useState({ type: '', message: '' });
  const [activeTab, setActiveTab] = useState('available'); // 'available' or 'installed'
  const [systemInfo, setSystemInfo] = useState({ gpu_available: false, memory_limit: '4GB', message: '' });

  useEffect(() => {
    loadAvailableModels();
    loadInstalledModels();
    loadSystemInfo();
  }, []);

  const loadAvailableModels = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL}/available-models`);
      setAvailableModels(response.data.available_models || []);
    } catch (error) {
      console.error('Failed to load available models:', error);
    }
  };

  const loadInstalledModels = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL}/models`);
      setInstalledModels(response.data.models || []);
    } catch (error) {
      console.error('Failed to load installed models:', error);
    }
  };

  const loadSystemInfo = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL}/system-info`);
      setSystemInfo(response.data);
    } catch (error) {
      console.error('Failed to load system info:', error);
    }
  };

  const createModel = async (modelName) => {
    if (!modelName.trim()) {
      setStatus({ type: 'error', message: 'Please enter a model name' });
      return;
    }

    setIsCreating(true);
    setCreateProgress(0);
    setStatus({ type: 'loading', message: `Creating ${modelName} model...` });

    // Simulate progress updates
    const progressInterval = setInterval(() => {
      setCreateProgress(prev => {
        if (prev < 90) return prev + 10;
        return prev;
      });
    }, 2000);

    try {
      const response = await axios.post(`${API_BASE_URL}/create-dockerfile`, {
        model: modelName
      });

      clearInterval(progressInterval);
      setCreateProgress(100);
      
      const isExisting = response.data.already_exists;
      setStatus({ 
        type: 'success', 
        message: isExisting 
          ? `${modelName} model was already available and is now ready!` 
          : `${modelName} model has been created successfully!`
      });
      
      // Refresh installed models list
      await loadInstalledModels();
      
      // Clear custom model name if it was used
      if (customModelName === modelName) {
        setCustomModelName('');
      }
    } catch (error) {
      clearInterval(progressInterval);
      console.error('Error creating model:', error);
      setStatus({ 
        type: 'error', 
        message: error.response?.data?.error || 'Failed to create model' 
      });
      setCreateProgress(0);
    } finally {
      setIsCreating(false);
    }
  };

  const deleteModel = async (modelName) => {
    if (!window.confirm(`Are you sure you want to delete the ${modelName} model?`)) {
      return;
    }

    try {
      await axios.delete(`${API_BASE_URL}/models/${modelName}`);
      setStatus({ 
        type: 'success', 
        message: `${modelName} model deleted successfully` 
      });
      await loadInstalledModels();
    } catch (error) {
      console.error('Error deleting model:', error);
      setStatus({ 
        type: 'error', 
        message: error.response?.data?.error || 'Failed to delete model' 
      });
    }
  };

  const startChatWithModel = (modelName) => {
    navigate('/chat', { state: { selectedModel: modelName } });
  };

  return (
    <div className="model-management">
      <div className="header">
        <div className="header-content">
          <div className="logo-section">
            <div className="logo">🤖</div>
            <div className="title-section">
              <h1>OwnGPT</h1>
              <p>Model Management</p>
            </div>
          </div>
          
          <div className="header-actions">
            {installedModels.some(model => model.is_running) && (
              <button 
                className="btn-primary quick-chat"
                onClick={() => {
                  const runningModel = installedModels.find(model => model.is_running);
                  if (runningModel) startChatWithModel(runningModel.name);
                }}
              >
                💬 Quick Chat
              </button>
            )}
          </div>
        </div>
      </div>

      <div className="main-content">
        <div className="tabs">
          <button 
            className={`tab ${activeTab === 'available' ? 'active' : ''}`}
            onClick={() => setActiveTab('available')}
          >
            Available Models
          </button>
          <button 
            className={`tab ${activeTab === 'installed' ? 'active' : ''}`}
            onClick={() => setActiveTab('installed')}
          >
            Installed Models ({installedModels.length})
          </button>
        </div>

        {status.message && (
          <div className={`status ${status.type}`}>
            <div className="status-content">
              {status.type === 'loading' && (
                <div className="progress-container">
                  <div className="progress-bar">
                    <div 
                      className="progress-fill" 
                      style={{ width: `${createProgress}%` }}
                    ></div>
                  </div>
                  <span className="progress-text">{createProgress}%</span>
                </div>
              )}
              <span className="status-message">{status.message}</span>
            </div>
          </div>
        )}

        {activeTab === 'available' && (
          <div className="tab-content">
            <div className="custom-model-section">
              <h3>Create Custom Model</h3>
              <div className="system-info-banner">
                <div className="system-info-content">
                  <span className="system-info-icon">
                    {systemInfo.gpu_available ? '🚀' : '💻'}
                  </span>
                  <div className="system-info-text">
                    <strong>{systemInfo.gpu_available ? 'GPU Acceleration Available' : 'CPU Processing'}</strong>
                    <p>Memory limit: {systemInfo.memory_limit} | {systemInfo.message}</p>
                  </div>
                </div>
              </div>
              <p className="section-description">
                Enter any model name to pull and create it. You can use models from Ollama library 
                like llama2, mistral, codellama, or any other available model.
              </p>
              <div className="input-group">
                <input
                  type="text"
                  value={customModelName}
                  onChange={(e) => setCustomModelName(e.target.value)}
                  placeholder="Type any model name (e.g., llama3, gemma, phi3, qwen)"
                  disabled={isCreating}
                  className="model-input"
                  onKeyPress={(e) => {
                    if (e.key === 'Enter' && customModelName.trim() && !isCreating) {
                      createModel(customModelName);
                    }
                  }}
                />
                <button 
                  className={`btn-primary ${isCreating ? 'loading' : ''}`}
                  onClick={() => createModel(customModelName)}
                  disabled={isCreating || !customModelName.trim()}
                >
                  {isCreating ? 'Creating...' : 'Pull & Create'}
                </button>
              </div>
              <div className="model-suggestions">
                <span className="suggestions-label">Popular suggestions:</span>
                {['llama3', 'gemma', 'phi3', 'qwen', 'codellama', 'mistral:7b'].map(suggestion => (
                  <button
                    key={suggestion}
                    className="suggestion-tag"
                    onClick={() => setCustomModelName(suggestion)}
                    disabled={isCreating}
                  >
                    {suggestion}
                  </button>
                ))}
              </div>
            </div>

            <div className="popular-models-section">
              <h3>Popular Models</h3>
              <div className="models-grid">
                {availableModels.map((model, index) => (
                  <div key={index} className="model-card">
                    <div className="model-header">
                      <h4>{model.name}</h4>
                      <span className="model-size">{model.size}</span>
                    </div>
                    <p className="model-description">{model.description}</p>
                    <button 
                      className="btn-secondary"
                      onClick={() => createModel(model.name)}
                      disabled={isCreating}
                    >
                      Install Model
                    </button>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {activeTab === 'installed' && (
          <div className="tab-content">
            {installedModels.length === 0 ? (
              <div className="empty-state">
                <div className="empty-icon">📦</div>
                <h3>No Models Installed</h3>
                <p>Install a model from the Available Models tab to get started</p>
                <button 
                  className="btn-primary"
                  onClick={() => setActiveTab('available')}
                >
                  Browse Available Models
                </button>
              </div>
            ) : (
              <div className="models-grid">
                {installedModels.map((model, index) => (
                  <div key={index} className="model-card installed">
                    <div className="model-header">
                      <h4>{model.name}</h4>
                      <div className="model-status">
                        <span className={`status-indicator ${model.is_running ? 'running' : 'stopped'}`}></span>
                        {model.is_running ? 'Running' : 'Stopped'}
                      </div>
                    </div>
                    <p className="model-info">Container: {model.container_name}</p>
                    <div className="model-actions">
                      <button 
                        className="btn-primary chat-button"
                        onClick={() => startChatWithModel(model.name)}
                        disabled={!model.is_running}
                      >
                        {model.is_running ? '💬 Chat Now' : '⏸️ Model Stopped'}
                      </button>
                      <button 
                        className="btn-danger"
                        onClick={() => deleteModel(model.name)}
                      >
                        🗑️ Delete
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

export default ModelManagement;
