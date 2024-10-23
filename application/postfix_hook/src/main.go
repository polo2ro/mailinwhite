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
)

const (
	inspectDir = "/var/spool/filter"
	sendmail   = "/usr/sbin/sendmail"
)

// postfix mail filter
// https://www.postfix.org/FILTER_README.html
func filterContent(senderEmail string) error {
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
		return errors.New("mail address not found")
	} else {
		return nil
	}
}

func main() {
	// Vérifier les arguments
	if len(os.Args) < 4 || os.Args[1] != "-f" {
		fmt.Println("Usage: script -f sender recipients...")
		os.Exit(1)
	}

	// Changer le répertoire de travail
	if err := os.Chdir(inspectDir); err != nil {
		fmt.Printf("%s does not exist\n", inspectDir)
		os.Exit(75) // EX_TEMPFAIL
	}

	// Créer un fichier temporaire
	tmpFile, err := os.CreateTemp(inspectDir, "in.")
	if err != nil {
		fmt.Println("Cannot create temporary file")
		os.Exit(75) // EX_TEMPFAIL
	}
	defer os.Remove(tmpFile.Name())

	// Configurer le nettoyage en cas d'interruption
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove(tmpFile.Name())
		os.Exit(1)
	}()

	// Copier l'entrée standard dans le fichier temporaire
	if _, err := io.Copy(tmpFile, os.Stdin); err != nil {
		fmt.Println("Cannot save mail to file")
		os.Exit(75) // EX_TEMPFAIL
	}

	// Fermer le fichier pour s'assurer que tout est écrit
	tmpFile.Close()

	// Ici, vous pouvez ajouter votre logique de filtrage personnalisée
	// Par exemple :
	if err := filterContent(os.Args[1]); err != nil {
		fmt.Println("Message content rejected")
		os.Exit(69) // EX_UNAVAILABLE
	}

	// Préparer les arguments pour sendmail
	args := append([]string{"-G", "-i"}, os.Args[2:]...)

	// Exécuter sendmail
	cmd := exec.Command(sendmail, args...)
	cmd.Stdin, err = os.Open(tmpFile.Name())
	if err != nil {
		fmt.Println("Cannot open temporary file")
		os.Exit(75) // EX_TEMPFAIL
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error running sendmail: %v\n", err)
		os.Exit(1)
	}
}
