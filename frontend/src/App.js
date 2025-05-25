// import React, { useState, useEffect } from 'react';
import React from 'react';
import { BrowserRouter as Router, Route, Routes, Navigate } from 'react-router-dom';
import axios from 'axios';
import './App.css';
import Login from './login'; // Import the Login component
import { Button } from '@mui/material'; 
import EmpHome from './emp/emp-home';
import FillInvoice from './emp/fill-invoice';

const App = () => {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<Navigate to='/login' replace/>}  />
        <Route path='/login' element={<Login />} />

        <Route path='/emp-home' element={<EmpHome />}/>
        <Route path='/fill-invoice' element={<FillInvoice />}/>
      </Routes>
    </Router>
  );
}

export default App;