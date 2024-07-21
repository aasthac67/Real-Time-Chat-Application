import React, { useState } from "react";
import axios, { AxiosError } from "axios";

/**
 * SignupForm component provides a form for user signup.
 * It handles user input validation, sends signup request to the server,
 * and displays error messages if signup fails.
 */
const SignupForm: React.FC = () => {
  // State variables for username, password, and error message
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // Handles signup button click event
  const handleSignup = async () => {
    // Validation: Checks if username and password are provided
    if (!username || !password) {
      setErrorMessage("Username and password are required.");
      return;
    }

    // Validation: Checks if password length is between 8 to 20 characters
    if (password.length < 8 || password.length > 20) {
      setErrorMessage("Password must be between 8 to 20 characters.");
      return;
    }

    try {
      // Sends signup request to the server
      const response = await axios.post<{ message?: string; error?: string }>(
        "http://127.0.0.1:8080/signup",
        { username, password }
      );

      // Checks server response for errors
      if (response.data.error) {
        setErrorMessage(response.data.error);
      } else {
        setErrorMessage(null);
      }

      // Redirects to homepage on successful signup
      window.location.href = "/";
    } catch (error) {
      // Handles errors from axios request
      if (axios.isAxiosError(error)) {
        const axiosError = error as AxiosError<{ error?: string }>;
        if (
          axiosError.response &&
          axiosError.response.data &&
          axiosError.response.data.error
        ) {
          setErrorMessage(axiosError.response.data.error);
        } else {
          setErrorMessage("Failed to sign up. Please try again later.");
        }
      } else {
        setErrorMessage("Failed to sign up. Please try again later.");
      }
      console.error("Signup error:", error);
    }
  };

  // Renders the signup form with username, password inputs, and signup button
  return (
    <div className="card-body">
      <h2 className="card-title text-center mb-4">Signup Form</h2>
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
      <button className="btn btn-primary btn-block" onClick={handleSignup}>
        Signup
      </button>
      {errorMessage && (
        <div className="form-group mb-3">
          <p style={{ color: "red" }}>{errorMessage}</p>
        </div>
      )}
    </div>
  );
};

export default SignupForm;
