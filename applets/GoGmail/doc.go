// Cairo-Dock Applet: GoGmail
//

// GoGmail is simple mail applet for the Cairo-Dock project.
//
// Problems:
// * Playing sound: wether using play, aplay or paplay, I got strange results. The first
// call can't be heard, and others only if called shortly after, but at low level.
// If I call it twice in a row really fast, the 2nd if played with full volume. This
// behaviour is the same if launched from a console, and I seem to use very different
// applications for all these calls:
//
// 	-rwxr-xr-x 1 root root 60688 avril  4  2012 /usr/bin/aplay
// 	lrwxrwxrwx 1 root root     5 juin   1 01:42 /usr/bin/paplay -> pacat
// 	lrwxrwxrwx 1 root root     3 déc.  27  2011 /usr/bin/play -> sox
//
// TODO:
// 	* split display to new config page with mail themes selection.
//	* use template for mails dialog.
//
package documentation
