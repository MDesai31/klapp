# TO DO

## High Priority
- ~~Bug: A worker, Thomas, had forgotten to punch on Friday to saturday. SO there were hours marked on saturday. SO when I edited saturday hours, to make it Friday, it still remains as saturday.~~ Fixed: AdminUpdate now recalculates day/pay_period from the new start time.
- ~~Add button to be able to delete or add a time sheet entry.~~ Done: delete button per row; "+ Add entry" links to /admin/punches/new.
- ~~Auto punch-out any worker still clocked in at 9pm. Mark those timecards as non-compliant (new flag on the punch record). Non-compliant entries should appear at the top of the Timesheet tab, visually distinct from late entries.~~ Done: migration 0009, non_compliant column, nightly goroutine, red rows at top of Timesheet.

### msmtp setup (for invoice email to mylawncut@aol.com)
The invoice "Submit" button calls `msmtp --account=default mylawncut@aol.com`.
Steps to wire it up on the server:
1. `sudo apt install msmtp` (or `brew install msmtp` on Mac)
2. Create `~/.msmtprc` (mode 600) for the user that runs the service:
   ```
   defaults
   auth           on
   tls            on
   tls_trust_file /etc/ssl/certs/ca-certificates.crt
   logfile        ~/.msmtp.log

   account        default
   host           smtp.aol.com
   port           587
   from           mylawncut@aol.com
   user           mylawncut@aol.com
   password       <AOL app password>
   ```
   AOL requires an app-specific password if 2FA is on — generate one at
   https://login.aol.com/account/security.
3. Test: `echo -e "To: mylawncut@aol.com\nSubject: test\n\nhello" | msmtp -t`
4. The service user (e.g. `klapp`) needs to own the `~/.msmtprc` file.
   If running under a system user with no home dir, set
   `MSMTP_CONFIG=/etc/msmtprc` and update the path in the systemd unit.
5. Restart `klapp.service` once msmtp is verified working.

## Medium Priority
- Text notification for workers who have punched in to punch out at 5pm OR if the admin decides to manually send the notification. If the notification was sent manually by the admin. with a link and maybe even their PIN.

## Low priority
- Separate time reporting (Timesheet, Summary) from the admin dashboard into its own site or section, so invoicing and time-keeping concerns are more clearly divided.
- A tab with Map visualization of where workers punched out. Like there would be markers that when you click on the marker it would say the name of the worker. And where they punched out.
- Email summary of hours each worker worked that day

## Nice to have
- Optimize the CLAUDE.md file. So there needs less context each time
- Flag the time cards where location was not within a certain radius of certain locations of interest. 
    - This could be more complicated, as a condition I would want is if a worker punched out, outside of Mary
