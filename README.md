# mailinwhite

open source anti-spam with CAPTCHA challenge

## How it works

This is a Postfix relay. You have to configure a domain MX to this server and set another SMTP for delivery. The app will maintain a list of known contacts with a "human" status and require a CAPTCHA resolution from the sender to proceed for delivery.

Before:

MX -> SMTP server

After:

MX -> mailinwhite -> SMTP server

mailinwhite processes a new mail with unknown sender:

* The mail is stored in Redis
* A link is sent by mail to the sender
* The sender submits CAPTCHA, then the mail is delivered

mailinwhite processes a new mail with existing sender:

* If the sender has already validated a CAPTCHA, the mail is delivered to the next SMTP server
* If the sender has not yet validated the CAPTCHA, a new link is sent by mail to the sender; all the pending emails will be sent upon validation

## Development

This project uses Docker Compose and Make. To start, use:

```bash
make install

make build

make run
```

Open the fake SMTP receiver on URL http://localhost:8081 

Send a fake mail:

```bash
make test_mail
```
