import React, { useState, useEffect, useRef } from 'react';
import websocketService from '../services/websocket';
import MessageList from '../components/MessageList';
import MessageInput from '../components/MessageInput';
import UserList from '../components/UserList';
import RoomSelector from '../components/RoomSelector';
import '../styles/ChatPage.css';

/**
 * ChatPage component manages the main chat interface
 * This component handles room selection, message display, and user interactions
 */
function ChatPage({ user, token, onLogout }) {
  // Chat state management
  const [currentRoom, setCurrentRoom] = useState('');
  const [messages, setMessages] = useState([]);
  const [users, setUsers] = useState([]);
  const [connectionStatus, setConnectionStatus] = useState('disconnected');
  const [error, setError] = useState('');
  const [isJoiningRoom, setIsJoiningRoom] = useState(false);

  // Refs for managing component state
  const wsConnectedRef = useRef(false);
  const currentRoomRef = useRef('');
  // Room we've asked to join but the server hasn't confirmed yet. The server
  // sends the user list and history before JOINED_ROOM, so those must be
  // accepted for this room too
  const joiningRoomRef = useRef('');
  
  // Cache for storing messages per room
  const roomMessagesCache = useRef({});
  const roomUsersCache = useRef({});

  // Update refs when state changes
  useEffect(() => {
    currentRoomRef.current = currentRoom;
  }, [currentRoom]);

  // Initialize WebSocket connection
  useEffect(() => {
    const initializeWebSocket = async () => {
      try {
        setConnectionStatus('connecting');
        
        // Clear any existing listeners to prevent duplicates
        websocketService.clearAllListeners();
        
        // Set up event listeners before connecting
        setupWebSocketListeners();
        
        // Connect to WebSocket
        await websocketService.connect(token);
        wsConnectedRef.current = true;
        setConnectionStatus('connected');
        setError('');
        
      } catch (error) {
        console.error('Failed to connect to WebSocket:', error);
        setConnectionStatus('disconnected');
        setError('Failed to connect to chat server. Please refresh the page.');
      }
    };

    initializeWebSocket();

    // Cleanup on component unmount
    return () => {
      if (wsConnectedRef.current) {
        websocketService.disconnect();
        wsConnectedRef.current = false;
      }
      websocketService.clearAllListeners();
    };
  }, [token]);

  // Whether an event for this room should update the visible chat
  const isActiveRoom = (roomName) =>
    roomName === currentRoomRef.current || roomName === joiningRoomRef.current;

  // Set up WebSocket event listeners
  const setupWebSocketListeners = () => {
    // Connection events
    websocketService.on('open', () => {
      setConnectionStatus('connected');
      setError('');
    });

    websocketService.on('close', () => {
      setConnectionStatus('disconnected');
      wsConnectedRef.current = false;
    });

    websocketService.on('error', (error) => {
      console.error('WebSocket error:', error);
      setConnectionStatus('error');
      setError('Connection error occurred');
    });

    // Chat events
    websocketService.on('joinedRoom', (data) => {
      console.log('Joined room:', data);
      currentRoomRef.current = data.room_name;
      joiningRoomRef.current = '';
      setCurrentRoom(data.room_name);
      setIsJoiningRoom(false);
      setError('');
    });

    websocketService.on('leftRoom', (data) => {
      console.log('Left room:', data);
      currentRoomRef.current = '';
      setCurrentRoom('');
      // setMessages([]);
      setUsers([]);
    });

    websocketService.on('newMessage', (data) => {
      // Add message to the room's cache
      if (!roomMessagesCache.current[data.room_name]) {
        roomMessagesCache.current[data.room_name] = [];
      }
      roomMessagesCache.current[data.room_name].push(data.message);
      
      // Only update UI if it's for the current room
      if (isActiveRoom(data.room_name)) {
        setMessages(prev => [...prev, data.message]);
      }
    });

    websocketService.on('userJoined', (data) => {
      if (data.room_name === currentRoomRef.current) {
        console.log('User joined:', data.username);
        // User list will be updated via USER_LIST message
      }
    });

    websocketService.on('userLeft', (data) => {
      if (data.room_name === currentRoomRef.current) {
        console.log('User left:', data.username);
        // User list will be updated via USER_LIST message
      }
    });

    websocketService.on('userList', (data) => {
      // Cache the user list for this room
      roomUsersCache.current[data.room_name] = data.users || [];
      
      // Only update UI if it's for the current room
      if (isActiveRoom(data.room_name)) {
        setUsers(data.users || []);
      }
    });

    websocketService.on('roomHistory', (data) => {
      // Cache the room history
      roomMessagesCache.current[data.room_name] = data.messages || [];
      
      // Only update UI if it's for the current room
      if (isActiveRoom(data.room_name)) {
        setMessages(data.messages || []);
      }
    });
  };

  // Handle room selection
  const handleRoomSelect = async (roomName) => {
    if (roomName === currentRoom || isJoiningRoom) {
      return;
    }

    setIsJoiningRoom(true);
    setError('');

    try {
      // Leave current room if in one
      if (currentRoom) {
        // Save current room state to cache before leaving
        roomMessagesCache.current[currentRoom] = messages;
        roomUsersCache.current[currentRoom] = users;
        
        websocketService.leaveRoom(currentRoom, user.username);
        currentRoomRef.current = '';
        setCurrentRoom('');
        setMessages([]);
        setUsers([]);
      }

      // Join new room
      if (roomName) {
        // Check if we have cached data for this room
        if (roomMessagesCache.current[roomName]) {
          setMessages(roomMessagesCache.current[roomName]);
        }
        if (roomUsersCache.current[roomName]) {
          setUsers(roomUsersCache.current[roomName]);
        }
        
        joiningRoomRef.current = roomName;
        const success = websocketService.joinRoom(roomName, user.username);
        if (!success) {
          joiningRoomRef.current = '';
          throw new Error('Failed to send join request');
        }

        // Set a timeout to prevent infinite loading
        setTimeout(() => {
          if (isJoiningRoom) {
            setIsJoiningRoom(false);
            setError('Room join timed out. Please check your connection and try again.');
          }
        }, 10000); // 10 second timeout

        // Room state will be updated via WebSocket events (which may override cached data with fresh data)
      } else {
        setIsJoiningRoom(false);
      }
    } catch (error) {
      console.error('Error changing rooms:', error);
      setError('Failed to join room. Please try again.');
      setIsJoiningRoom(false);
    }
  };

  // Handle sending messages
  const handleSendMessage = (content) => {
    if (!currentRoom || !content.trim()) {
      return;
    }

    const success = websocketService.sendMessage(currentRoom, user.username, content.trim());
    if (!success) {
      setError('Failed to send message. Please check your connection.');
    }
  };

  // Handle logout
  const handleLogout = () => {
    // Leave current room before logging out
    if (currentRoom) {
      websocketService.leaveRoom(currentRoom, user.username);
    }
    
    // Disconnect WebSocket
    websocketService.disconnect();
    
    // Call parent logout handler
    onLogout();
  };

  return (
    <div className="chat-page">
      {/* Header */}
      <header className="chat-header">
        <div className="header-left">
          <h1>GoChat</h1>
          <div className="connection-status">
            <span className={`status-indicator ${connectionStatus}`}></span>
            <span className="status-text">
              {connectionStatus === 'connected' && 'Connected'}
              {connectionStatus === 'connecting' && 'Connecting...'}
              {connectionStatus === 'disconnected' && 'Disconnected'}
              {connectionStatus === 'error' && 'Connection Error'}
            </span>
          </div>
        </div>
        
        <div className="header-right">
          <span className="username">Welcome, {user.username}</span>
          <button onClick={handleLogout} className="logout-button">
            Logout
          </button>
        </div>
      </header>

      {/* Error Message */}
      {error && (
        <div className="error-banner">
          <span>{error}</span>
          <button onClick={() => setError('')} className="close-error">×</button>
        </div>
      )}

      {/* Main Chat Interface */}
      <div className="chat-container">
        {/* Sidebar */}
        <aside className="chat-sidebar">
          <RoomSelector
            currentRoom={currentRoom}
            onRoomSelect={handleRoomSelect}
            isJoining={isJoiningRoom}
          />
          
          {currentRoom && (
            <UserList
              users={users}
              currentUser={user.username}
              roomName={currentRoom}
            />
          )}
        </aside>

        {/* Main Chat Area */}
        <main className="chat-main">
          {currentRoom ? (
            <>
              {/* Room Header */}
              <div className="room-header">
                <h2>#{currentRoom}</h2>
                <span className="user-count">
                  {users.length} user{users.length !== 1 ? 's' : ''} online
                </span>
              </div>

              {/* Messages Area */}
              <div className="messages-container">
                <MessageList
                  messages={messages}
                  currentUser={user.username}
                />
              </div>

              {/* Message Input */}
              <div className="message-input-container">
                <MessageInput
                  onSendMessage={handleSendMessage}
                  disabled={connectionStatus !== 'connected'}
                  placeholder={
                    connectionStatus !== 'connected'
                      ? 'Connecting...'
                      : `Message #${currentRoom}`
                  }
                />
              </div>
            </>
          ) : (
            /* Welcome Screen */
            <div className="welcome-screen">
              <div className="welcome-content">
                <h2>Welcome to GoChat!</h2>
                <p>Select a room from the sidebar to start chatting.</p>
                <div className="welcome-features">
                  <h3>Features:</h3>
                  <ul>
                    <li>Real-time messaging</li>
                    <li>Multiple chat rooms</li>
                    <li>User presence indicators</li>
                    <li>Message history</li>
                  </ul>
                </div>
              </div>
            </div>
          )}
        </main>
      </div>
    </div>
  );
}

export default ChatPage;