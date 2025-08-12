import React, { useState } from 'react';

/**
 * RoomSelector component allows users to join different chat rooms
 * This component provides both predefined rooms and custom room creation
 */
function RoomSelector({ currentRoom, onRoomSelect, isJoining }) {
  const [newRoomName, setNewRoomName] = useState('');
  const [showCustomRoom, setShowCustomRoom] = useState(false);

  // Predefined popular rooms
  const predefinedRooms = [
    { name: 'general', description: 'General discussion' },
    { name: 'random', description: 'Random chats' },
    { name: 'tech', description: 'Technology talk' },
    { name: 'gaming', description: 'Gaming discussions' },
    { name: 'music', description: 'Music and entertainment' }
  ];

  // Handle predefined room selection
  const handleRoomClick = (roomName) => {
    if (roomName !== currentRoom && !isJoining) {
      onRoomSelect(roomName);
    }
  };

  // Handle custom room creation
  const handleCustomRoomSubmit = (e) => {
    e.preventDefault();
    
    const trimmedName = newRoomName.trim().toLowerCase();
    
    // Validation
    if (!trimmedName) {
      return;
    }
    
    if (trimmedName.length < 2) {
      alert('Room name must be at least 2 characters long');
      return;
    }
    
    if (trimmedName.length > 30) {
      alert('Room name must be less than 30 characters');
      return;
    }
    
    // Check for valid characters (alphanumeric, hyphens, underscores)
    if (!/^[a-z0-9_-]+$/.test(trimmedName)) {
      alert('Room name can only contain letters, numbers, hyphens, and underscores');
      return;
    }
    
    // Check if room already exists in predefined list
    if (predefinedRooms.some(room => room.name === trimmedName)) {
      alert('This room already exists in the predefined list');
      return;
    }
    
    // Join the custom room
    onRoomSelect(trimmedName);
    setNewRoomName('');
    setShowCustomRoom(false);
  };

  // Handle leaving current room
  const handleLeaveRoom = () => {
    if (!isJoining) {
      onRoomSelect('');
    }
  };

  return (
    <div className="room-selector">
      <div className="room-selector-header">
        <h3>Chat Rooms</h3>
        {currentRoom && (
          <button
            onClick={handleLeaveRoom}
            disabled={isJoining}
            className="leave-room-button"
            title="Leave current room"
          >
            ×
          </button>
        )}
      </div>

      {/* Current room indicator */}
      {currentRoom && (
        <div className="current-room">
          <div className="current-room-indicator">
            <span className="room-icon">#</span>
            <span className="room-name">{currentRoom}</span>
            {isJoining && <span className="joining-indicator">Joining...</span>}
          </div>
        </div>
      )}

      {/* Predefined rooms */}
      <div className="predefined-rooms">
        <h4>Popular Rooms</h4>
        <div className="room-list">
          {predefinedRooms.map((room) => (
            <button
              key={room.name}
              onClick={() => handleRoomClick(room.name)}
              disabled={isJoining || room.name === currentRoom}
              className={`room-item ${room.name === currentRoom ? 'active' : ''}`}
            >
              <div className="room-info">
                <span className="room-name">#{room.name}</span>
                <span className="room-description">{room.description}</span>
              </div>
              {room.name === currentRoom && (
                <span className="active-indicator">●</span>
              )}
            </button>
          ))}
        </div>
      </div>

      {/* Custom room section */}
      <div className="custom-room-section">
        <button
          onClick={() => setShowCustomRoom(!showCustomRoom)}
          className="toggle-custom-room"
          disabled={isJoining}
        >
          {showCustomRoom ? 'Hide' : 'Join Custom Room'}
        </button>

        {showCustomRoom && (
          <form onSubmit={handleCustomRoomSubmit} className="custom-room-form">
            <div className="input-group">
              <input
                type="text"
                value={newRoomName}
                onChange={(e) => setNewRoomName(e.target.value)}
                placeholder="Enter room name"
                disabled={isJoining}
                className="room-name-input"
                maxLength={30}
                autoComplete="off"
              />
              <button
                type="submit"
                disabled={!newRoomName.trim() || isJoining}
                className="join-button"
              >
                Join
              </button>
            </div>
            <div className="input-hint">
              Use letters, numbers, hyphens, and underscores only
            </div>
          </form>
        )}
      </div>

      {/* Loading indicator */}
      {isJoining && (
        <div className="joining-status">
          <div className="loading-spinner-small"></div>
          <span>Joining room...</span>
        </div>
      )}
    </div>
  );
}

export default RoomSelector;