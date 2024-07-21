import React, { useEffect, useState, useRef } from "react";
import axios from "axios";
import { useParams, useLocation, Link } from "react-router-dom";
import { FaArrowUp, FaArrowDown } from "react-icons/fa";
import "./styles/Chat.css";

interface Message {
  id: string;
  sender: string;
  receiver: string;
  content: string;
  upvotes: number;
  downvotes: number;
}

/**
 * Chat component manages a real-time chat interface between users.
 * It displays messages, allows sending messages, and handles WebSocket communication for real-time updates.
 */
const Chat: React.FC = () => {
  // Retrieves the username from the URL params
  const { username } = useParams<{ username: string }>();

  // State variables to manage messages, current message being typed, and WebSocket connection
  const [messages, setMessages] = useState<Message[]>([]);
  const [message, setMessage] = useState("");
  const [ws, setWs] = useState<WebSocket | null>(null);

  // Retrieves current user's username from URL query parameters
  const location = useLocation();
  const currentUser =
    new URLSearchParams(location.search).get("currentUser") || "";

  // Reference for scrolling to the bottom of messages container
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Scrolls to the bottom of the messages container whenever messages update
  const scrollToBottom = () => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: "smooth" });
    }
  };

  // Fetches initial messages between current user and the selected chat user
  useEffect(() => {
    const fetchMessages = async () => {
      try {
        const response = await axios.get(
          `http://127.0.0.1:8080/messages?sender=${currentUser}&receiver=${username}`
        );
        console.log("API Response:", response.data);
        setMessages(response.data.messages || []);
      } catch (error) {
        console.error("Error fetching messages:", error);
      }
    };

    fetchMessages();
  }, [username, currentUser]);

  // Sets up WebSocket connection for real-time message updates
  useEffect(() => {
    const socket = new WebSocket(
      "ws://127.0.0.1:8080/ws?user_id=" + currentUser
    );
    setWs(socket);

    // Listens for incoming messages and updates state accordingly
    socket.onmessage = (event) => {
      const updatedMessage: Message = JSON.parse(event.data);
      if (
        (updatedMessage.sender === currentUser &&
          updatedMessage.receiver === username) ||
        (updatedMessage.sender === username &&
          updatedMessage.receiver === currentUser)
      ) {
        setMessages((prevMessages) => {
          const messageIndex = prevMessages.findIndex(
            (msg) => msg.id === updatedMessage.id
          );

          if (messageIndex !== -1) {
            return prevMessages.map((msg, index) =>
              index === messageIndex ? updatedMessage : msg
            );
          } else {
            return [...prevMessages, updatedMessage];
          }
        });
      }
    };

    // Cleans up WebSocket connection when component unmounts
    return () => {
      socket.close();
    };
  }, [currentUser, username]);

  // Scrolls to the bottom of the messages container whenever messages update
  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // Sends the current message to the server
  const handleSendMessage = async () => {
    if (message.trim() === "") return;

    try {
      const response = await axios.post("http://127.0.0.1:8080/messages", {
        sender: currentUser,
        receiver: username,
        content: message,
      });

      const insertedMessage: Message = response.data.message;

      const newMessage: Message = {
        id: insertedMessage.id,
        sender: insertedMessage.sender,
        receiver: insertedMessage.receiver,
        content: insertedMessage.content,
        upvotes: insertedMessage.upvotes,
        downvotes: insertedMessage.downvotes,
      };

      setMessage("");
    } catch (error) {
      console.error("Error sending message:", error);
    }
  };

  // Handles upvoting a message by its ID
  const handleUpvote = async (messageId: string) => {
    try {
      console.log("handleUpvote: " + messageId);
      await axios.post(
        `http://127.0.0.1:8080/messages/${messageId}/upvote`,
        null,
        {
          params: { user_id: currentUser },
        }
      );
    } catch (error) {
      console.error("Error upvoting message:", error);
    }
  };

  // Handles downvoting a message by its ID
  const handleDownvote = async (messageId: string) => {
    console.log("handleDownvote: " + messageId);
    try {
      await axios.post(
        `http://127.0.0.1:8080/messages/${messageId}/downvote`,
        null,
        {
          params: { user_id: currentUser },
        }
      );
    } catch (error) {
      console.error("Error downvoting message:", error);
    }
  };

  // Logs out the current user by removing username from localStorage
  const handleLogout = () => {
    localStorage.removeItem("username");
  };

  // Renders the chat interface with messages, input box, and buttons
  return (
    <div className="container chat-container">
      <div className="top-bar d-flex justify-content-between align-items-center mb-3 mt-3">
        <div className="left-buttons">
          <Link
            to={`/users?username=${currentUser}`}
            className="btn btn-primary back-button"
          >
            Back
          </Link>
        </div>
        <div className="right-buttons">
          <Link
            to="/"
            onClick={handleLogout}
            className="btn btn-danger logout-button"
          >
            Logout
          </Link>
        </div>
      </div>
      <h2 className="mt-4 mb-3">Chat with {username}</h2>
      <div className="chat-messages">
        {messages.map((msg) => (
          <div
            key={msg.id}
            className={`messageContainer ${
              msg.sender === currentUser
                ? "currentUserMessage"
                : "otherUserMessage"
            }`}
          >
            <div className="message-content">
              <strong>{msg.sender}:</strong> {msg.content}
            </div>
            <div className="vote-buttons">
              <button
                className="voteButton"
                onClick={() => handleUpvote(msg.id)}
              >
                <FaArrowUp className="voteIcon upvote" />{" "}
                <span className="upvote-count">{msg.upvotes}</span>
              </button>
              <button
                className="voteButton"
                onClick={() => handleDownvote(msg.id)}
              >
                <FaArrowDown className="voteIcon downvote" />{" "}
                <span className="downvote-count">{msg.downvotes}</span>
              </button>
            </div>
          </div>
        ))}
        <div ref={messagesEndRef}></div>
      </div>
      <div className="message-input-container">
        <input
          type="text"
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          className="form-control message-input"
          placeholder="Type your message..."
        />
        <button
          onClick={handleSendMessage}
          className="btn btn-primary send-button"
        >
          Send
        </button>
      </div>
    </div>
  );
};

export default Chat;
