import React from 'react';
import ReactDOM from 'react-dom/client';
import './styles/index.css';
import App from './App';

/**
 * Entry point for the React application
 * This file renders the main App component into the DOM
 */

// Get the root element from the HTML
const root = ReactDOM.createRoot(document.getElementById('root'));

// Render the App component
root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);