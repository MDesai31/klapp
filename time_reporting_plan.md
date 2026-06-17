# Time Reporting Design

## Objectives

- Worker can punch in and punch out
- Keep track of who has punched in and punched out

## Worker Interface

- This will be accessed on web browser on phone over open internet
- Enter PIN given to worker associated with them
- It will display: Hello WorkerName
  - If the worker has not punched in:
    - it will display a large green button to punch in
  - If the worker has already punched in or the worker presses the punch in button:
    - it will display the time the worker has punched in
    - It will display a large red button to punch out
- When accessing this page, the browser will prompt the user to allow location data to be collected.
  - We will keep track of location data when punched in and out
  - If the user punches in far from storage area, it might be flagged.


## Administrator Interface

This page is for the administrator to monitor the time reporting of the workers and make changes to them if necessary.

- This webpage will only be accessible from home network or VPN if on the field. From phone or PC.
- This interface will show punch in and punch out times of every worker
- The administrator can edit the times of workers here
- At the top of the screen the admin will see the current workers that have punched in and punched out for the day
- Scroll a little down and can see punch in and punch out times for every worker this pay period
- A page with all the workers and their PINs can be found, if a new worker comes in there would also be an option to add another worker with a PIN and all.
- A link to another page with previous pay periods can also be accessed.

Since the administrator can easily see who has or hasn't punched in, they can easily message someone for forgetting to punch in or out. Make corrections. Evaluate if the expected hours are met. Certain time reporting can be flagged if unusual. Such as if working too many hours, or started work at an unexpected location.

## Worker Missed Punch In / Punch Out

Workers failing to punch in or out will likely be a common occurrence.

### If a Worker Forgets to PUNCH IN

There are two approaches I can think of to handle this.

- **Option 1**: They can punch in whenever they remember and have to ask the administrator to enter the correct time by text or talking:
  - This might disincentivise the worker from slacking on punching in, since it will be annoying to have to keep asking the administrator to enter the time. They can be lightly scolded each time, and therefore, remember to enter the time.
  - If a worker is a repeat offender, a threat of docking the pay can be made when scolding the worker.
- **Option 2**: They can themselves enter the correct start time on the webpage
  - This might be easier to the administrator, since they do not have to do anything to correct the mistake, however, it also incentivises not punching in, and just correcting the time later.
- **Recommendation**:
  - Option 1: Since I do believe option 2 would devolve into nobody punching in, and just fixing the time later, as it is the path of least resistance. And it is what I would do.


### If a Worker Forgets to PUNCH OUT

For every worker that has punched in, but not punched out:

- A reminder can go out for all of those who have punched in around 6pm to remember to punch out with the link.
  - The administrator can send out this reminder at will when the day is over.
- After a certain time, let's say 9pm, the people who punched in, but did not punch out get texted a reminder they have not punched out, and the number of times they have not punched out in the past, and a link to the website with a box to enter the time they finished work. The text can also contain a threat that if the person does not enter the time they punched out, the day will not count. Another counter of number of times they haven't punched out will be displayed on the webpage as well.
  - The administrator can also trigger this at will.

The simplicity of going to the link and pressing the red button should maybe incentivise people to not forget. With the reminder with a link that all you have to do is press the red button is also easy to access and use. Hopefully the threat of docking the pay can have people remembering to do so.

## Issues

- If worker has bad internet for the time and can't connect to the website:
  - Just text the administrator the times, this is unlikely to happen often.
- If website is down:
  - A second server monitoring the first would notify the administrator that it is down, and the administrator can notify the workers.

## Integration with Current Work

Workers can keep doing what they are doing with the physical paper.

- Pay Period 1: First the owner starts punching in and punching out to see if the system is working and if there are any improvements necessary.
- Pay Period 2-3: Get a few trusted workers to use the new time reporting system. See if any issues arise and make changes if necessary.
- Pay Period 4: Rollout to everyone else.

## Technical Aspect

- External/Worker Website:
  - Should be https
  - Can be region locked to the USA
  - The only connection to backend should be when:
    - the PIN is entered, the backend should identify which worker is using the website
    - the punch in / punch out button is pushed, provide backend with: employee, datetime, punch in/out location
  - Worries: How safe is this? How to make it safer without having a complex log in and log out.

- Internal/Administrator Website:
  - Will not be accessible to outside internet, only through Wireguard. I do not believe security will be much of a concern.
  - Can view and edit the current pay period hours.

- Backend:
  - Websites will be hosted with Go programming language. For external access, Caddy sits in front as the reverse proxy (automatic HTTPS).
  - There will be a SQLite database with tables:
    - **worker**: To keep worker\_name, pin
    - **time\_reporting**: Columns: employee\_name, pay\_period (start date of the pay period), day, start\_time, end\_time, start\_lat, start\_lon, end\_lat, end\_lon, late (boolean with true if was entered late), modified\_by\_admin (boolean for true if this entry was modified by admin)
  - The server will have a Telit modem connected so it is able to send text notification through AT commands.
  - The backend will handle the data processing and requests from admin and workers.
  - At the end of the pay period, the backend will email the admin the times of each worker, create PDF reports for each, print them out, and eventually, create some data that can be imported into QuickBooks.
