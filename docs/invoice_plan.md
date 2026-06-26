# Invoice Plan

## Overview
Create invoices. Store in db. Send to quickbooks once aproved by admin.

## Additions
- New invoice site in a different port. 8083. Also internal only site, only accessible through the VPN
- Invoice table, columns: date, submited_by, house_number, customer_name, time_arrived, time_left, no_of_workers, job_descriptions, materials_used, comments, reviewd.
- Customer trable. Customer columns: id, name, phone number, house number, address
- Job_description table: Just serve as a list of possible jobs done
- materials table: List of materials, columns: material, price, quantity, unit
- Invoice tab to admin site.
    - job descriptions page in invoice tab.
- Customers tab to admin site 

## Invoice site
Can have a separate binary if that is better. We must then update the ./deploy scripts. Let me know if you think it is better, or else we could do all on the same binary. But I would like there to be some separation between this code and the time keeping code. Different directories and such, even if it gets decided that same binary is better.
Should be served on port 8083. Only accessible via wireguard. 

### The Form
Should be a simple form. When accessing the website, the worker is met with entering a PIN, same as to punch in or out, this would identify the submited_by column and a simple login. Depending on the PIN entered by the worker, the page should be displayed in english or in spanish. The only required fields are: date, house_number, number_of_workers, time_arrived, time_left
- date: Basic date picker thing, defaults on today, can change it.
- house_number, can type it, only numbers 
- customer_name: Once house number is typed, this should be populated with the name of the customer
- number of workers: Just an int
- time_arrived
- time_left
- job description: This should be 3 long horizontal text boxes, with a plus sign at the bottom to be able to add more horizontal text boxes. Once you start typing, options show up underneath to potentially fill them in with job_descriptions already in the job decsription table
- materials used: same as job description but for materials used.
- comments: Just a large text box where you can type comments.

## Admin dashboard

### Invoice tab
Admin dashboard should have an invoice tab. Every submitted invoice must be reviewed. Every invoice submitted should be there (of course, have it so you have to go to different pages, we dont want to load hundreds of invoices...). If not reviewed, they should be highlighted yellow. They should be horizontal rows, displaying the date, customer name, and a button to view. Once view is clicked, should be taken to the page just like the invoice form page, except, with two buttons at the bottom, SUBMIT and CANCEL. Submit can be grayed out once submited the first time, and marked as reviewed on the database. The submit button should send it as an email to "mylawncut@aol.com" and to quickbooks as described on the next section.

### Customers tab
List all the customers. Can search them, can add new ones. When you click on one, shows the invoices linked to this customer.


## Quickbooks integration
Once reviewd, an invoice will go to quickbooks. This could use the quickbooks API. Let's plan this sections together to be implemented at a later date



