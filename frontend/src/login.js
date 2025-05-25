import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

const Login = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const navigate = useNavigate();
  const [error, setError] = useState(''); 
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError(''); 
    setLoading(true);

    try {
      const response = await axios.post('http://localhost:5000/api/login', {
        username,
        password,
      });
      console.log('reaches here')
      if (response.status === 200) {
        const { token, role } = response.data;
        alert(`Login successful! Token: ${token}`);
        localStorage.setItem('token', token);
        localStorage.setItem('role', role);
        // Redirect the user based on their role
        if (role === 'admin') {
          navigate('/admin-home');
        } else {
          navigate('/emp-home');
        }
      } else {
        setError('Login failed: Invalid response from server.'); // Handle non-200 responses
      }
    } catch (err) {
      console.error(err);
      if (err.response) {
        setError(err.response.data.message || 'Invalid credentials');
      } else {
        setError('An unexpected error occurred.');
      }

    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <form onSubmit={handleSubmit}>
        <h2>Login</h2>
        <input
          type="text"
          placeholder="Username"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          disabled={loading}
        />
        <input
          type="password"
          placeholder="Password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          disabled={loading}
        />
        <button type="submit" disabled={loading}>
          {loading ? 'Logging in...' : 'Login'}
        </button>
      </form>
      {error && <p style={{ color: 'red' }}>{error}</p>}
      <button onClick={() => navigate('/register')}>Go to Register</button>
    </div>
  );
};

export default Login;