package uptoshare_test

import (
	"github.com/stretchr/testify/assert"

	"github.com/sqp/godock/libs/log"
	"github.com/sqp/godock/libs/net/uptoshare"

	"fmt"
	"io/ioutil"
	"testing"
)

const (
	tmpHistory = "test_history.crap"

	imageToUpload = "/usr/share/cairo-dock/images/cairo-dock-logo.png"
	// imageToUpload = "/usr/share/cairo-dock/cairo-dock.svg"
	textToUpload = "/usr/share/cairo-dock/ChangeLog.txt"
	// textToUpload = "this is a test"
)

var testUploadText = []string{
// "Codepad.org",
// "Pastebin.com",
// "Pastebin.mozilla.org",
// "Paste-ubuntu.com",
// "Play.golang.org",
}

var testUploadImage = []string{
// "Imagebam.com",
// "Imagebin.ca",
// "ImageShack.us",
// "Imgclick.net",
// "Imgland.net",
// "Imgur.com",           // FAIL
// "Postimage.org",
}

var testUploadVideo = []string{
// "VideoBin.org",
}

var testUploadFile = []string{
// "Filebin.ca",
// "Freemov.top",
// "Leopard.hosting",
// "Pixeldra.in",
// "Transfer.sh",
}

func newUpclt() (*uptoshare.Uploader, error) {
	tmp, e := ioutil.TempFile("", tmpHistory+".")
	if e != nil {
		return nil, e
	}
	// Uptoshare actions
	up := uptoshare.New()
	up.Log = log.NewLog(log.Logs).SetName("uptoshare_test")
	// up.SetPreCheck(func() {  })
	// up.SetPostCheck(func() {  })

	// Uptoshare settings.
	up.SetHistoryFile(tmp.Name())
	up.SetHistorySize(42)
	// up.LimitRate = app.conf.UploadRateLimit
	// up.PostAnonymous = app.conf.PostAnonymous

	return up, nil
}

func onResult(t *testing.T) func(uptoshare.Links) {
	return func(links uptoshare.Links) {
		if !assert.NotEmpty(t, links, "links ") {
			return
		}
		e, ok := links["error"]
		assert.False(t, ok, "upload error:"+e)

		fmt.Println(links["link"])
		link, ok := links["link"]
		assert.True(t, ok, "links[link] missing")
		assert.NotEmpty(t, link, "links[link]")

		fmt.Println("links", links)
	}
}

func TestUpload(t *testing.T) {
	up, e := newUpclt()
	if !assert.NoError(t, e, "temp file") {
		return
	}

	up.SetOnResult(onResult(t))

	for _, site := range testUploadText {
		up.SiteText(site)
		up.UploadGuess(textToUpload)
	}

	for _, site := range testUploadImage {
		up.SiteImage(site)
		up.UploadGuess(imageToUpload)
	}

	up.FileForAll = true
	for _, site := range testUploadFile {
		up.SiteFile(site)
		up.UploadGuess(imageToUpload)
	}
}

func TestLeopard(t *testing.T) {
	str := `{"upload":{"support":"consider making a donation","downloadURL":"http:\/\/leopard.hosting\/download.php?f=muivh","fileCode":"muivh","deleteKey":"hdpkshtnipjkmrlqgdqjfbwznsliag","deleteURL":"http:\/\/leopard.hosting\/delete.php?f=muivh&key=hdpkshtnipjkmrlqgdqjfbwznsliag"}}`
	links := uptoshare.LeopardHosting.Parse(uptoshare.LeopardHosting, str)
	assert.NotEmpty(t, links, "LeopardHostingAnswer")
}
