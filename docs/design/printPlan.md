# Print Plan

## Overview
On the summary page, there will be a print button. Once the print button is pressed, the data for that pay period would be sent over to my home server, and it would print the worker hours for that pay period there.

## What I need here
- A print button
- To send data to the home server in some format, like JSON, there will be a listener there that once received, will build a pdf. 
    - To generate this JSON I want a separate binary, that takes the pay period, the IP and Port of the home server as arguments, gets the information for the active workers with hours in that pay period, and makes the JSON and sends it over. 
    
## Home Server information
- The IP is: 10.9.0.7 (it is connected to our wireguard instance)
- It will be listening on the port: 5555
- THis info should be in a config file

## Listener
- Build the listener too
- In scripts we have build_schedule.py, which builds a pdf with empty values in the cells
- Change the scripts directory to schedule_listener. 
- Here should have build_schedule.py (with the addition of being able to have filled in values), a config file, a go server that will listen on a port for the incoming pay period data and send it to build schedule, a .service file to run this server.
- There, build the listener that would listen on a certain port (5555) take the JSON it gets from there, parses it, and builds the pdf schedule with the values filled in, with the format from build_schedule.py. 
- The final thing would be the printing, that can be done later on

