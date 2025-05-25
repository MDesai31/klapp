import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';

const EmpHome = () => {
    const navigate = useNavigate();

    const handleLogout = () => {
        localStorage.removeItem('token');
        localStorage.removeItem('username');
        localStorage.removeItem('role');
        navigate('/login'); // Redirect to login page after logout
    };

    const handleFillInvoice = () => {
        // Navigate to the fill invoice page
        navigate('/fill-invoice');
    };

    const handleCheckHours = () => {
        // Navigate to the check hours page
        navigate('/check-hours');
    };

    const handleReportNewHours = () => {
        // Navigate to the report new hours page
        navigate('/report-new-hours');
    };

    return (
        <div>
            <div>
                <h1>
                    Welcome!
                </h1>
                <h2>
                    Employee Dashboard
                </h2>

                <div>
                    <button
                        onClick={handleFillInvoice}
                    >
                        Fill Invoice
                    </button>
                    <button
                        onClick={handleCheckHours}
                    >
                        Check Hours
                    </button>
                    <button
                        onClick={handleReportNewHours}
                    >
                        Report New Hours
                    </button>
                </div>

                <button
                    onClick={handleLogout}
                >
                    Logout
                </button>
            </div>
        </div>
    );
};

export default EmpHome;