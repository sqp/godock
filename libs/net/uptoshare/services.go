package uptoshare

import (
	"net/url"
	"strings"
)

//
//-----------------------------------------------------------[ TEXT BACKENDS ]--

//
//------------------------------------------------------------[ PASTEBIN.COM ]--

// PasteBinCom settings.
const (
	// URLPasteBinCom is the POST url for that service.
	URLPasteBinCom    = "http://pastebin.com/api/api_post.php"
	PasteBinComFormat = "text"                             // pastebin is only used for text.
	PasteBinComExpire = "1M"                               // 1 month
	PasteBinComKey    = "4dacb211338b25bfad20bc6d4358e555" // if you re-use this code, please make your own key !
	PasteBinComOption = "paste"
)

// PasteBinCom uploads text data to this website.
//
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

	link, e := postSimple(URLPasteBinCom, values)
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

// PasteUbuntuCom settings.
const (
	URLPasteUbuntuCom    = "http://paste.ubuntu.com"
	PasteUbuntuComFormat = "text"   // only used for text.
	PasteUbuntuComSubmit = "Paste!" // button.
)

// PasteUbuntuCom uploads text data to this website.
//
func PasteUbuntuCom(text, cLocalDir string, bAnonymous bool, limitRate int) Links {
	values := make(url.Values)
	values.Set("syntax", PasteUbuntuComFormat)
	values.Set("submit", PasteUbuntuComSubmit)
	values.Set("poster", "Anonymous") // bAnonymous ? "" : g_getenv ("USER"),

	values.Set("content", text)

	data, e := postSimple(URLPasteUbuntuCom, values)
	if e != nil {
		return linkErr(e, "PasteUbuntuCom")
	}

	ID := findPrefix(data, `class="pturl" href="/`, "/plain/")
	if ID == "" {
		return linkWarn("PasteUbuntuCom: parse failed\n" + data)
	}

	return NewLinks().Add("link", URLPasteUbuntuCom+"/"+ID)
}

//
//----------------------------------------------------[ PASTEBIN.MOZILLA.ORG ]--

// PasteBinMozillaOrg settings.
const (
	URLPasteBinMozillaOrg    = "http://pastebin.mozilla.org"
	PasteBinMozillaOrgFormat = "text" // only used for text.
	PasteBinMozillaOrgExpiry = "d"    // a day?
	PasteBinMozillaOrgSubmit = "Send" // button.
)

// PasteBinMozillaOrg uploads text data to this website.
//
func PasteBinMozillaOrg(text, cLocalDir string, bAnonymous bool, limitRate int) Links {
	values := make(url.Values)
	values.Set("format", PasteBinMozillaOrgFormat)
	values.Set("paste", PasteBinMozillaOrgSubmit)
	values.Set("expiry", PasteBinMozillaOrgExpiry)
	values.Set("remember", "0")
	values.Set("parent_pid", "")
	values.Set("poster", "Anonymous") // bAnonymous ? "" : g_getenv ("USER"),

	values.Set("code2", text)

	data, e := postSimple(URLPasteBinMozillaOrg, values)
	if e != nil {
		return linkErr(e, "PasteBinMozillaOrg")
	}

	ID := findLink(data, "/?dl=", `"`)
	if ID == "" {
		return linkWarn("PasteBinMozillaOrg: parse failed\n" + data)
	}

	list := NewLinks()
	list.Add("link", URLPasteBinMozillaOrg+ID)
	list.Add("dl", URLPasteBinMozillaOrg+"/"+ID[5:])
	return list
}

//
//-------------------------------------------------------------[ CODEPAD.ORG ]--

// CodePadOrg settings
const (
	URLCodePadOrg    = "http://codepad.org"
	CodePadOrgFormat = "Plain Text" // only used for text.
	CodePadOrgSubmit = "Submit"     // button.
)

// CodePadOrg uploads text data to this website.
//
func CodePadOrg(text, cLocalDir string, bAnonymous bool, limitRate int) Links {
	values := make(url.Values)
	values.Set("lang", CodePadOrgFormat)
	values.Set("submit", CodePadOrgSubmit)

	values.Set("code", text)

	data, e := postSimple(URLCodePadOrg, values)
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

// ImageBinCa settings.
const URLImageBinCa = "http://imagebin.ca/upload.php"

// ImageBinCa uploads an image file to this website.
//
func ImageBinCa(filepath, cLocalDir string, bAnonymous bool, limitRate int) Links {
	opts := []string{"private=true"}
	data, e := curlExec(URLImageBinCa, limitRate, "file", filepath, opts)
	if e != nil {
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

// ImageShackUs settings.
const (
	URLImageShackUs = "http://imageshack.us/upload_api.php"
	ImageShackUsKey = "ABDGHOQS7d32e206ee33ef8cefb208d55dd030a6"
)

// ImageShackUs uploads an image file to this website.
//
func ImageShackUs(filepath, cLocalDir string, bAnonymous bool, limitRate int) Links {
	opts := []string{
		"key=" + ImageShackUsKey,
		"public=no",
		"xml=yes",
	}
	data, e := curlExec(URLImageShackUs, limitRate, "fileupload", filepath, opts)
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

// ImgurComKey settings.
const (
	URLImgurCom = "http://imgur.com/api/upload.xml"
	ImgurComKey = "b3625162d3418ac51a9ee805b1840452"
)

// ImgurCom uploads an image file to this website.
//
func ImgurCom(filepath, cLocalDir string, bAnonymous bool, limitRate int) Links {
	opts := []string{"key=" + ImgurComKey}
	data, e := curlExec(URLImgurCom, limitRate, "image", filepath, opts)
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

// PixToileLibreOrg settings.
const URLPixToileLibreOrg = "http://pix.toile-libre.org/?action=upload"

// PixToileLibreOrg uploads an image file to this website.
//
func PixToileLibreOrg(filepath, cLocalDir string, bAnonymous bool, limitRate int) Links {
	data, e := curlExecArgs("-L",
		"-F", "MAX_FILE_SIZE=15360000",
		"-F", "img=@"+filepath+";type=image/png;filename="+filepath,
		URLPixToileLibreOrg)

	if e != nil {
		return linkErr(e, "PixToileLibreOrg")
	}

	return Links{
		"link":   findLink(data, "http://pix.toile-libre.org/upload/original", "</textarea>"),
		"thumb":  findLink(data, "http://pix.toile-libre.org/upload/thumb", "[/img]"),
		"medium": findLink(data, "http://pix.toile-libre.org/upload/img", "[/img]")}
}

//
//-----------------------------------------------------------[ POSTIMAGE.ORG ]--

// PostimageOrg settings.
const URLPostimageOrg = "http://postimage.org/"

// PostimageOrg uploads an image file to this website.
//
func PostimageOrg(filepath, cLocalDir string, bAnonymous bool, limitRate int) Links {
	opts := []string{"adult=no"}
	data, e := curlExec(URLPostimageOrg, limitRate, "upload[]", filepath, opts)
	if e != nil {
		return linkErr(e, "PostimageOrg")
	}

	list := NewLinks()
	list.Add("link", findLink(data, "http://postimg.org/image/", "' "))
	list.Add("thumb", findPrefix(data, "[img]", "[/img]"))
	list.Add("del", findLink(data, "http://postimg.org/delete/", "<"))
	return list
}

//
//---------------------------------------------------------------[ UPPIX.COM ]--

// UppixCom settings.
const URLUppixCom = "http://uppix.com/upload"

// UppixCom uploads an image file to this website.
//
func UppixCom(filepath, cLocalDir string, bAnonymous bool, limitRate int) Links {
	opts := []string{"u_submit=Upload", "u_agb=yes"}
	data, e := curlExec(URLUppixCom, limitRate, "u_file", filepath, opts)
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

// VideoBinOrg settings.
const URLVideoBinOrg = "http://videobin.org/add"

// VideoBinOrg uploads a video file to this website.
//
func VideoBinOrg(filepath, cLocalDir string, bAnonymous bool, limitRate int) Links {
	opts := []string{"api=1"}
	data, e := curlExec(URLVideoBinOrg, limitRate, "videoFile", filepath, opts)
	if e != nil {
		return linkErr(e, "VideoBinOrg")
	}

	return NewLinks().Add("link", data)
}

//
//-----------------------------------------------------------[ FILE BACKENDS ]--

//
//-----------------------------------------------------------------[ FREE.FR ]--

// DlFreeFr settings.
const URLDlFreeFr = "ftp://dl.free.fr/"

// DlFreeFr uploads any file to this website.
//
func DlFreeFr(filepath, cLocalDir string, bAnonymous bool, limitRate int) Links {
	body, e := curlExecArgs("-q", "-v", "-T", filepath, "-u", "cairo@dock.org:toto", URLDlFreeFr)
	if e != nil {
		return linkErr(e, "FreeFr")
	}

	list := NewLinks()
	list.Add("link", findLink(body, "http://dl.free.fr/", "\r\n"))
	list.Add("del", findLink(body, "http://dl.free.fr/rm.pl?", "\r\n"))
	return list
}