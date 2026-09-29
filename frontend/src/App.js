import React, { useState, useEffect } from 'react';
import './styles/App.css';
import LoginPage from './pages/LoginPage';
import ChatPage from './pages/ChatPage';
import { API_BASE_URL } from './config';

/**
 * Main application component that manages authentication state and routing
 * This component handles the overall application flow between login and chat
 */
function App() {
  // Authentication state management
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [user, setUser] = useState(null);
  const [token, setToken] = useState(null);
  const [loading, setLoading] = useState(true);

  // Check for existing authentication on app startup
  useEffect(() => {
    const checkExistingAuth = async () => {
      try {
        // Check if we have a stored token
        const storedToken = localStorage.getItem('gochat_token');
        const storedUser = localStorage.getItem('gochat_user');

        if (storedToken && storedUser) {
          // Validate the stored token with the server
          const response = await fetch(`${API_BASE_URL}/api/auth/validate`, {
            method: 'POST',
            headers: {
              'Authorization': `Bearer ${storedToken}`,
              'Content-Type': 'application/json'
            }
          });

          if (response.ok) {
            // Token is valid, restore authentication state
            const userData = JSON.parse(storedUser);
            setToken(storedToken);
            setUser(userData);
            setIsAuthenticated(true);
          } else {
            // Token is invalid, clear stored data
            localStorage.removeItem('gochat_token');
            localStorage.removeItem('gochat_user');
          }
        }
      } catch (error) {
        console.error('Error checking authentication:', error);
        // Clear potentially corrupted data
        localStorage.removeItem('gochat_token');
        localStorage.removeItem('gochat_user');
      } finally {
        setLoading(false);
      }
    };

    checkExistingAuth();
  }, []);

  // Handle successful login
  const handleLogin = (userData, authToken) => {
    // Store authentication data
    setUser(userData);
    setToken(authToken);
    setIsAuthenticated(true);

    // Persist to localStorage for session management
    localStorage.setItem('gochat_token', authToken);
    localStorage.setItem('gochat_user', JSON.stringify(userData));
  };

  // Handle logout
  const handleLogout = () => {
    // Clear authentication state
    setUser(null);
    setToken(null);
    setIsAuthenticated(false);

    // Clear stored data
    localStorage.removeItem('gochat_token');
    localStorage.removeItem('gochat_user');
  };

  // Show loading spinner while checking authentication
  if (loading) {
    return (
      <div className="app">
        <div className="loading-container">
          <div className="loading-spinner"></div>
          <p>Loading GoChat...</p>
        </div>
      </div>
    );
  }

  // Render appropriate page based on authentication status
  return (
    <div className="app">
      {isAuthenticated ? (
        <ChatPage 
          user={user} 
          token={token} 
          onLogout={handleLogout} 
        />
      ) : (
        <LoginPage onLogin={handleLogin} />
      )}
    </div>
  );
}

export default App;