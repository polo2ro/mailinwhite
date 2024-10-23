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
update_main_cf smtp_sasl_password_maps hash:/etc/postfix/sasl_passwd

update_main_cf relayhost "${SMTP_HOST}:${SMTP_PORT}"
update_main_cf smtp_sasl_auth_enable yes
update_main_cf smtp_tls_security_level enrypt
update_main_cf smtp_sasl_security_options noanonymous
update_main_cf maillog_file /dev/stdout

/usr/sbin/postfix -c /etc/postfix start-fg

