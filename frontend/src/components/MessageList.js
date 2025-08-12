import React, { useEffect, useRef } from 'react';

/**
 * MessageList component displays a list of chat messages
 * This component handles message rendering and auto-scrolling to latest messages
 */
function MessageList({ messages, currentUser }) {
  const messagesEndRef = useRef(null);
  const containerRef = useRef(null);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  // Format timestamp for display
  const formatTimestamp = (timestamp) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffInHours = (now - date) / (1000 * 60 * 60);

    if (diffInHours < 24) {
      // Show time for messages from today
      return date.toLocaleTimeString([], { 
        hour: '2-digit', 
        minute: '2-digit' 
      });
    } else {
      // Show date for older messages
      return date.toLocaleDateString([], { 
        month: 'short', 
        day: 'numeric',
        hour: '2-digit', 
        minute: '2-digit' 
      });
    }
  };

  // Group consecutive messages from the same user
  const groupMessages = (messages) => {
    const grouped = [];
    let currentGroup = null;

    messages.forEach((message, index) => {
      const prevMessage = messages[index - 1];
      const isSameUser = prevMessage && prevMessage.sender_username === message.sender_username;
      const timeDiff = prevMessage ? new Date(message.timestamp) - new Date(prevMessage.timestamp) : 0;
      const isWithinGroupTime = timeDiff < 5 * 60 * 1000; // 5 minutes

      if (isSameUser && isWithinGroupTime) {
        // Add to current group
        currentGroup.messages.push(message);
      } else {
        // Start new group
        currentGroup = {
          sender_username: message.sender_username,
          timestamp: message.timestamp,
          messages: [message],
          isCurrentUser: message.sender_username === currentUser
        };
        grouped.push(currentGroup);
      }
    });

    return grouped;
  };

  const messageGroups = groupMessages(messages);

  if (messages.length === 0) {
    return (
      <div className="message-list empty">
        <div className="empty-state">
          <p>No messages yet. Start the conversation!</p>
        </div>
      </div>
    );
  }

  return (
    <div className="message-list" ref={containerRef}>
      {messageGroups.map((group, groupIndex) => (
        <div 
          key={groupIndex} 
          className={`message-group ${group.isCurrentUser ? 'own-message' : 'other-message'}`}
        >
          {/* User avatar and name (only for other users) */}
          {!group.isCurrentUser && (
            <div className="message-header">
              <div className="user-avatar">
                {group.sender_username.charAt(0).toUpperCase()}
              </div>
              <div className="message-meta">
                <span className="username">{group.sender_username}</span>
                <span className="timestamp">{formatTimestamp(group.timestamp)}</span>
              </div>
            </div>
          )}

          {/* Messages in group */}
          <div className="message-content">
            {group.messages.map((message, messageIndex) => (
              <div 
                key={message.id || `${groupIndex}-${messageIndex}`} 
                className="message-bubble"
              >
                <span className="message-text">{message.content}</span>
                {/* Show timestamp for own messages */}
                {group.isCurrentUser && (
                  <span className="message-time">
                    {formatTimestamp(message.timestamp)}
                  </span>
                )}
              </div>
            ))}
          </div>
        </div>
      ))}
      
      {/* Invisible element to scroll to */}
      <div ref={messagesEndRef} />
    </div>
  );
}

export default MessageList;