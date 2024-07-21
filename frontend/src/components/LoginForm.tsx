import React, { useState } from "react";
import axios, { AxiosError } from "axios";
import { useNavigate } from "react-router-dom";
import "bootstrap/dist/css/bootstrap.min.css";

/**
 * LoginForm component handles user authentication via username and password.
 * Displays a form for users to input their credentials and log in.
 */
const LoginForm: React.FC = () => {
  // State variables to hold username, password, and error message
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // Hook from React Router DOM for programmatic navigation
  const navigate = useNavigate();

  /**
   * Handles the login button click.
   * Sends a POST request to the server with username and password.
   * On success, navigates to the user list page.
   * On failure, displays an error message.
   */
  const handleLogin = async () => {
    try {
      // Send login request to the server
      const response = await axios.post("http://127.0.0.1:8080/login", {
        username,
        password,
      });

      // Log server response and reset error message
      console.log(response.data);
      setErrorMessage(null);

      // Navigate to the user list page with username query parameter
      navigate(`/users?username=${username}`);
    } catch (error) {
      // Handle axios errors, display appropriate error message
      if (axios.isAxiosError(error)) {
        const axiosError = error as AxiosError<{ error?: string }>;
        if (
          axiosError.response &&
          axiosError.response.data &&
          axiosError.response.data.error
        ) {
          setErrorMessage(axiosError.response.data.error);
        } else {
          setErrorMessage("Failed to login. Please try again later.");
        }
      } else {
        setErrorMessage("Failed to login. Please try again later.");
      }
      console.error("Login error:", error);
    }
  };

  // Render the login form with username, password inputs and error message
  return (
    <div className="card-body">
      <h2 className="card-title text-center mb-4">Login Form</h2>
      <form>
        <div className="form-group mb-3">
          <label htmlFor="username">Username</label>
          <input
            type="text"
            className="form-control"
            id="username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
          />
        </div>
        <div className="form-group mb-3">
          <label htmlFor="password">Password</label>
          <input
            type="password"
            className="form-control"
            id="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </div>
        <button
          type="button"
          className="btn btn-primary btn-block"
          onClick={handleLogin}
        >
          Login
        </button>
        {errorMessage && (
          <p className="text-danger text-center mt-3">{errorMessage}</p>
        )}
      </form>
    </div>
  );
};

export default LoginForm;
