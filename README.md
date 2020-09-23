# onesecmail
onesecmail is a Go library for accessing the www.1secmail.com API.

## Usage
The GoDoc for this module is at: https://pkg.go.dev/github.com/z11i/onesecmail

```go
package main

import (
    "fmt"
    
    "github.com/z11i/onesecmail"
)

func main() {
    // Generate a random mailbox name
    // The domain name can be either 1secmail.org, 1secmail.com, or 1secmail.net
    mailboxName := "randomname@1secmail.org"
    
    // Send emails to the generated email
    // ...
    
    // Create a mailbox struct for checking 1secmail
    mailbox := onesecmail.NewMailbox("randomname", "1secmail.org", nil)
    // mailbox.Address() == mailboxName
    
    // Check inbox
    mails, err := mailbox.CheckInbox()
    if err != nil {
        // handle err
    }
    
    // Read messages
    for _, mail := range mails {
        fmt.Printf("Received mail from %s with subject %s on %s\n", mail.From, mail.Subject, mail.Date)
    }
}
```
