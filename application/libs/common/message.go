package common

import "strings"

// recipients list is extracted from a postfix command line
// reused as is by sendmail if recipient address already approved
// the list may contain -- before the recipients, filter it out
func GetValidRecipients(recipients []string) []string {
	var output []string
	for _, r := range recipients {
		if strings.Contains(r, "@") {
			output = append(output, r)
		}
	}

	return output
}
