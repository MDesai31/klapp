import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';

const FillInvoice = () => {
    const navigate = useNavigate();

    // State for invoice details
    const [customers, setCustomers] = useState([]);
    const [selectedCustomer, setSelectedCustomer] = useState('');
    const [invoiceDate, setInvoiceDate] = useState('');
    const [timeArrived, setTimeArrived] = useState('');
    const [timeLeft, setTimeLeft] = useState('');
    const [numWorkers, setNumWorkers] = useState('');

    const [employees, setEmployees] = useState([]);
    const [selectedEmployees, setSelectedEmployees] = useState([]); // Array of selected employee IDs

    const [jobs, setJobs] = useState([]);
    const [selectedJobs, setSelectedJobs] = useState([]); // Array of selected job IDs
    const [otherJobDescription, setOtherJobDescription] = useState('');

    const [materials, setMaterials] = useState([]);
    const [selectedMaterials, setSelectedMaterials] = useState([]); // Array of selected material IDs

    const [additionalComments, setAdditionalComments] = useState('');

    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [successMessage, setSuccessMessage] = useState('');

    // Fetch initial data (customers, employees, jobs, materials)
    useEffect(() => {
        const fetchData = async () => {
            setLoading(true);
            setError('');
            try {
                const token = localStorage.getItem('token');
                if (!token) {
                    navigate('/login');
                    return;
                }

                const headers = { 'Authorization': `Bearer ${token}` };

                const [
                    customersRes,
                    employeesRes,
                    jobsRes,
                    materialsRes
                ] = await Promise.all([
                    axios.get('http://localhost:5000/api/customers', { headers }),
                    axios.get('http://localhost:5000/api/employees', { headers }),
                    axios.get('http://localhost:5000/api/jobs', { headers }),
                    axios.get('http://localhost:5000/api/materials', { headers })
                ]);

                setCustomers(customersRes.data);
                setEmployees(employeesRes.data);
                setJobs(jobsRes.data);
                setMaterials(materialsRes.data);

            } catch (err) {
                console.error("Error fetching initial data:", err);
                if (err.response && err.response.status === 401) {
                    setError('Session expired. Please log in again.');
                    localStorage.clear();
                    navigate('/login');
                } else {
                    setError('Failed to load form data. Please try again.');
                }
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, [navigate]);

    // Handle dynamic employee selection based on numWorkers
    useEffect(() => {
        const num = parseInt(numWorkers, 10);
        if (isNaN(num) || num <= 0) {
            setSelectedEmployees([]); // Clear selection if invalid number
        } else {
            // Ensure selectedEmployees array has the correct size, preserving existing selections
            setSelectedEmployees(prev => {
                const newSelection = Array(num).fill(null);
                for (let i = 0; i < num; i++) {
                    newSelection[i] = prev[i] || ''; // Use previous selection or empty string
                }
                return newSelection;
            });
        }
    }, [numWorkers]);

    const handleEmployeeSelection = (index, empId) => {
        setSelectedEmployees(prev => {
            const newSelections = [...prev];
            newSelections[index] = empId;
            return newSelections;
        });
    };

    const handleJobSelection = (e) => {
        const options = Array.from(e.target.options)
            .filter(option => option.selected)
            .map(option => option.value);
        setSelectedJobs(options);
    };

    const handleMaterialSelection = (e) => {
        const options = Array.from(e.target.options)
            .filter(option => option.selected)
            .map(option => option.value);
        setSelectedMaterials(options);
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setLoading(true);
        setError('');
        setSuccessMessage('');

        // Basic validation
        if (!selectedCustomer || !invoiceDate || !timeArrived || !timeLeft || !numWorkers || numWorkers <= 0) {
            setError('Please fill in all required invoice details.');
            setLoading(false);
            return;
        }

        if (selectedEmployees.filter(id => id !== '' && id !== null).length !== parseInt(numWorkers, 10)) {
            setError(`Please select exactly ${numWorkers} employees.`);
            setLoading(false);
            return;
        }

        if (selectedJobs.length === 0 && !otherJobDescription) {
            setError('Please select at least one job or provide an "Other" description.');
            setLoading(false);
            return;
        }

        const invoiceData = {
            customer_id: selectedCustomer,
            date: invoiceDate,
            time_arrived: timeArrived,
            time_left: timeLeft,
            num_workers: parseInt(numWorkers, 10),
            employee_ids: selectedEmployees.filter(id => id !== '' && id !== null), // Filter out empty selections
            job_ids: selectedJobs.filter(id => id !== 'other'), // Filter out 'other' if selected
            other_job_description: selectedJobs.includes('other') ? otherJobDescription : null,
            material_ids: selectedMaterials,
            additional_comments: additionalComments
        };

        try {
            const token = localStorage.getItem('token');
            const headers = { 'Authorization': `Bearer ${token}` };
            const response = await axios.post('http://localhost:5000/api/invoice', invoiceData, { headers });

            if (response.status === 201) {
                setSuccessMessage('Invoice submitted successfully!');
                // Optionally clear form or redirect
                // navigate('/emp-home');
            } else {
                setError(response.data.message || 'Failed to submit invoice.');
            }
        } catch (err) {
            console.error("Error submitting invoice:", err);
            if (err.response) {
                setError(err.response.data.message || 'Error submitting invoice.');
            } else {
                setError('Network error or server unreachable.');
            }
        } finally {
            setLoading(false);
        }
    };

    return (
        <div style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            minHeight: '100vh',
            backgroundColor: '#f5f5f5',
            padding: '16px'
        }}>
            <div style={{
                width: '100%',
                maxWidth: '600px',
                padding: '40px',
                backgroundColor: '#ffffff',
                borderRadius: '12px',
                boxShadow: '0px 8px 24px rgba(0, 0, 0, 0.15)',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                textAlign: 'center'
            }}>
                <h1 style={{
                    fontSize: '2.2rem',
                    fontWeight: '700',
                    color: '#4CAF50',
                    marginBottom: '1rem',
                }}>
                    Fill Invoice
                </h1>

                {loading && <p>Loading form data...</p>}
                {error && <p style={{ color: 'red' }}>Error: {error}</p>}
                {successMessage && <p style={{ color: 'green' }}>{successMessage}</p>}

                <form onSubmit={handleSubmit} style={{ width: '100%', display: 'flex', flexDirection: 'column', gap: '15px' }}>
                    {/* Customer Name */}
                    <div>
                        <label htmlFor="customer" style={{ display: 'block', marginBottom: '5px', fontWeight: 'bold' }}>Customer Name:</label>
                        <select
                            id="customer"
                            value={selectedCustomer}
                            onChange={(e) => setSelectedCustomer(e.target.value)}
                            required
                            style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ccc' }}
                            disabled={loading}
                        >
                            <option value="">Select Customer</option>
                            {customers.map(customer => (
                                <option key={customer.customer_id} value={customer.customer_id}>
                                    {customer.name}
                                </option>
                            ))}
                        </select>
                    </div>

                    {/* Date */}
                    <div>
                        <label htmlFor="invoiceDate" style={{ display: 'block', marginBottom: '5px', fontWeight: 'bold' }}>Date:</label>
                        <input
                            type="date"
                            id="invoiceDate"
                            value={invoiceDate}
                            onChange={(e) => setInvoiceDate(e.target.value)}
                            required
                            style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ccc' }}
                            disabled={loading}
                        />
                    </div>

                    {/* Time Arrived */}
                    <div>
                        <label htmlFor="timeArrived" style={{ display: 'block', marginBottom: '5px', fontWeight: 'bold' }}>Time Arrived:</label>
                        <input
                            type="time"
                            id="timeArrived"
                            value={timeArrived}
                            onChange={(e) => setTimeArrived(e.target.value)}
                            required
                            style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ccc' }}
                            disabled={loading}
                        />
                    </div>

                    {/* Time Left */}
                    <div>
                        <label htmlFor="timeLeft" style={{ display: 'block', marginBottom: '5px', fontWeight: 'bold' }}>Time Left:</label>
                        <input
                            type="time"
                            id="timeLeft"
                            value={timeLeft}
                            onChange={(e) => setTimeLeft(e.target.value)}
                            required
                            style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ccc' }}
                            disabled={loading}
                        />
                    </div>

                    {/* Number of Workers */}
                    <div>
                        <label htmlFor="numWorkers" style={{ display: 'block', marginBottom: '5px', fontWeight: 'bold' }}>Number of Workers:</label>
                        <input
                            type="number"
                            id="numWorkers"
                            value={numWorkers}
                            onChange={(e) => setNumWorkers(e.target.value)}
                            min="1"
                            required
                            style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ccc' }}
                            disabled={loading}
                        />
                    </div>

                    {/* Employee Selection (dynamic based on numWorkers) */}
                    {parseInt(numWorkers, 10) > 0 && (
                        <div>
                            <label style={{ display: 'block', marginBottom: '5px', fontWeight: 'bold' }}>Select Employees:</label>
                            {selectedEmployees.map((empId, index) => (
                                <select
                                    key={index}
                                    value={empId}
                                    onChange={(e) => handleEmployeeSelection(index, e.target.value)}
                                    required
                                    style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ccc', marginBottom: '5px' }}
                                    disabled={loading}
                                >
                                    <option value="">Select Employee {index + 1}</option>
                                    {employees.map(employee => (
                                        <option key={employee.emp_id} value={employee.emp_id}>
                                            {employee.name}
                                        </option>
                                    ))}
                                </select>
                            ))}
                        </div>
                    )}

                    {/* Choose Jobs */}
                    <div>
                        <label htmlFor="jobs" style={{ display: 'block', marginBottom: '5px', fontWeight: 'bold' }}>Choose Jobs (Ctrl/Cmd + Click to select multiple):</label>
                        <select
                            id="jobs"
                            multiple
                            value={selectedJobs}
                            onChange={handleJobSelection}
                            style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ccc', minHeight: '100px' }}
                            disabled={loading}
                        >
                            {jobs.map(job => (
                                <option key={job.job_id} value={job.job_id}>
                                    {job.description}
                                </option>
                            ))}
                            <option value="other">Other (specify below)</option>
                        </select>
                    </div>

                    {/* Other Job Description */}
                    {selectedJobs.includes('other') && (
                        <div>
                            <label htmlFor="otherJob" style={{ display: 'block', marginBottom: '5px', fontWeight: 'bold' }}>Other Job Description:</label>
                            <input
                                type="text"
                                id="otherJob"
                                value={otherJobDescription}
                                onChange={(e) => setOtherJobDescription(e.target.value)}
                                required={selectedJobs.includes('other')}
                                style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ccc' }}
                                disabled={loading}
                            />
                        </div>
                    )}

                    {/* Materials Used */}
                    <div>
                        <label htmlFor="materials" style={{ display: 'block', marginBottom: '5px', fontWeight: 'bold' }}>Materials Used (Ctrl/Cmd + Click to select multiple):</label>
                        <select
                            id="materials"
                            multiple
                            value={selectedMaterials}
                            onChange={handleMaterialSelection}
                            style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ccc', minHeight: '100px' }}
                            disabled={loading}
                        >
                            {materials.map(material => (
                                <option key={material.material_id} value={material.material_id}>
                                    {material.description}
                                </option>
                            ))}
                        </select>
                    </div>

                    {/* Additional Comments */}
                    <div>
                        <label htmlFor="comments" style={{ display: 'block', marginBottom: '5px', fontWeight: 'bold' }}>Additional Comments (Optional):</label>
                        <textarea
                            id="comments"
                            value={additionalComments}
                            onChange={(e) => setAdditionalComments(e.target.value)}
                            rows="4"
                            style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ccc' }}
                            disabled={loading}
                        ></textarea>
                    </div>

                    <button
                        type="submit"
                        disabled={loading}
                    >
                        {loading ? 'Submitting...' : 'Submit Invoice'}
                    </button>
                </form>

                <button
                    onClick={() => navigate('/emp-home')}
                >
                    Back to Dashboard
                </button>
            </div>
        </div>
    );
};

export default FillInvoice;