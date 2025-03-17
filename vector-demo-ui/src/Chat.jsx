import React, { useState } from 'react';
import { TextField, Button, Paper, Typography, CircularProgress, Box, Container, Avatar } from '@mui/material';

import logo from './assets/logo.png';

const Chat = () => {
  const [query, setQuery] = useState('');
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!query.trim()) return;

    const userMessage = { role: 'user', content: query };
    setMessages((prev) => [...prev, userMessage]);
    setLoading(true);

    try {
      const res = await fetch('http://localhost:8080/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query }),
      });
      const data = await res.json();

      const botMessage = { role: 'bot', content: data.answer };
      setMessages((prev) => [...prev, botMessage]);
    } catch (error) {
      console.error('Error fetching:', error);
      setMessages((prev) => [...prev, { role: 'bot', content: 'Error: Unable to get response.' }]);
    } finally {
      setLoading(false);
      setQuery('');
    }
  };

  return (
    <Container maxWidth="sm" sx={{ mt: 4, backgroundColor: '#F0F4F8', borderRadius: 1, padding: 4 }}>
      {/* Header Section with Logo */}
      <Typography variant="h4" sx={{ textAlign: 'center', mb: 3, fontWeight: 'bold', color: '#3F51B5' }}>
        {/* Center the logo */}
        <Avatar
      src={logo}
      alt="90DaysOfDevOps"
      className="logo" // Apply the class here
      sx={{
        display: 'block',
        margin: '0 auto',
        width: 140,
        height: 140,
        objectFit: 'contain',
        borderRadius: '50%',
      }}
        />
        💬 AI Chat Assistant
      </Typography>

      <Paper elevation={3} sx={{ p: 2, backgroundColor: '#fff', borderRadius: 2, boxShadow: '0px 4px 6px rgba(0, 0, 0, 0.1)', maxHeight: '70vh', overflowY: 'auto', mb: 2 }}>
        {messages.map((msg, index) => (
          <Box key={index} sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
            <Typography
              variant="body1"
              sx={{
                backgroundColor: msg.role === 'user' ? '#3F51B5' : '#E0E0E0',
                color: msg.role === 'user' ? '#fff' : '#000',
                borderRadius: '20px',
                padding: '10px 15px',
                maxWidth: '70%',
                marginLeft: msg.role === 'user' ? 'auto' : 0,
                marginRight: msg.role === 'user' ? 0 : 'auto',
              }}
            >
              {msg.content}
            </Typography>
          </Box>
        ))}
        {loading && <CircularProgress size={24} />}
      </Paper>

      {/* Input Section */}
      <form onSubmit={handleSubmit} sx={{ display: 'flex', gap: 2, mt: 2 }}>
        <TextField
          variant="outlined"
          fullWidth
          label="Ask a question..."
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          sx={{ borderRadius: 3, backgroundColor: '#FFF' }}
        />
        <Button
          type="submit"
          variant="contained"
          color="primary"
          disabled={loading}
          sx={{ borderRadius: 3 }}
        >
          Send
        </Button>
      </form>
    </Container>
  );
};

export default Chat;
