package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-ldap/ldap/v3"
	"github.com/gorilla/mux"
)

func AcceptHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{"message": "Accept"}
	json.NewEncoder(w).Encode(response)
}

func getContactHandler(w http.ResponseWriter, r *http.Request) {
	uid := mux.Vars(r)["uid"]

	l, err := ldap.DialURL("ldap://openldap:1389")
	if err != nil {
		http.Error(w, "LDAP connexion error", http.StatusInternalServerError)
		return
	}
	defer l.Close()

	searchRequest := ldap.NewSearchRequest(
		"dc=example,dc=com",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(objectGUID="+uid+")",
		[]string{"cn", "mail", "telephoneNumber"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		http.Error(w, "LDAP search error", http.StatusInternalServerError)
		return
	}

	if len(sr.Entries) != 1 {
		http.Error(w, "contact not found", http.StatusNotFound)
		return
	}

	contact := map[string]string{
		"name":  sr.Entries[0].GetAttributeValue("cn"),
		"email": sr.Entries[0].GetAttributeValue("mail"),
		"phone": sr.Entries[0].GetAttributeValue("telephoneNumber"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contact)
}
