import React, { useEffect, useState } from "react";
import axios from "axios";
import { useLocation, Link } from "react-router-dom";

/**
 * UserList component displays a list of users available for chat.
 * It fetches the list of users based on the provided username,
 * allows selecting a user to start a chat, and handles user logout.
 */
const UserList: React.FC = () => {
  // State variables for users list, selected user, and current time
  const [users, setUsers] = useState<string[]>([]);
  const [selectedUser, setSelectedUser] = useState<string | null>(null);
  const [currentTime, setCurrentTime] = useState(new Date());

  // Retrieves the username from the URL query parameters
  const location = useLocation();
  const username = new URLSearchParams(location.search).get("username");

  // Effect hook to fetch users and update the list periodically
  useEffect(() => {
    const fetchUsers = async () => {
      try {
        const response = await axios.get(
          `http://127.0.0.1:8080/users?username=${username}`
        );
        // Sets the users state with the fetched user list
        setUsers(response.data.users || []);
      } catch (error) {
        console.error("Error fetching users:", error);
      }
    };

    // Checks if username is available, fetches users, and sets an interval for periodic fetching
    if (username) {
      fetchUsers();
      const interval = setInterval(() => {
        fetchUsers();
      }, 1000);

      // Cleans up the interval on component unmount
      return () => clearInterval(interval);
    }
  }, [username]);

  // Handles user click event to select a user and redirect to chat with the selected user
  const handleUserClick = (user: string) => {
    setSelectedUser(user);
    window.location.href = `/chat/${user}?currentUser=${username}`;
  };

  // Handles logout event by removing username from local storage and redirecting to homepage
  const handleLogout = () => {
    localStorage.removeItem("username");
    window.location.href = "/";
  };

  // Renders the user list with selectable users and logout button
  return (
    <div className="container">
      <div className="d-flex justify-content-between align-items-center mb-3">
        <h2 className="mt-4">Simple Chat Room</h2>
        <button onClick={handleLogout} className="btn btn-danger logout-button">
          Logout
        </button>
      </div>
      <h5 className="mt-4 text-center">Select the person to chat with</h5>
      <ul className="list-group">
        {users.map((user) => (
          <li
            key={user}
            className="list-group-item d-flex justify-content-center align-items-center"
            style={{ minHeight: "60px" }}
          >
            <button
              className={`btn btn-block ${
                selectedUser === user ? "btn-dark" : "btn-primary"
              }`}
              onClick={() => handleUserClick(user)}
              style={{ width: "100%" }}
            >
              {user}
            </button>
          </li>
        ))}
      </ul>
    </div>
  );
};

export default UserList;
