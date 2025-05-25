CREATE SCHEMA `klapp`;
USE `klapp`;

CREATE TABLE users (
    user_id INTEGER PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    `password` VARCHAR(255) NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE
);

INSERT INTO users (username, `password`, is_admin)
VALUES
    ('admin', 'admin123', TRUE),
    ('manth', 'manth123', FALSE),
    ('thom', 'thom123', FALSE);

CREATE TABLE employee (
    emp_id INTEGER PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(50) NOT NULL,
    user_id INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users (user_id)
);

INSERT INTO employee (`name`, user_id)
VALUES
    ('Manthan', 2),
    ('Thomas', 3);

CREATE TABLE customer (
    customer_id INTEGER PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(50) NOT NULL
);

INSERT INTO customer (name) VALUES
    ('Customer 1'),
    ('Customer 2'),
    ('Customer 3');

CREATE TABLE materials (
    material_id INTEGER PRIMARY KEY AUTO_INCREMENT,
    `description` VARCHAR(100) NOT NULL
);

INSERT INTO materials (description) VALUES
    ('Topsoil'),
    ('Mulch (Brown)'),
    ('Mulch (Red)'),
    ('Gravel');

CREATE TABLE jobs (
    job_id INTEGER PRIMARY KEY AUTO_INCREMENT,
    `description` VARCHAR(100) NOT NULL
);

INSERT INTO jobs (description) VALUES
    ('Lawn Mowing'),
    ('Weeding'),
    ('Tree Pruning'),
    ('Hedge Trimming'),
    ('Garden Bed Installation');

CREATE TABLE invoice (
    invoice_id INTEGER PRIMARY KEY AUTO_INCREMENT,
    customer_id INTEGER NOT NULL,
    `date` DATE NOT NULL,
    time_arrived TIME NOT NULL,
    time_left TIME NOT NULL,
    num_workers INTEGER NOT NULL,
    comments VARCHAR (200),
    FOREIGN KEY (customer_id) REFERENCES customer (customer_id)
);

CREATE TABLE invoice_workers (
    invoice_id INTEGER,
    emp_id INTEGER,
    PRIMARY KEY (invoice_id,emp_id),
    FOREIGN KEY (invoice_id) REFERENCES invoice (invoice_id),
    FOREIGN KEY (emp_id) REFERENCES employee (emp_id)
);

CREATE TABLE invoice_materials (
   invoice_id INTEGER,
   material_id INTEGER,
   PRIMARY KEY (invoice_id,material_id),
   FOREIGN KEY (invoice_id) REFERENCES invoice (invoice_id),
   FOREIGN KEY (material_id) REFERENCES materials (material_id)
);

CREATE TABLE invoice_jobs (
   invoice_id INTEGER,
   job_id INTEGER,
   PRIMARY KEY (invoice_id,job_id),
   FOREIGN KEY (invoice_id) REFERENCES invoice (invoice_id),
   FOREIGN KEY (job_id) REFERENCES jobs (job_id)
);