package GoGmail

import (
	"github.com/sqp/godock/libs/log" // Display info in terminal.
	"io/ioutil"

	"encoding/base64"
	"encoding/xml"
	"errors"
	"net/http"
	"strings"
)

//---------------------------------------------------------[ MAIL INTERFACES ]--

type Mailbox interface {
	// Timer management.
	Check()

	// Mail management.
	IsValid() bool
	Count() (nbInbox int)
	Clear()
	Data() interface{}
	LoadLogin(filepath string)
	SaveLogin(login string)
}

// Mail rendering interface. Displays inbox mail count on the icon.
//
type RendererMail interface {
	Set(count int) // Set new value.
	Error(e error) // Set error.
}

//--------------------------------------------------------[ GMAIL CONNECTION ]--

// Single Email data. Disabled fields are just examples of what is supposed to
// be parsed if you want to use them.
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

// Gmail inbox data. Some fields are disabled only because they are unused. They
// could be enabled simply by removing the comment in front of them.
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

func NewFeed(onResult func(int, bool, error)) *Feed {
	return &Feed{callResult: onResult}
}

// Get number of unread mails.
//
func (feed *Feed) Count() int {
	//~ return len(feed.Mail)
	return feed.Total
}

// Get feed original data. You will have to recast to *Feed.
//
func (feed *Feed) Data() interface{} {
	return feed
}

// Reset mail list.
//
func (feed *Feed) Clear() {
	feed.Mail = nil
	feed.Total = 0
}

// IsValid return true is the user provided informations. Only tells if login
// and password were submitted, not if they are valid.
//
func (feed *Feed) IsValid() bool {
	return feed.login != ""
}

// callback for poller mail checking. Launch the check mail action.
// Check for new mails. Return the mails count delta (change since last check).
//
func (feed *Feed) Check() {
	log.Debug("Check mails")
	if !feed.IsValid() {
		feed.callResult(0, false, errors.New("No account informations provided."))
		return
	}

	count := feed.Count() // save current count.
	feed.Clear()          // reset list.
	e := feed.get()       // Get new data.

	feed.callResult(feed.Count()-count, count == 0, e)
}

// Get mails from server.
//
func (feed *Feed) get() error {
	log.Debug("Get mails from server")

	// Prepare request with header
	request, e := http.NewRequest("GET", feedGmail, nil)
	if e != nil {
		return e
	}
	request.Header.Add("Authorization", "Basic "+feed.login)

	// Try to get data from source.
	response, e2 := new(http.Client).Do(request)
	if e2 != nil {
		return e2
	}
	defer response.Body.Close()

	body, e3 := ioutil.ReadAll(response.Body)
	if e3 != nil {
		return e3
	}

	// Parse data.
	if e := xml.Unmarshal(body, feed); e != nil {
		return e
	}
	return nil
}

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

// Save login informations in the same format as the Gmail applet.
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
