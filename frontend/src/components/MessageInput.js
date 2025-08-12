import React, { useState, useRef, useEffect } from 'react';

/**
 * MessageInput component provides an input field for sending chat messages
 * This component handles message composition, submission, and keyboard shortcuts
 */
function MessageInput({ onSendMessage, disabled, placeholder = 'Type your message...' }) {
  const [message, setMessage] = useState('');
  const inputRef = useRef(null);

  // Focus input when component mounts or becomes enabled
  useEffect(() => {
    if (!disabled && inputRef.current) {
      inputRef.current.focus();
    }
  }, [disabled]);

  // Handle input changes
  const handleInputChange = (e) => {
    setMessage(e.target.value);
  };

  // Handle form submission
  const handleSubmit = (e) => {
    e.preventDefault();
    sendMessage();
  };

  // Handle keyboard events
  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      // Send message on Enter (without Shift)
      e.preventDefault();
      sendMessage();
    }
    // Allow Shift+Enter for line breaks (handled by textarea)
  };

  // Send message function
  const sendMessage = () => {
    const trimmedMessage = message.trim();
    
    if (!trimmedMessage || disabled) {
      return;
    }

    // Validate message length
    if (trimmedMessage.length > 1000) {
      alert('Message is too long. Please keep it under 1000 characters.');
      return;
    }

    // Send the message
    onSendMessage(trimmedMessage);
    
    // Clear the input
    setMessage('');
    
    // Refocus the input
    if (inputRef.current) {
      inputRef.current.focus();
    }
  };

  return (
    <form onSubmit={handleSubmit} className="message-input-form">
      <div className="input-container">
        {/* Message input field */}
        <textarea
          ref={inputRef}
          value={message}
          onChange={handleInputChange}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          disabled={disabled}
          className="message-input"
          rows={1}
          maxLength={1000}
          autoComplete="off"
          autoCorrect="off"
          autoCapitalize="off"
          spellCheck="true"
        />
        
        {/* Send button */}
        <button
          type="submit"
          disabled={disabled || !message.trim()}
          className="send-button"
          title="Send message (Enter)"
        >
          <svg 
            width="20" 
            height="20" 
            viewBox="0 0 24 24" 
            fill="none" 
            stroke="currentColor" 
            strokeWidth="2" 
            strokeLinecap="round" 
            strokeLinejoin="round"
          >
            <line x1="22" y1="2" x2="11" y2="13"></line>
            <polygon points="22,2 15,22 11,13 2,9"></polygon>
          </svg>
        </button>
      </div>
      
      {/* Character counter */}
      {message.length > 800 && (
        <div className="character-counter">
          {message.length}/1000
        </div>
      )}
      
      {/* Keyboard shortcut hint */}
      <div className="input-hint">
        Press Enter to send • Shift+Enter for new line
      </div>
    </form>
  );
}

export default MessageInput;