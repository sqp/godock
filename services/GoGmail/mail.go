package GoGmail

import (
	"github.com/sqp/godock/libs/net/download"

	"encoding/base64"
	"errors"
	"io/ioutil"
	"strings"
)

//---------------------------------------------------------[ MAIL INTERFACES ]--

// Mailbox is a mail client interface.
//
type Mailbox interface {
	// Timer management.
	Check()

	// Mail management.
	IsValid() bool
	Count() (nbInbox int)
	Clear()
	LoadLogin(filepath string)
	SaveLogin(login string)
}

// RendererMail is a display interface to show inbox mail count on the icon.
//
type RendererMail interface {
	Set(count int) // Set new value.
	Error(e error) // Set error.
}

//--------------------------------------------------------[ GMAIL CONNECTION ]--

// Email is a single email data. Disabled fields are just examples of what is
// supposed to be parsed if you want to use them.
//
type Email struct {
	Title       string `xml:"title"`
	Summary     string `xml:"summary"`
	Modified    string `xml:"modified"`
	Issued      string `xml:"issued"`
	AuthorName  string `xml:"author>name"`
	AuthorEmail string `xml:"author>email"`
	//~ <link rel="alternate" href="http://mail.google.com/mail?account_id=###&extsrc=atom" type="text/html"/>
	//~ <id>tag:gmail.google.com,204:14257</id>
}

// Feed contains Gmail inbox data. Some fields are disabled because they are
// unused. They could be enabled simply by uncommenting them.
//
type Feed struct {
	Title   string `xml:"title"`
	Tagline string `xml:"tagline"`
	Total   int    `xml:"fullcount"`
	//~ Link  string   `xml:"href,attr"`
	//~ Modified string   `xml:"modified"`
	Mail []*Email `xml:"entry"`

	// Display fields.
	New      int
	Plural   bool
	MailsNew []*Email

	// Mail polling data.
	login      string                 // Login informations.
	file       string                 // Login file location.
	restart    chan bool              // restart channel to forward user requests.
	callResult func(int, bool, error) // Action to execute to send polling results.
}

// NewFeed create a new Gmail inbox feed.
//
func NewFeed(onResult func(int, bool, error)) *Feed {
	return &Feed{callResult: onResult}
}

// Count return the number of unread mails.
//
func (feed *Feed) Count() int {
	return feed.Total
}

// Clear reset the mail list.
//
func (feed *Feed) Clear() {
	feed.Mail = nil
	feed.Total = 0
}

// IsValid return true is user informations were provided.
// Only tells if login and password were submitted, not if they are valid.
//
func (feed *Feed) IsValid() bool {
	return feed.login != ""
}

// Check callback for poller mail checking. Launch the check mail action.
// Check for new mails and launch the result callback with the mails count delta
// (change since last check).
//
func (feed *Feed) Check() {
	if !feed.IsValid() {
		feed.callResult(0, false, errors.New("no account informations provided"))
		return
	}

	count := feed.Count() // save current count.
	feed.Clear()          // reset list.

	// Get new data.
	source := download.Header{"Authorization": "Basic " + feed.login}
	e := source.XML(feedGmail, feed)

	feed.callResult(feed.Count()-count, count == 0, e)
}

// LoadLogin get user login information from file.
//
func (feed *Feed) LoadLogin(filename string) {
	feed.file = filename
	b, err := ioutil.ReadFile(feed.file)
	if err == nil {
		t, e2 := base64.StdEncoding.DecodeString(string(b))
		if e2 == nil {
			split := strings.Split(string(t), "\n")
			feed.login = base64.StdEncoding.EncodeToString([]byte(split[0] + ":" + split[1]))
		}
	}
}

// SaveLogin login informations to file with the same format as the Gmail applet.
//
func (feed *Feed) SaveLogin(login string) {
	if login == "" {
		return
	}
	str := []byte(strings.Replace(login, ":", "\n", 1))
	coded := base64.StdEncoding.EncodeToString(str)
	ioutil.WriteFile(feed.file, []byte(coded), 0600)
	feed.LoadLogin(feed.file)
}
