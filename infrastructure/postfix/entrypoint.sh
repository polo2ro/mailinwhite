#!/bin/sh

update_main_cf() {
    config_file="/etc/postfix/main.cf"
    key="$1"
    value="$2"
    if grep -q "^$key" "$config_file"; then
        sed -i "s|^$key.*|$key = $value|" "$config_file"
    else
        echo "$key = $value" >> "$config_file"
    fi
}

newaliases # generates a /etc/postfix/aliases.db

echo "${SMTP_HOST}:${SMTP_PORT} ${SMTP_LOGIN}:${SMTP_PASSWORD}" > /etc/postfix/sasl_passwd
postmap /etc/postfix/sasl_passwd
chmod 600 /etc/postfix/sasl_passwd*
update_main_cf smtp_sasl_password_maps lmdb:/etc/postfix/sasl_passwd
update_main_cf myhostname "${MYHOSTNAME}"
update_main_cf relayhost "${SMTP_HOST}:${SMTP_PORT}"
update_main_cf smtp_sasl_auth_enable yes
update_main_cf smtp_tls_security_level "${SECURITY_LEVEL}"
update_main_cf smtp_sasl_security_options noanonymous
update_main_cf maillog_file /dev/stdout
update_main_cf smtpd_sasl_local_domain "\$myhostname"
update_main_cf smtpd_recipient_restrictions "permit_mynetworks,permit_sasl_authenticated,reject_unauth_destination"
update_main_cf mynetworks "127.0.0.0/8, 172.20.0.0/16"

export -p > /home/filter/.env && chown filter:filter /home/filter/.env

/usr/sbin/postfix -c /etc/postfix start-fg

