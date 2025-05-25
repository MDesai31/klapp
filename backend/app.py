from flask import Flask, jsonify, request
from flask_cors import CORS 
import jwt
import datetime
from functools import wraps
import os
import mysql.connector

app = Flask(__name__)
CORS(app)

# Secret key for JWT (use a strong, random string -  NEVER store it directly in the code)
app.config['SECRET_KEY'] = os.environ.get('SECRET_KEY', 'your-secret-key') #  fallback

# MySQL Database Configuration
DB_CONFIG = {
    'host': 'localhost',
    'user': 'root',
    'password': 'M@nthan2001',
    'database': 'klapp'
}

# Helper function to get a database connection
def get_db_connection():
    try:
        cnx = mysql.connector.connect(**DB_CONFIG)
        return cnx
    except mysql.connector.Error as err:
        print(f"Error connecting to database: {err}")
        return None

# Helper function to close database connection and cursor
# Ensure this function is defined before any routes that use it.
def close_db_connection(cnx, cursor):
    if cursor:
        cursor.close()
    if cnx and cnx.is_connected():
        cnx.close()

# Function to generate a JWT token
def generate_token(username, role):
    """
    Generates a JSON Web Token (JWT) for the given username and role.
    """
    payload = {
        'exp': datetime.datetime.utcnow() + datetime.timedelta(hours=24),
        'iat': datetime.datetime.utcnow(), 
        'sub': username,
        'role': role
    }
    return jwt.encode(payload, app.config['SECRET_KEY'], algorithm='HS256')

# Function to verify the JWT token
def verify_token(f):
    """
    Decorator function to verify the JWT token in the request headers.
    """
    @wraps(f)
    def decorated(*args, **kwargs):
        token = request.headers.get('Authorization')
        if not token:
            return jsonify({'message': 'Token is missing!'}), 401

        try:
            token = token.split(" ")[1]
            data = jwt.decode(token, app.config['SECRET_KEY'], algorithms=['HS256'])
        except jwt.ExpiredSignatureError:
            return jsonify({'message': 'Token has expired!'}), 401
        except jwt.InvalidTokenError:
            return jsonify({'message': 'Invalid token!'}), 401

        return f(data, *args, **kwargs)
    return decorated

@app.route('/api/login', methods=['POST'])
def login():
    """
    Handles user login.  Expects username and password in the request body.
    Returns a JWT token on successful authentication.
    """
    auth = request.get_json()
    if not auth or not auth.get('username') or not auth.get('password'):
        return jsonify({'message': 'Please provide username and password'}), 400

    username = auth.get('username')
    password = auth.get('password')

    user_data = None
    try:
        # Establish database connection
        cnx = mysql.connector.connect(**DB_CONFIG)
        cursor = cnx.cursor(dictionary=True)

        # Query the 'authentication' table
        query = "SELECT username, password, is_admin FROM users WHERE username = %s"
        cursor.execute(query, (username,))
        user_data = cursor.fetchone()

    except mysql.connector.Error as err:
        print(f"Error: {err}")
        return jsonify({'message': 'Database error during login.'}), 500
    finally:
        if 'cnx' in locals() and cnx.is_connected():
            cursor.close()
            cnx.close()

    if not user_data:
        return jsonify({'message': 'Invalid username'}), 401

    if not user_data['password'] == password:
        return jsonify({'message': 'Invalid password'}), 401

    # Determine the role based on is_admin flag
    role = "admin" if user_data['is_admin'] else "user"

    # Authentication successful, generate a token
    token = generate_token(username, role) # Pass the role here
    return jsonify({'success': True, 'token': token, 'message': 'Login successful!', 'role': role}), 200

@app.route('/api/customers', methods=['GET'])
@verify_token
def get_customers(user_data):
    cnx = None
    cursor = None
    try:
        cnx = get_db_connection()
        if not cnx:
            return jsonify({'message': 'Database connection error'}), 500
        cursor = cnx.cursor(dictionary=True)
        cursor.execute("SELECT customer_id, name FROM customer")
        customers = cursor.fetchall()
        return jsonify(customers), 200
    except Exception as e:
        print(f"Error fetching customers: {e}")
        return jsonify({'message': 'Failed to fetch customers'}), 500
    finally:
        close_db_connection(cnx, cursor)

@app.route('/api/employees', methods=['GET'])
@verify_token
def get_employees(user_data):
    cnx = None
    cursor = None
    try:
        cnx = get_db_connection()
        if not cnx:
            return jsonify({'message': 'Database connection error'}), 500
        cursor = cnx.cursor(dictionary=True)
        cursor.execute("SELECT emp_id, name FROM employee")
        employees = cursor.fetchall()
        return jsonify(employees), 200
    except Exception as e:
        print(f"Error fetching employees: {e}")
        return jsonify({'message': 'Failed to fetch employees'}), 500
    finally:
        close_db_connection(cnx, cursor)

@app.route('/api/jobs', methods=['GET'])
@verify_token
def get_jobs(user_data):
    cnx = None
    cursor = None
    try:
        cnx = get_db_connection()
        if not cnx:
            return jsonify({'message': 'Database connection error'}), 500
        cursor = cnx.cursor(dictionary=True)
        cursor.execute("SELECT job_id, description FROM jobs")
        jobs = cursor.fetchall()
        return jsonify(jobs), 200
    except Exception as e:
        print(f"Error fetching jobs: {e}")
        return jsonify({'message': 'Failed to fetch jobs'}), 500
    finally:
        close_db_connection(cnx, cursor)

@app.route('/api/materials', methods=['GET'])
@verify_token
def get_materials(user_data):
    cnx = None
    cursor = None
    try:
        cnx = get_db_connection()
        if not cnx:
            return jsonify({'message': 'Database connection error'}), 500
        cursor = cnx.cursor(dictionary=True)
        cursor.execute("SELECT material_id, description FROM materials")
        materials = cursor.fetchall()
        return jsonify(materials), 200
    except Exception as e:
        print(f"Error fetching materials: {e}")
        return jsonify({'message': 'Failed to fetch materials'}), 500
    finally:
        close_db_connection(cnx, cursor)

# --- API Endpoint for Submitting Invoice ---

@app.route('/api/invoice', methods=['POST'])
@verify_token
def submit_invoice(user_data):
    cnx = None
    cursor = None
    try:
        invoice_data = request.get_json()

        customer_id = invoice_data.get('customer_id')
        invoice_date = invoice_data.get('date')
        time_arrived = invoice_data.get('time_arrived')
        time_left = invoice_data.get('time_left')
        num_workers = invoice_data.get('num_workers')
        employee_ids = invoice_data.get('employee_ids', [])
        job_ids = invoice_data.get('job_ids', [])
        other_job_description = invoice_data.get('other_job_description')
        material_ids = invoice_data.get('material_ids', [])
        additional_comments = invoice_data.get('additional_comments')

        # Basic validation (more robust validation should be done here)
        if not all([customer_id, invoice_date, time_arrived, time_left, num_workers]):
            return jsonify({'message': 'Missing required invoice fields.'}), 400

        cnx = get_db_connection()
        if not cnx:
            return jsonify({'message': 'Database connection error'}), 500
        cursor = cnx.cursor() # Use non-dictionary cursor for inserts

        # 1. Insert into invoice table
        invoice_insert_query = """
            INSERT INTO invoice (customer_id, date, time_arrived, time_left, num_workers)
            VALUES (%s, %s, %s, %s, %s)
        """
        cursor.execute(invoice_insert_query, (customer_id, invoice_date, time_arrived, time_left, num_workers))
        invoice_id = cursor.lastrowid # Get the ID of the newly inserted invoice

        # 2. Insert into invoice_workers
        for emp_id in employee_ids:
            cursor.execute("INSERT INTO invoice_workers (invoice_id, emp_id) VALUES (%s, %s)", (invoice_id, emp_id))

        # 3. Insert into invoice_jobs
        # If 'other' job description is provided, you might want to add it to the 'jobs' table first
        # and then link its new job_id to the invoice. For simplicity, we'll just link existing jobs.
        for job_id in job_ids:
            cursor.execute("INSERT INTO invoice_jobs (invoice_id, job_id) VALUES (%s, %s)", (invoice_id, job_id))

        # Handle 'other' job description: If it's a new job, insert it into the 'jobs' table
        # and then link it to the invoice.
        if other_job_description:
            # Check if this "other" job description already exists in the jobs table
            cursor.execute("SELECT job_id FROM jobs WHERE description = %s", (other_job_description,))
            existing_job = cursor.fetchone()
            if existing_job:
                other_job_id = existing_job[0]
            else:
                # If it doesn't exist, insert it as a new job
                cursor.execute("INSERT INTO jobs (description) VALUES (%s)", (other_job_description,))
                other_job_id = cursor.lastrowid
            # Link the 'other' job to the invoice
            cursor.execute("INSERT INTO invoice_jobs (invoice_id, job_id) VALUES (%s, %s)", (invoice_id, other_job_id))


        # 4. Insert into invoice_materials
        for material_id in material_ids:
            cursor.execute("INSERT INTO invoice_materials (invoice_id, material_id) VALUES (%s, %s)", (invoice_id, material_id))

        # 5. Handle additional comments: Add a column to the 'invoice' table for this, or a separate table.
        # For now, let's assume you've added a 'comments' column to the 'invoice' table.
        # If not, you'd need to modify your 'invoice' table schema.
        if additional_comments:
            cursor.execute("UPDATE invoice SET comments = %s WHERE invoice_id = %s", (additional_comments, invoice_id))


        cnx.commit() # Commit the transaction

        return jsonify({'message': 'Invoice submitted successfully!', 'invoice_id': invoice_id}), 201

    except mysql.connector.Error as err:
        print(f"Database error during invoice submission: {err}")
        if cnx:
            cnx.rollback() # Rollback on error
        return jsonify({'message': f'Database error: {err}'}), 500
    except Exception as e:
        print(f"General error during invoice submission: {e}")
        if cnx:
            cnx.rollback()
        return jsonify({'message': f'An unexpected error occurred: {e}'}), 500
    finally:
        close_db_connection(cnx, cursor)

# @app.route('/api/data', methods=['GET'])
# def get_data():
#     return jsonify({'message': 'Hello from Flask!'})

if __name__ == '__main__':
    app.run(debug=True)