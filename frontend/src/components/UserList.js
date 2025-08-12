import React from 'react';

/**
 * UserList component displays the list of users currently in a chat room
 * This component shows user presence and highlights the current user
 */
function UserList({ users, currentUser, roomName }) {
  // Sort users alphabetically, with current user at the top
  const sortedUsers = [...users].sort((a, b) => {
    if (a === currentUser) return -1;
    if (b === currentUser) return 1;
    return a.localeCompare(b);
  });

  return (
    <div className="user-list">
      <div className="user-list-header">
        <h3>Users in #{roomName}</h3>
        <span className="user-count">
          {users.length} online
        </span>
      </div>
      
      {users.length === 0 ? (
        <div className="user-list-empty">
          <p>No users online</p>
        </div>
      ) : (
        <div className="user-list-content">
          {sortedUsers.map((username) => (
            <div 
              key={username} 
              className={`user-item ${username === currentUser ? 'current-user' : ''}`}
            >
              {/* User avatar */}
              <div className="user-avatar">
                {username.charAt(0).toUpperCase()}
              </div>
              
              {/* User info */}
              <div className="user-info">
                <span className="username">
                  {username}
                  {username === currentUser && <span className="you-label"> (you)</span>}
                </span>
                <div className="user-status">
                  <span className="status-indicator online"></span>
                  <span className="status-text">Online</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default UserList;