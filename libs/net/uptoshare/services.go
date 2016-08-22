package uptoshare

import (
	"github.com/sqp/godock/libs/net/upload"

	"encoding/json"
	"path/filepath"
	"strings"
)

var (
	backendImage = map[string]Sender{
		"Imagebam.com":  ImagebamCom,
		"Imagebin.ca":   ImageBinCa,
		"ImageShack.us": ImageShackUs,
		"Imgclick.net":  ImgclickNet,
		"Imgland.net":   ImglandNet,
		"Imgur.com":     ImgurCom, // Broken for me. Left for tests.
		// "pix.Toile-Libre.org": PixToileLibreOrg, // Need curl. Disabled unless someone really asks for it.
		"Postimage.org": PostimageOrg,
	}

	backendText = map[string]Sender{
		"Codepad.org":          CodePadOrg,
		"Pastebin.com":         PasteBinCom,
		"Pastebin.mozilla.org": PasteBinMozillaOrg,
		"Paste-ubuntu.com":     PasteUbuntuCom,
		"Play.golang.org":      PlayGolangOrg,
	}

	backendVideo = map[string]Sender{
		"VideoBin.org": VideoBinOrg,
	}

	backendFile = map[string]Sender{
		// "dl.free.fr":  DlFreeFr, // Need curl. Disabled unless someone really asks for it.
		"Filebin.ca":      FileBinCa,
		"Freemov.top":     FreemovTop,
		"Leopard.hosting": LeopardHosting,
		"Pixeldra.in":     PixeldraIn,
		"Transfer.sh":     TransferSh,
	}
)

//
//===========================================================[ TEXT BACKENDS ]==

//
//------------------------------------------------------------[ PASTEBIN.COM ]--

// PasteBinCom implements a Text Sender to the pastebin.com website.
//
var PasteBinCom = &Host{
	Poster: &upload.AsPostFormHTTP{
		UpBase:     upload.NewBaseURL("http://pastebin.com/api/api_post.php"),
		ContentRef: "api_paste_code",
		Options: map[string]string{
			"api_option":            "paste",
			"api_paste_private":     "1",                                // bAnonymous ? "1" : "0", // unlisted or public
			"api_paste_expire_date": "1M",                               // 1 month
			"api_paste_format":      "text",                             // pastebin is only used for text.
			"api_dev_key":           "4dacb211338b25bfad20bc6d4358e555", // if you re-use this code, please make your own key !
		},
		PrePost: func(hp *upload.AsPostFormHTTP, content *string) error {
			hp.Options["api_user_key"] = ""
			// "api_paste_name": , // bAnonymous ? "" : g_getenv ("USER"),
			return nil
		},
	},
	Parse: func(h *Host, data string) Links {
		if !strings.HasPrefix(data, "http://") {
			return linkWarn(h.Name() + ": bad format: " + data)
		}
		return NewLinks(data)
	},
}

//
//--------------------------------------------------------[ PASTE.UBUNTU.COM ]--

// PasteUbuntuCom implements a Text Sender to the paste.ubuntu.com website.
//
var PasteUbuntuCom = &Host{
	Poster: &upload.AsPostFormHTTP{
		UpBase:     upload.NewBaseURL("http://paste.ubuntu.com"),
		ContentRef: "content",
		Options: map[string]string{
			"syntax": "text",
			"submit": "Paste!",
		},
		PrePost: func(hp *upload.AsPostFormHTTP, content *string) error {
			hp.Options["poster"] = "Anonymous" // bAnonymous ? "" : g_getenv ("USER"),
			return nil
		},
	},
	Parse: func(h *Host, data string) Links {
		ID := findPrefix(data, `class="pturl" href="/`, "/plain/")
		if ID == "" {
			return linkWarn(h.Name() + ": parse failed\n" + data)
		}
		return NewLinks(h.GetURL("") + "/" + ID)
	},
}

//
//----------------------------------------------------[ PASTEBIN.MOZILLA.ORG ]--

// PasteBinMozillaOrg implements a Text Sender to the pastebin.mozilla.org website.
//
var PasteBinMozillaOrg = &Host{
	Poster: &upload.AsPostFormHTTP{
		UpBase:     upload.NewBaseURL("https://pastebin.mozilla.org"),
		ContentRef: "code2",
		Options: map[string]string{
			"format":     "text",
			"paste":      "Send",
			"expiry":     "d",
			"remember":   "0",
			"parent_pid": "",
		},
		PrePost: func(hp *upload.AsPostFormHTTP, content *string) error {
			hp.Options["poster"] = "Anonymous" // bAnonymous ? "" : g_getenv ("USER"),
			return nil
		},
	},
	Parse: func(h *Host, data string) Links {
		ID := findLink(data, "/?dl=", `"`)
		if ID == "" {
			return linkWarn(h.Name() + ": parse failed\n" + data)
		}
		uri := h.GetURL("")
		return NewLinks(uri+ID).
			Add("dl", uri+"/"+ID[5:])
	},
}

//
//-------------------------------------------------------------[ CODEPAD.ORG ]--

// CodePadOrg implements a Text Sender to the codepad.org website.
//
var CodePadOrg = &Host{
	Poster: upload.NewPostForm("http://codepad.org/",
		"code",
		"lang", "Plain Text",
		"submit", "Submit",
	),
	Parse: func(h *Host, data string) Links {
		uri := h.GetURL("")
		ID := findLink(string(data), uri, "\"")
		if ID == "" {
			return linkWarn(h.Name() + ": parse failed") // \n" + data
		}
		return NewLinks(ID).
			Add("dl", ID+"/raw.txt").
			Add("fork", ID+"/fork")
	},
}

//
//---------------------------------------------------------[ PLAY.GOLANG.ORG ]--

// PlayGolangOrg implements a Text Sender to the play.golang.org website.
//
var PlayGolangOrg = &Host{
	Poster: upload.NewRequester("POST", "https://play.golang.org/share"),
	Parse: func(h *Host, data string) Links {
		return NewLinks("https://play.golang.org/p/" + data)
	},
}

//
//==========================================================[ IMAGE BACKENDS ]==

//
//------------------------------------------------------------[ IMAGEBAM.COM ]--

// ImagebamCom implements an Image Sender to the imagebam.com website.
//
var ImagebamCom = &Host{
	Poster: upload.NewMultiparter("http://www.imagebam.com/sys/upload/save",
		"file[]",
		"content_type", "1", // 1 == NSFW            Options: 0, 1.
		"thumb_size", "250", // 250x250.             Options: 150, 180, 250, 300, 350.
		"thumb_aspect_ratio", "resize", // resize= keep ratio.  Options: resize, crop.
		"thumb_file_type", "jpg", //                      Options: jpg, gif.
	),
	Parse: func(h *Host, data string) Links {
		return NewLinks(findPrefix(data, "[URL=", "][IMG]")).
			Add("del", findLink(data, "http://www.imagebam.com/remove/", "'></div>"))
	},
}

//
//-------------------------------------------------------------[ IMAGEBIN.CA ]--

// ImageBinCa implements an Image Sender to the imagebin.ca website.
//
var ImageBinCa = &Host{
	Poster: upload.NewMultiparter("https://imagebin.ca/upload.php",
		"file",
		"private", "true",
	),
	Parse: func(h *Host, data string) Links {
		ID := findPrefix(data, "url:", "\n")
		if ID == "" {
			return linkWarn(h.Name() + ": parse failed")
		}
		return NewLinks(ID).
			Add("page", strings.Replace(ID, "http://ibin.co/", "http://imagebin.ca/v/", 1))
	},
}

//
//-----------------------------------------------------------[ IMAGESHACK.US ]--

// ImageShackUs implements an Image Sender to the imageshack.us website.
//
var ImageShackUs = &Host{
	Poster: upload.NewMultiparter("http://imageshack.us/upload_api.php",
		"fileupload",
		"key", "ABDGHOQS7d32e206ee33ef8cefb208d55dd030a6",
		"public", "no",
		"xml", "yes",
	),
	Parse: func(h *Host, data string) Links {
		return NewLinks(findPrefix(data, "<image_link>", "</image_link>")).
			Add("thumb", findPrefix(data, "<thumb_link>", "</thumb_link>"))
	},
}

//
//------------------------------------------------------------[ FREEIMAGEHOSTING.NET ]--

// ImglandNet implements an Image Sender to the imagebam.com website.
//
var ImglandNet = &Host{
	Poster: upload.NewMultiparter("https://imgland.net/process.php?subAPI=mainsite",
		"imagefile[]",
		"usubmit", "true",
	),
	Parse: func(h *Host, data string) Links {
		ret := &imglandNetJSON{}
		e := json.Unmarshal([]byte(data), ret)
		if e != nil {
			return linkErr(e, h.Name())
		}
		return NewLinks(ret.URL)
	},
}

type imglandNetJSON struct {
	URL string `json:"url"`
}

//
//------------------------------------------------------------[ IMGCLICK.NET ]--

// ImgclickNet implements an Image Sender to the imgclick.net website.
//
var ImgclickNet = &Host{
	Poster: upload.NewMultiparter("http://main.imgclick.net/cgi-bin/upload_file.cgi?upload_id=",
		"file_0",
		"upload_type", "file",
	),
	Parse: func(h *Host, data string) Links {
		return NewLinks(findLink(data, "http://main.imgclick.net/i/", "[/IMG][/URL]"))
	},
}

//
//---------------------------------------------------------------[ IMGUR.COM ]--

// ImgurCom implements an Image Sender to the imgur.com website.
//
var ImgurCom = &Host{
	Poster: upload.NewMultiparter("http://imgur.com/api/upload.xml",
		"image",
		"key", "b3625162d3418ac51a9ee805b1840452",
	),
	Parse: func(h *Host, data string) Links {
		println(data)
		return NewLinks(findPrefix(data, "<original_image>", "</original_image>")).
			Add("thumb", findPrefix(data, "<small_thumbnail>", "</small_thumbnail>")).
			Add("large", findPrefix(data, "<large_thumbnail>", "</large_thumbnail>")).
			Add("page", findPrefix(data, "<imgur_page>", "</imgur_page>")).
			Add("del", findPrefix(data, "<delete_page>", "</delete_page>"))
	},
}

//
//-----------------------------------------------------------[ POSTIMAGE.ORG ]--

// PostimageOrg implements an Image Sender to the postimage.org website.
//
var PostimageOrg = &Host{
	Poster: upload.NewMultiparter("https://old.postimage.org/",
		"upload[]",
		"adult", "no",
	),
	Parse: func(h *Host, data string) Links {
		return NewLinks(findLink(data, "http://postimg.org/image/", "' ")).
			Add("thumb", findPrefix(data, "[img]", "[/img]")).
			Add("del", findLink(data, "http://postimg.org/delete/", "<"))
	},
}

//
//==========================================================[ VIDEO BACKENDS ]==

//
//------------------------------------------------------------[ VIDEOBIN.ORG ]--

// VideoBinOrg implements a Video Sender to the postimage.org website.
//
var VideoBinOrg = &Host{
	Poster: upload.NewMultiparter("https://videobin.org/add",
		"videoFile",
		"api", "1",
	),
	Parse: func(h *Host, data string) Links {
		return NewLinks(data)
	},
}

//
//===========================================================[ FILE BACKENDS ]==

//
//--------------------------------------------------------------[ FILEBIN.CA ]--

// FileBinCa implements a File Sender to the imagebin.ca website.
//
var FileBinCa = &Host{
	Poster: upload.NewMultiparter("http://filebin.ca/upload.php", "file"),
	Parse: func(h *Host, data string) Links {
		ID := findPrefix(data, "url:", "\n")
		if ID == "" {
			return linkWarn(h.Name() + ": parse failed\n" + data)
		}
		return NewLinks(ID).
			Add("page", strings.Replace(ID, "http://ibin.co/", "http://filebin.ca/v/", 1))
	},
}

//
//-------------------------------------------------------------[ FREEMOV.TOP ]--

// FreemovTop implements a File Sender to the freemov.top website.
//
var FreemovTop = &Host{
	Poster: upload.NewMultiparter("http://freemov.top/",
		"upload[]",
		"submit", "submit",
	),
	Parse: func(h *Host, data string) Links {
		link := findPrefix(data, `id="name" value=`, ` />`)
		if link == "" {
			println(h.Name()+" -- ANSWER\n", data)
			return linkWarn(h.Name() + ": bad format")
		}
		return NewLinks("http://" + link)
	},
}

//
//---------------------------------------------------------[ LEOPARD.HOSTING ]--

// LeopardHosting implements a File Sender to the leopard.hosting website.
//
var LeopardHosting = &Host{
	Poster: upload.NewMultiparter("http://leopard.hosting/upload.php",
		"uploadContent",
		// "password", "",
		"public", "no",
		"showname", "no",
		"json", "true",
	),
	Parse: func(h *Host, data string) Links {
		ret := &leopardHostingJSON{}
		e := json.Unmarshal([]byte(data), ret)
		if e != nil {
			return linkErr(e, h.Name())
		}
		return NewLinks(ret.Upload.DownloadURL).
			Add("del", ret.Upload.DeleteURL).
			Add("support", ret.Upload.Support).
			Add("ID", ret.Upload.FileCode)
	},
}

type leopardHostingJSON struct {
	Upload struct {
		Support     string `json:"support"`
		DownloadURL string `json:"downloadURL"`
		FileCode    string `json:"fileCode"`
		// DeleteKey   string `json:"deleteKey"`
		DeleteURL string `json:"deleteURL"`
	} `json:"upload"`
}

//
//-------------------------------------------------------------[ PIXELDRA.IN ]--

// PixeldraIn implements a File Sender to the pixeldra.in website.
//
var PixeldraIn = &Host{
	Poster: &upload.AsMultipart{
		UpBase:  upload.NewBaseURL("http://pixeldra.in/api/upload"),
		FileRef: "file",
		PrePost: func(m *upload.AsMultipart, file string) error {
			m.Options = map[string]string{
				"fileName": filepath.Base(file),
			}
			return nil
		},
	},
	Parse: func(h *Host, data string) Links {
		ret := pixeldraInJSON{}
		e := json.Unmarshal([]byte(data), &ret)
		if e != nil {
			return linkErr(e, h.Name())
		}
		return NewLinks(ret.URL).
			Add("del", ret.ID)
	},
}

type pixeldraInJSON struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

//
//-------------------------------------------------------------[ TRANSFER.SH ]--

// TransferSh implements a File Sender to the transfer.sh website.
//
var TransferSh = &Host{
	Poster: &upload.AsRequestHTTP{
		Method: "PUT",
		UpBase: upload.NewBaseCB(func(file string) string {
			return "https://transfer.sh/" + filepath.Base(file)
		}),
	},
	Parse: func(h *Host, data string) Links {
		data = strings.Trim(data, " \n")
		if data == "Not Found" {
			return linkWarn(h.Name() + " has answered: " + data)
		}
		return NewLinks(data)
	},
}
