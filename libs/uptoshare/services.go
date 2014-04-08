package uptoshare

import (
	"github.com/andelf/go-curl" // imported as curl

	"github.com/sqp/godock/libs/log" // Display info in terminal.
	"github.com/sqp/godock/libs/ternary"

	"io/ioutil"
	"net/url"
	"os/exec"
	"path"
	"strings"
)

//
//-----------------------------------------------------------[ TEXT BACKENDS ]--

//
//------------------------------------------------------------[ PASTEBIN.COM ]--

const (
	PasteBinComURL    = "http://pastebin.com/api/api_post.php"
	PasteBinComFormat = "text"                             // pastebin is only used for text.
	PasteBinComExpire = "1M"                               // 1 month
	PasteBinComKey    = "4dacb211338b25bfad20bc6d4358e555" // if you re-use this code, please make your own key !
	PasteBinComOption = "paste"
)

func PasteBinCom(text, cLocalDir string, bAnonymous bool, limitRate int) Links { //, gchar **cResultUrls, GError **pError)
	values := make(url.Values)
	values.Set("api_option", PasteBinComOption)
	values.Set("api_user_key", "")
	values.Set("api_paste_private", "1") // bAnonymous ? "1" : "0", // unlisted or public
	// values.Set("api_paste_name", ) // bAnonymous ? "" : g_getenv ("USER"),
	values.Set("api_paste_expire_date", PasteBinComExpire)
	values.Set("api_paste_format", PasteBinComFormat)
	values.Set("api_dev_key", PasteBinComKey)

	values.Set("api_paste_code", text)

	link, e := postSimple(PasteBinComURL, values)
	if e != nil {
		return linkErr(e, "Pastebin")
	}

	if !strings.HasPrefix(link, "http://") {
		return linkWarn("Pastebin bad format: " + link)
	}

	return NewLinks().Add("link", link)
}

//
//--------------------------------------------------------[ PASTE.UBUNTU.COM ]--

const (
	PasteUbuntuComURL    = "http://paste.ubuntu.com"
	PasteUbuntuComFormat = "text"   // only used for text.
	PasteUbuntuComSubmit = "Paste!" // button.
)

func PasteUbuntuCom(text, cLocalDir string, bAnonymous bool, limitRate int) Links {
	values := make(url.Values)
	values.Set("syntax", PasteUbuntuComFormat)
	values.Set("submit", PasteUbuntuComSubmit)
	values.Set("poster", "Anonymous") // bAnonymous ? "" : g_getenv ("USER"),

	values.Set("content", text)

	data, e := postSimple(PasteUbuntuComURL, values)
	if e != nil {
		return linkErr(e, "PasteUbuntuCom")
	}

	ID := findPrefix(data, `class="pturl" href="/`, "/plain/")
	if ID == "" {
		return linkWarn("PasteUbuntuCom: parse failed\n" + data)
	}

	return NewLinks().Add("link", PasteUbuntuComURL+"/"+ID)
}

//
//--------------------------------------------------------[ PASTE.UBUNTU.COM ]--

const (
	PasteBinMozillaOrgURL    = "http://pastebin.mozilla.org"
	PasteBinMozillaOrgFormat = "text" // only used for text.
	PasteBinMozillaOrgExpiry = "d"    // a day?
	PasteBinMozillaOrgSubmit = "Send" // button.
)

func PasteBinMozillaOrg(text, cLocalDir string, bAnonymous bool, limitRate int) Links {
	values := make(url.Values)
	values.Set("format", PasteBinMozillaOrgFormat)
	values.Set("paste", PasteBinMozillaOrgSubmit)
	values.Set("expiry", PasteBinMozillaOrgExpiry)
	values.Set("remember", "0")
	values.Set("parent_pid", "")
	values.Set("poster", "Anonymous") // bAnonymous ? "" : g_getenv ("USER"),

	values.Set("code2", text)

	data, e := postSimple(PasteBinMozillaOrgURL, values)
	if e != nil {
		return linkErr(e, "PasteBinMozillaOrg")
	}

	ID := findLink(data, "/?dl=", `"`)
	if ID == "" {
		return linkWarn("PasteBinMozillaOrg: parse failed\n" + data)
	}

	list := NewLinks()
	list.Add("link", PasteBinMozillaOrgURL+ID)
	list.Add("dl", PasteBinMozillaOrgURL+"/"+ID[5:])
	return list
}

//
//-------------------------------------------------------------[ CODEPAD.ORG ]--

const (
	CodePadOrgURL    = "http://codepad.org"
	CodePadOrgFormat = "Plain Text" // only used for text.
	CodePadOrgSubmit = "Submit"     // button.
)

func CodePadOrg(text, cLocalDir string, bAnonymous bool, limitRate int) Links {
	values := make(url.Values)
	values.Set("lang", CodePadOrgFormat)
	values.Set("submit", CodePadOrgSubmit)

	values.Set("code", text)

	data, e := postSimple(CodePadOrgURL, values)
	if e != nil {
		return linkErr(e, "CodePadOrg")
	}

	ID := findLink(string(data), "http://codepad.org/", "\"")
	if ID == "" {
		return linkWarn("CodePadOrg: parse failed") // \n" + data
	}

	return Links{
		"link": ID,
		"dl":   ID + "/raw.txt",
		"fork": ID + "/fork",
	}
}

//
//----------------------------------------------------------[ IMAGE BACKENDS ]--

//
//-------------------------------------------------------------[ IMAGEBIN.CA ]--

const ImageBinCaUrl = "http://imagebin.ca/upload.php"

func ImageBinCa(filepath, cLocalDir string, bAnonymous bool, limitRate int) Links {
	opt := map[string]string{
		// "t":       "file",
		"private": "true",
	}
	curly := curler(ImageBinCaUrl, "file", filepath, opt)
	defer curly.Cleanup()

	var data string
	curly.Setopt(curl.OPT_WRITEFUNCTION, func(ptr []byte, unk interface{}) bool {
		data += string(ptr)
		return true
	})

	if e := curly.Perform(); e != nil {
		return linkErr(e, "ImageBinCa")
	}

	ID := findLink(data, "http://", "\n")
	if ID == "" {
		return linkWarn("ImageBinCa: parse failed")
	}

	return Links{
		"link": ID,
		"page": strings.Replace(ID, "http://ibin.co/", "http://imagebin.ca/v/", 1)}
}

//
//-----------------------------------------------------------[ IMAGESHACK.US ]--

const ImageShackUsUrl = "http://imageshack.us/upload_api.php"
const ImageShackUsKey = "ABDGHOQS7d32e206ee33ef8cefb208d55dd030a6"

func ImageShackUs(filepath, cLocalDir string, bAnonymous bool, limitRate int) Links {
	opts := []string{
		"key=" + ImageShackUsKey,
		"public=no",
		"xml=yes",
	}
	data, e := curlExec(ImageShackUsUrl, limitRate, "fileupload", filepath, opts)
	if e != nil {
		return linkErr(e, "ImageShackUs")
	}

	list := NewLinks()
	list.Add("link", findPrefix(data, "<image_link>", "</image_link>"))
	list.Add("thumb", findPrefix(data, "<thumb_link>", "</thumb_link>"))
	return list
}

//
//---------------------------------------------------------------[ IMGUR.COM ]--

const ImgurComUrl = "http://imgur.com/api/upload.xml"
const ImgurComKey = "b3625162d3418ac51a9ee805b1840452"

func ImgurCom(filepath, cLocalDir string, bAnonymous bool, limitRate int) Links {
	opts := []string{"key=" + ImgurComKey}
	data, e := curlExec(ImgurComUrl, limitRate, "image", filepath, opts)
	if e != nil {
		return linkErr(e, "ImgurCom")
	}

	list := NewLinks()
	list.Add("link", findPrefix(data, "<original_image>", "</original_image>"))
	list.Add("thumb", findPrefix(data, "<small_thumbnail>", "</small_thumbnail>"))
	list.Add("large", findPrefix(data, "<large_thumbnail>", "</large_thumbnail>"))
	list.Add("page", findPrefix(data, "<imgur_page>", "</imgur_page>"))
	list.Add("del", findPrefix(data, "<delete_page>", "</delete_page>"))

	return list
}

//
//-----------------------------------------------------[ PIX.TOILE-LIBRE.ORG ]--

const PixToileLibreOrgUrl = "http://pix.toile-libre.org/?action=upload"

func PixToileLibreOrg(filepath, cLocalDir string, bAnonymous bool, limitRate int) Links {
	opt := map[string]string{
		"submit":  "Envoyer",
		"private": "yes"}
	curly := curler(PixToileLibreOrgUrl, "img", filepath, opt)
	defer curly.Cleanup()

	curly.Setopt(curl.OPT_WRITEFUNCTION, writeNull) // We do nothing, just prevent console flood.

	url, e := curlEffectiveUrl(curly)
	if e != nil {
		return linkErr(e, "PixToileLibreOrg")
	}

	return Links{
		"link":  strings.Replace(url, "?img=", "upload/original/", 1),
		"thumb": strings.Replace(url, "?img=", "upload/thumb/", 1)}
}

//
//---------------------------------------------------------------[ UPPIX.COM ]--

const UppixComUrl = "http://uppix.com/upload"

func UppixCom(filepath, cLocalDir string, bAnonymous bool, limitRate int) Links {

	// opt := map[string]string{"api": "1", "u_submit": "Upload", "u_agb": "yes"}
	// curly := curler(UppixUrl, "u_file", filepath, opt)
	// defer curly.Cleanup()

	// // disable HTTP/1.1 Expect: 100-continue
	// curly.Setopt(curl.OPT_HTTPHEADER, []string{"Expect:"})

	// if err := curly.Perform(); err != nil {
	// 	println("ERROR: ", err.Error())
	// }

	// a, e := curly.Getinfo(curl.INFO_EFFECTIVE_URL)
	// log.Info("out", a, e)

	// return

	opts := []string{"u_submit=Upload", "u_agb=yes"}
	data, e := curlExec(UppixComUrl, limitRate, "u_file", filepath, opts)
	if e != nil {
		return linkErr(e, "UppixCom")
	}

	list := NewLinks()
	list.Add("link", findLink(data, "http://uppix.com/f-", "&quot;"))
	list.Add("thumb", findLink(data, "http://uppix.com/t-", "&quot;"))
	return list
}

//
//----------------------------------------------------------[ VIDEO BACKENDS ]--

//
//------------------------------------------------------------[ VIDEOBIN.ORG ]--

const VideoBinOrgUrl = "http://videobin.org/add"

func VideoBinOrg(filepath, cLocalDir string, bAnonymous bool, limitRate int) Links {
	opt := map[string]string{"api": "1"}
	curly := curler(VideoBinOrgUrl, "videoFile", filepath, opt)
	defer curly.Cleanup()

	url, e := curlEffectiveUrl(curly)
	if e != nil {
		return linkErr(e, "VideoBinOrg")
	}

	return NewLinks().Add("link", url)

	// list := NewLinks()

	// if link, e := curly.Getinfo(curl.INFO_EFFECTIVE_URL); !log.Err(e, "get URL") {
	// 	list.Add("link", link.(string))
	// }

	// // log.Info("out", list)

	// return list
}

//
//-----------------------------------------------------------[ FILE BACKENDS ]--

//
//-----------------------------------------------------------------[ FREE.FR ]--

const FreeFrUrl = "ftp://dl.free.fr/"

// Use curl command for now as it allow the CombinedOutput to get the result from error pipe.
//
func DlFreeFr(filepath, cLocalDir string, bAnonymous bool, limitRate int) Links {
	body, e := exec.Command("curl", "-q", "-v", "-T", filepath, "-u", "cairo@dock.org:toto", FreeFrUrl).CombinedOutput()
	if e != nil || len(body) == 0 {
		return linkErr(e, "FreeFr")
	}

	list := NewLinks()
	list.Add("link", findLink(string(body), "http://dl.free.fr/", "\r\n"))
	list.Add("del", findLink(string(body), "http://dl.free.fr/rm.pl?", "\r\n"))
	return list
}

//
//---------------------------------------------------------[ UNUSED BACKENDS ]--

//
const FreeFrBlock = 10000

// Not working, but almost.
// It upload the file but can't get the answer as it's only forwarded with the verbose flood.
//
func FreeFrCurl(filepath, cLocalDir string, bAnonymous bool, limitRate int) Links {
	data, e := ioutil.ReadFile(filepath)
	if e != nil || len(data) == 0 {
		return linkErr(e, "FreeFr read file")
	}

	curly := curl.EasyInit()
	defer curly.Cleanup()

	curly.Setopt(curl.OPT_URL, FreeFrUrl+path.Base(filepath))

	curly.Setopt(curl.OPT_VERBOSE, true)

	curly.Setopt(curl.OPT_USERPWD, "cairo@dock.org:toto")

	curly.Setopt(curl.OPT_TRANSFERTEXT, false)
	curly.Setopt(curl.OPT_FTP_USE_EPSV, false)
	curly.Setopt(curl.OPT_UPLOAD, true)

	var current int
	curly.Setopt(curl.OPT_READFUNCTION, func(ptr []byte, _ interface{}) int {
		log.Info("read")
		size := ternary.Min(FreeFrBlock, len(data)-current)
		sent := copy(ptr, data[current:current+size]) // WARNING: never use append()
		current += sent
		return sent
	})

	// curly.Setopt(curl.OPT_WRITEFUNCTION, func(ptr []byte, unk interface{}) bool {
	// 	log.Info("", string(ptr), unk)
	// 	return true
	// })

	if err := curly.Perform(); err != nil {
		println("ERROR: ", err.Error())
	}

	a, e := curly.Getinfo(curl.INFO_FTP_ENTRY_PATH)
	log.Info("out", a, e)
	return nil
}
