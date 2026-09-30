import React, { useState } from 'react';
import '../styles/LoginPage.css';
import { API_BASE_URL } from '../config';

/**
 * LoginPage component handles user authentication (login and registration)
 * This component provides forms for both login and registration functionality
 */
function LoginPage({ onLogin }) {
  // Form state management
  const [isLoginMode, setIsLoginMode] = useState(true);
  const [formData, setFormData] = useState({
    username: '',
    password: ''
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // Handle input changes
  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
    // Clear error when user starts typing
    if (error) setError('');
  };

  // Handle form submission
  const handleSubmit = async (e) => {
    e.preventDefault();
    
    // Basic validation
    if (!formData.username.trim() || !formData.password.trim()) {
      setError('Please fill in all fields');
      return;
    }

    if (formData.username.length < 3) {
      setError('Username must be at least 3 characters long');
      return;
    }

    if (formData.password.length < 6) {
      setError('Password must be at least 6 characters long');
      return;
    }

    setLoading(true);
    setError('');

    try {
      // Determine endpoint based on mode
      const endpoint = isLoginMode
        ? `${API_BASE_URL}/api/auth/login`
        : `${API_BASE_URL}/api/auth/register`;
      
      // Make API request
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          username: formData.username.trim(),
          password: formData.password
        })
      });

      const data = await response.json();

      if (response.ok && data.success) {
        // Authentication successful
        onLogin(data.user, data.token);
      } else {
        // Authentication failed
        setError(data.message || 'Authentication failed');
      }
    } catch (error) {
      console.error('Authentication error:', error);
      setError('Network error. Please check your connection and try again.');
    } finally {
      setLoading(false);
    }
  };

  // Toggle between login and registration modes
  const toggleMode = () => {
    setIsLoginMode(!isLoginMode);
    setError('');
    setFormData({ username: '', password: '' });
  };

  return (
    <div className="login-page">
      <div className="login-container">
        {/* Header */}
        <div className="login-header">
          <h1>GoChat</h1>
          <p>Real-time Multi-room Chat Application</p>
        </div>

        {/* Authentication Form */}
        <div className="login-form-container">
          <div className="form-tabs">
            <button 
              className={isLoginMode ? 'tab active' : 'tab'}
              onClick={() => setIsLoginMode(true)}
              type="button"
            >
              Login
            </button>
            <button 
              className={!isLoginMode ? 'tab active' : 'tab'}
              onClick={() => setIsLoginMode(false)}
              type="button"
            >
              Register
            </button>
          </div>

          <form onSubmit={handleSubmit} className="login-form">
            {/* Error Message */}
            {error && (
              <div className="error-message">
                {error}
              </div>
            )}

            {/* Username Field */}
            <div className="form-group">
              <label htmlFor="username">Username</label>
              <input
                type="text"
                id="username"
                name="username"
                value={formData.username}
                onChange={handleInputChange}
                placeholder="Enter your username"
                disabled={loading}
                autoComplete="username"
              />
            </div>

            {/* Password Field */}
            <div className="form-group">
              <label htmlFor="password">Password</label>
              <input
                type="password"
                id="password"
                name="password"
                value={formData.password}
                onChange={handleInputChange}
                placeholder="Enter your password"
                disabled={loading}
                autoComplete={isLoginMode ? "current-password" : "new-password"}
              />
            </div>

            {/* Submit Button */}
            <button 
              type="submit" 
              className="submit-button"
              disabled={loading}
            >
              {loading ? (
                <span className="loading-spinner-small"></span>
              ) : (
                isLoginMode ? 'Login' : 'Register'
              )}
            </button>
          </form>

          {/* Mode Toggle */}
          <div className="mode-toggle">
            <p>
              {isLoginMode ? "Don't have an account? " : "Already have an account? "}
              <button 
                type="button"
                className="toggle-button"
                onClick={toggleMode}
                disabled={loading}
              >
                {isLoginMode ? 'Register here' : 'Login here'}
              </button>
            </p>
          </div>
        </div>

        {/* Features Information */}
        <div className="features-info">
          <h3>Features</h3>
          <ul>
            <li>Real-time messaging with WebSocket technology</li>
            <li>Multiple chat rooms support</li>
            <li>User authentication and session management</li>
            <li>Live user presence indicators</li>
            <li>Message history for joined rooms</li>
          </ul>
        </div>
      </div>
    </div>
  );
}

export default LoginPage;