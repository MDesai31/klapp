# SMS notification setup

`cmd/sms` is a small CLI tool that sends a one-off text message via Twilio.
It's not wired into the worker or admin sites yet — for now it's a manual
tool you run from the command line, locally or on the Lightsail box.

## 1. Create a Twilio account

Go to twilio.com and sign up. A trial account is fine to start.

Trial accounts can only send to phone numbers you've manually verified in
the console (Console → Phone Numbers → Verified Caller IDs). To text any
worker's number without verifying it first, add a payment method to
upgrade out of trial mode (~$20 minimum, no subscription).

## 2. Get a Twilio phone number

Console → Phone Numbers → Manage → Buy a number. Trial accounts get one
free number; paid accounts pay roughly $1.15/mo for the number plus
~$0.0079 per SMS sent (US pricing). Make sure the number has SMS
capability enabled.

## 3. Grab your credentials

On the Console dashboard home page:

- **Account SID**
- **Auth Token** (click "show" to reveal)

Also note your Twilio phone number in E.164 format, e.g. `+15551234567`.

## 4. Set environment variables

The tool reads three environment variables — nothing is hardcoded or
committed to the repo:

- `TWILIO_ACCOUNT_SID`
- `TWILIO_AUTH_TOKEN`
- `TWILIO_FROM_NUMBER` (your Twilio number, E.164 format)

### Local testing (your machine)

Export them in your shell before running the tool:

```sh
export TWILIO_ACCOUNT_SID=ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
export TWILIO_AUTH_TOKEN=your_auth_token
export TWILIO_FROM_NUMBER=+15551234567

go run ./cmd/sms -to +15557654321 -body "Test message"
```

### On the Lightsail box

Don't put real tokens in `klapp.service` or any file tracked by git.
Instead:

1. Create `/opt/klapp/sms.env` (outside the repo):

   ```
   TWILIO_ACCOUNT_SID=ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
   TWILIO_AUTH_TOKEN=your_auth_token
   TWILIO_FROM_NUMBER=+15551234567
   ```

2. Lock it down:

   ```sh
   chmod 600 /opt/klapp/sms.env
   ```

3. Build the binary once, same as `web` in `deploy/update.sh`:

   ```sh
   go build -o /opt/klapp/sms ./cmd/sms
   ```

4. Load the env file and run it manually when needed:

   ```sh
   set -a; source /opt/klapp/sms.env; set +a
   /opt/klapp/sms -to +15557654321 -body "Test message"
   ```

If this later gets wired into the always-running `klapp` service (5pm
auto-reminder, admin "send now" button), `sms.env` can be referenced
directly from `klapp.service` via `EnvironmentFile=/opt/klapp/sms.env`
instead of sourcing it by hand.

## 5. Try it

```sh
go run ./cmd/sms -to +1XXXXXXXXXX -body "Hello from klapp"
```

A successful send prints `sent to +1XXXXXXXXXX`. Errors from Twilio (bad
number, unverified destination on a trial account, missing env vars,
etc.) are printed as-is to stderr.
