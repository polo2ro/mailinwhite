package main

// receive mail from postfix
// if contact is approved in directory, forward the mail
// if contact does not exists or is not approved, send a challenge and postpone the mail in a queue

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/go-ldap/ldap/v3"
)

const (
	ldapServer   = "ldap://localhost:389"
	ldapBindDN   = "cn=admin,dc=example,dc=com"
	ldapPassword = "adminpassword"
	ldapBaseDN   = "ou=users,dc=example,dc=com"

	inspectDir = "/home/filter/spool"

	// Postfix exit codes
	EX_OK          = 0  // Successful termination
	EX_TEMPFAIL    = 75 // Temporary failure
	EX_UNAVAILABLE = 69 // Service unavailable
	EX_USAGE       = 64 // Command line usage error
)

// postfix mail filter
// https://www.postfix.org/FILTER_README.html
func filterContent(senderEmail string) error {
	l, err := ldap.DialURL(ldapServer)
	if err != nil {
		return fmt.Errorf("LDAP serveur connexion error : %w", err)
	}
	defer l.Close()

	err = l.Bind(ldapBindDN, ldapPassword)
	if err != nil {
		return fmt.Errorf("LDAP authentication error: %v", err)
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
		return fmt.Errorf("LDAP search error: %v", err)
	}

	if len(sr.Entries) == 0 {
		return errors.New("mail address not found")
	} else {
		return nil
	}
}

func sendTmpFile(tmpFile *os.File, from string) error {
	args := append([]string{"-G", "-i", "-f", from}, os.Args[3:]...)
	cmd := exec.Command("/usr/sbin/sendmail", args...)
	var err error
	cmd.Stdin, err = os.Open(tmpFile.Name())
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running sendmail: %w", err)
	}

	return nil
}

func main() {
	if len(os.Args) < 4 || os.Args[1] != "-f" {
		fmt.Fprintln(os.Stderr, "Usage: script -f sender recipients...")
		os.Exit(EX_USAGE)
	}

	if err := os.Chdir(inspectDir); err != nil {
		fmt.Fprintf(os.Stderr, "%s does not exist\n", inspectDir)
		os.Exit(EX_TEMPFAIL)
	}

	// Create a temporary file
	tmpFile, err := os.CreateTemp(inspectDir, "in.")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot create temporary file")
		os.Exit(EX_TEMPFAIL)
	}
	defer os.Remove(tmpFile.Name())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove(tmpFile.Name())
		os.Exit(1)
	}()

	// Copy standard input to the temporary file
	if _, err := io.Copy(tmpFile, os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, "Cannot save mail to file")
		os.Exit(EX_TEMPFAIL)
	}

	tmpFile.Close()

	from := os.Args[2]

	// Ici, vous pouvez ajouter votre logique de filtrage personnalisée
	// Par exemple :
	if err := filterContent(from); err != nil {
		log.Println(err)
		// fmt.Fprintf(os.Stderr, "Message rejected: %s\n", err)
		// os.Exit(EX_UNAVAILABLE)
	}

	if err := sendTmpFile(tmpFile, from); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(EX_TEMPFAIL)
	}

	os.Exit(EX_OK)
}
