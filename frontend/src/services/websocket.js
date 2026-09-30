/**
 * WebSocket service for real-time communication with the GoChat server
 * This service manages the WebSocket connection, message sending, and event handling
 */

import { API_BASE_URL } from '../config';

class WebSocketService {
  constructor() {
    this.ws = null;
    this.url = null;
    this.token = null;
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 5;
    this.reconnectInterval = 3000;
    this.isConnecting = false;
    this.currentRoom = null;
    
    // Event listeners
    this.eventListeners = {
      open: [],
      close: [],
      error: [],
      message: [],
      userJoined: [],
      userLeft: [],
      newMessage: [],
      userList: [],
      joinedRoom: [],
      leftRoom: [],
      roomHistory: [],
      ping: []
    };
  }

  /**
   * Connect to the WebSocket server with authentication token
   * @param {string} token - JWT authentication token
   */
  connect(token) {
    if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
      return Promise.resolve();
    }

    this.token = token;
    this.url = this.getWebSocketUrl();
    this.isConnecting = true;

    return new Promise((resolve, reject) => {
      try {
        console.log('Connecting to WebSocket:', this.url);
        this.ws = new WebSocket(this.url);

        // Connection opened
        this.ws.onopen = (event) => {
          console.log('WebSocket connected successfully');
          this.isConnecting = false;
          this.reconnectAttempts = 0;
          this.emit('open', event);
          resolve();
        };

        // Message received
        this.ws.onmessage = (event) => {
          try {
            const message = JSON.parse(event.data);
            this.handleMessage(message);
          } catch (error) {
            console.error('Error parsing WebSocket message:', error);
          }
        };

        // Connection closed
        this.ws.onclose = (event) => {
          console.log('WebSocket connection closed:', event.code, event.reason);
          this.isConnecting = false;
          this.emit('close', event);
          
          // Attempt to reconnect if not manually closed
          if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnectAttempts) {
            this.scheduleReconnect();
          }
        };

        // Connection error
        this.ws.onerror = (event) => {
          console.error('WebSocket error:', event);
          this.isConnecting = false;
          this.emit('error', event);
          reject(new Error('WebSocket connection failed'));
        };

      } catch (error) {
        this.isConnecting = false;
        reject(error);
      }
    });
  }

  /**
   * Disconnect from the WebSocket server
   */
  disconnect() {
    if (this.ws) {
      this.ws.close(1000, 'Client disconnecting');
      this.ws = null;
    }
    this.reconnectAttempts = this.maxReconnectAttempts; // Prevent reconnection
  }

  /**
   * Schedule a reconnection attempt
   */
  scheduleReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts}) in ${this.reconnectInterval}ms`);

    setTimeout(() => {
      if (this.token) {
        this.connect(this.token).catch(error => {
          console.error('Reconnection failed:', error);
        });
      }
    }, this.reconnectInterval);
  }

  /**
   * Handle incoming messages from the server
   * @param {Object} message - Parsed WebSocket message
   */
  handleMessage(message) {
    console.log('Received message:', message.type, message);

    switch (message.type) {
      case 'NEW_MESSAGE':
        this.emit('newMessage', message.payload);
        break;
      case 'USER_JOINED':
        this.emit('userJoined', message.payload);
        break;
      case 'USER_LEFT':
        this.emit('userLeft', message.payload);
        break;
      case 'USER_LIST':
        this.emit('userList', message.payload);
        break;
      case 'JOINED_ROOM':
        this.currentRoom = message.payload.room_name;
        this.emit('joinedRoom', message.payload);
        break;
      case 'LEFT_ROOM':
        this.currentRoom = null;
        this.emit('leftRoom', message.payload);
        break;
      case 'ROOM_HISTORY':
        this.emit('roomHistory', message.payload);
        break;
      case 'ERROR':
        console.error('Server error:', message.payload);
        this.emit('error', message.payload);
        break;
      case 'PONG':
        this.emit('ping', message.payload);
        break;
      default:
        console.log('Unknown message type:', message.type);
    }

    // Emit general message event
    this.emit('message', message);
  }

  /**
   * Send a message to the server
   * @param {string} type - Message type
   * @param {Object} payload - Message payload
   */
  send(type, payload) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('WebSocket is not connected');
      return false;
    }

    const message = {
      type: type,
      payload: payload,
      timestamp: new Date().toISOString()
    };

    try {
      this.ws.send(JSON.stringify(message));
      console.log('Sent message:', type, payload);
      return true;
    } catch (error) {
      console.error('Error sending message:', error);
      return false;
    }
  }

  /**
   * Join a chat room
   * @param {string} roomName - Name of the room to join
   * @param {string} username - Username of the current user
   */
  joinRoom(roomName, username) {
    return this.send('JOIN', {
      room_name: roomName,
      username: username
    });
  }

  /**
   * Leave the current room
   * @param {string} roomName - Name of the room to leave
   * @param {string} username - Username of the current user
   */
  leaveRoom(roomName, username) {
    return this.send('LEAVE', {
      room_name: roomName,
      username: username
    });
  }

  /**
   * Send a chat message
   * @param {string} roomName - Name of the room
   * @param {string} username - Username of the sender
   * @param {string} content - Message content
   */
  sendMessage(roomName, username, content) {
    return this.send('MESSAGE', {
      room_name: roomName,
      username: username,
      content: content
    });
  }

  /**
   * Send a ping message for connection health check
   */
  ping() {
    return this.send('PING', {
      client_timestamp: new Date().toISOString()
    });
  }

  /**
   * Add an event listener
   * @param {string} event - Event name
   * @param {Function} callback - Callback function
   */
  on(event, callback) {
    if (this.eventListeners[event]) {
      this.eventListeners[event].push(callback);
    }
  }

  /**
   * Remove an event listener
   * @param {string} event - Event name
   * @param {Function} callback - Callback function to remove
   */
  off(event, callback) {
    if (this.eventListeners[event]) {
      const index = this.eventListeners[event].indexOf(callback);
      if (index > -1) {
        this.eventListeners[event].splice(index, 1);
      }
    }
  }

  /**
   * Clear all event listeners
   */
  clearAllListeners() {
    this.eventListeners = {
      open: [],
      close: [],
      error: [],
      message: [],
      userJoined: [],
      userLeft: [],
      newMessage: [],
      userList: [],
      joinedRoom: [],
      leftRoom: [],
      roomHistory: [],
      ping: []
    };
  }

  /**
   * Emit an event to all listeners
   * @param {string} event - Event name
   * @param {*} data - Event data
   */
  emit(event, data) {
    if (this.eventListeners[event]) {
      this.eventListeners[event].forEach(callback => {
        try {
          callback(data);
        } catch (error) {
          console.error('Error in event listener:', error);
        }
      });
    }
  }

  /**
   * Get the WebSocket URL with authentication token
   * @returns {string} WebSocket URL
   */
  getWebSocketUrl() {
    // Derive the WebSocket URL from the configured API base URL, including
    // the protocol: an https backend needs wss even if the page itself was
    // served over http, and vice versa.
    const url = new URL(API_BASE_URL);
    url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
    url.pathname = `${url.pathname.replace(/\/$/, '')}/ws`;
    url.search = `?token=${encodeURIComponent(this.token)}`;
    return url.toString();
  }

  /**
   * Check if the WebSocket is connected
   * @returns {boolean} Connection status
   */
  isConnected() {
    return this.ws && this.ws.readyState === WebSocket.OPEN;
  }

  /**
   * Get the current room name
   * @returns {string|null} Current room name
   */
  getCurrentRoom() {
    return this.currentRoom;
  }
}

// Create and export a singleton instance
const websocketService = new WebSocketService();
export default websocketService;