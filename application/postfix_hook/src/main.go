package main

// receive mail from postfix
// if contact is approved in directory, forward the mail
// if contact does not exists or is not approved, send a challenge and postpone the mail in a queue

import (
	"fmt"
	"log"
	"os"

	"github.com/go-ldap/ldap/v3"
)

const (
	ldapServer   = "ldap://localhost:389"
	ldapBindDN   = "cn=admin,dc=example,dc=com"
	ldapPassword = "adminpassword"
	ldapBaseDN   = "ou=users,dc=example,dc=com"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: ./check_email <sender_email>")
	}
	senderEmail := os.Args[1]

	l, err := ldap.DialURL(ldapServer)
	if err != nil {
		log.Fatalf("LDAP serveur connexion error : %v", err)
	}
	defer l.Close()

	err = l.Bind(ldapBindDN, ldapPassword)
	if err != nil {
		log.Fatalf("LDAP authentication error: %v", err)
	}

	searchRequest := ldap.NewSearchRequest(
		ldapBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(mail=%s)", senderEmail),
		[]string{"dn"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Fatalf("LDAP search error: %v", err)
	}

	if len(sr.Entries) == 0 {
		fmt.Println("mail address not found")
		os.Exit(1)
	} else {
		fmt.Println("ok")
		os.Exit(0)
	}
}
