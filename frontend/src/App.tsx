import React from "react";
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Link,
  useLocation,
} from "react-router-dom";
import LoginForm from "./components/LoginForm";
import SignupForm from "./components/SignupForm";
import UserList from "./components/UserList";
import Chat from "./components/Chat";

/**
 * App component sets up routing using React Router.
 * It defines routes for different pages such as login, signup, user list, and chat.
 */
const App: React.FC = () => {
  return (
    <Router>
      <div>
        <main>
          <Routes>
            <Route path="/signup" element={<SignupPage />} />
            <Route path="/" element={<LoginPage />} />
            <Route path="/users" element={<UserList />} />
            <Route path="/chat/:username" element={<Chat />} />
          </Routes>
        </main>
      </div>
    </Router>
  );
};

/**
 * LoginPage component renders the login form and navigation links.
 * It uses React Router's useLocation hook to conditionally render signup link.
 */
const LoginPage: React.FC = () => {
  const location = useLocation();
  return (
    <div className="row justify-content-center">
      <div className="col-md-6">
        <div className="card bg-light">
          <div className="card-body">
            <div className="form-group mb-3">
              <LoginForm />
            </div>
            <div className="form-group mb-3">
              <p className="mt-3 text-center">
                New user? <Link to="/signup">Signup here</Link>
              </p>
              {location.pathname === "/signup" && (
                <p>
                  Already have an account? <Link to="/">Back to login</Link>
                </p>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

/**
 * SignupPage component renders the signup form and navigation links.
 * It uses React Router's useLocation hook to conditionally render login link.
 */
const SignupPage: React.FC = () => {
  const location = useLocation();
  return (
    <div className="row justify-content-center">
      <div className="col-md-6">
        <div className="card bg-light">
          <div className="card-body">
            <div className="form-group mb-3"></div>
            <SignupForm />
          </div>
          <div className="form-group mb-3">
            <p className="mt-3 text-center">
              Already have an account? <Link to="/">Back to login</Link>
            </p>
            {location.pathname === "/" && (
              <p>
                New user? <Link to="/signup">Signup here</Link>
              </p>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default App;
