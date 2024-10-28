# mailinwhite
open source anti spam with captcha challenge

## how it work?

this a postfix relay, you have to configure a domain MX to this server and set another smtp for delivery, the app will maintain a list of known contacts with a "human" status and require a captcha resolution from the sender to proceed for delivery.

before:

MX -> SMTP server


after:

MX -> mailinwhite -> SMTP server




mailinwhite process a new mail with unknow sender:

* the mail is stored in redis
* a link is sent by mail to the sender
* the sender submit captcha, then the mail is delivered

mailinwhite process a new mail with existing sender:

* if the sender has already validated a captcha, the mail is delivered to the next smtp server
* if the sender has not yet validated the captcha, a new link is sent by mail to the sender, all the pending emails will be sent upon validation


## Developpment

This project use docker compose and make, to start use:

```bash
make install

make build

make run

```

open the fake smtp receiver on url http://localhost:8081

send a fake mail:

```bash
make test_mail
```
